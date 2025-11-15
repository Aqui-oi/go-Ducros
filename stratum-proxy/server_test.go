package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestServerLoginAndSubmit(t *testing.T) {
	work := [4]string{
		"0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"0x00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"0x1",
	}

	var (
		submitMu       sync.Mutex
		recordedNonce  string
		recordedHeader string
		recordedMix    string
	)

	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}

		resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
		switch req.Method {
		case "eth_blockNumber":
			resp.Result = mustRaw("0x1")
		case "randomx_getWork":
			resp.Result = mustRaw(work)
		case "randomx_submitWork":
			if len(req.Params) != 3 {
				t.Fatalf("unexpected params: %v", req.Params)
			}
			nonce, ok1 := req.Params[0].(string)
			header, ok2 := req.Params[1].(string)
			mix, ok3 := req.Params[2].(string)
			if !ok1 || !ok2 || !ok3 {
				t.Fatalf("params type assertion failed: %v", req.Params)
			}
			submitMu.Lock()
			recordedNonce = nonce
			recordedHeader = header
			recordedMix = mix
			submitMu.Unlock()
			resp.Result = mustRaw(true)
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode rpc response: %v", err)
		}
	}))
	defer rpc.Close()

	cfg := &ServerConfig{
		ListenAddr:  "127.0.0.1:0",
		GethRPC:     rpc.URL,
		InitialDiff: 1,
		Algorithm:   "rx/0",
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	// Populate initial work so login succeeds.
	srv.updateWork()

	conn, err := net.Dial("tcp", srv.listener.Addr().String())
	if err != nil {
		t.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	loginReq := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "login",
		"params": map[string]interface{}{
			"login": "0xabc",
			"pass":  "worker1",
			"agent": "xmrig/test",
		},
	}
	sendJSON(t, conn, loginReq)

	loginResp := readResponse(t, reader)
	resultMap, ok := loginResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected login result: %#v", loginResp.Result)
	}
	jobMap, ok := resultMap["job"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing job in login response: %#v", resultMap)
	}

	blob, ok := jobMap["blob"].(string)
	if !ok {
		t.Fatalf("blob missing or not string: %#v", jobMap["blob"])
	}
	if len(blob) < 72 {
		t.Fatalf("blob too short: %d", len(blob))
	}

	extraNonceHex := blob[64:72]
	extraNonceBytes, err := hex.DecodeString(extraNonceHex)
	if err != nil {
		t.Fatalf("decode extra nonce: %v", err)
	}
	extraNonce := binary.LittleEndian.Uint32(extraNonceBytes)

	nonceLE := "78563412"
	submitReq := map[string]interface{}{
		"id":      2,
		"jsonrpc": "2.0",
		"method":  "submit",
		"params": map[string]interface{}{
			"id":     resultMap["id"],
			"job_id": jobMap["job_id"],
			"nonce":  nonceLE,
			"result": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},
	}
	sendJSON(t, conn, submitReq)

	submitResp := readResponse(t, reader)
	if submitResp.Error != nil {
		t.Fatalf("submit returned error: %#v", submitResp.Error)
	}

	expectedNonce64 := (uint64(extraNonce) << 32) | 0x12345678

	submitMu.Lock()
	gotNonce := recordedNonce
	gotHeader := recordedHeader
	gotMix := recordedMix
	submitMu.Unlock()

	if gotNonce != formatNonce(expectedNonce64) {
		t.Fatalf("unexpected nonce: got %s want %s", gotNonce, formatNonce(expectedNonce64))
	}
	if gotHeader != work[0] {
		t.Fatalf("unexpected header: got %s want %s", gotHeader, work[0])
	}
	expectedMix := "0x" + submitReq["params"].(map[string]interface{})["result"].(string)
	if gotMix != expectedMix {
		t.Fatalf("unexpected mixdigest: got %s want %s", gotMix, expectedMix)
	}
}

func TestServerConcurrentMiners(t *testing.T) {
	work := [4]string{
		"0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"0x00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"0x1",
	}

	type submission struct {
		nonce  string
		header string
		mix    string
	}

	var (
		submitMu    sync.Mutex
		submissions []submission
	)

	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode rpc request: %v", err)
		}

		resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
		switch req.Method {
		case "eth_blockNumber":
			resp.Result = mustRaw("0x1")
		case "randomx_getWork":
			resp.Result = mustRaw(work)
		case "randomx_submitWork":
			if len(req.Params) != 3 {
				t.Fatalf("unexpected params: %v", req.Params)
			}
			nonce, ok1 := req.Params[0].(string)
			header, ok2 := req.Params[1].(string)
			mix, ok3 := req.Params[2].(string)
			if !ok1 || !ok2 || !ok3 {
				t.Fatalf("params type assertion failed: %v", req.Params)
			}
			submitMu.Lock()
			submissions = append(submissions, submission{nonce: nonce, header: header, mix: mix})
			submitMu.Unlock()
			resp.Result = mustRaw(true)
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode rpc response: %v", err)
		}
	}))
	defer rpc.Close()

	cfg := &ServerConfig{
		ListenAddr:  "127.0.0.1:0",
		GethRPC:     rpc.URL,
		InitialDiff: 1,
		Algorithm:   "rx/0",
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer srv.Stop()

	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	srv.updateWork()

	miners := 4

	start := make(chan struct{})
	var wg sync.WaitGroup

	var expectedMu sync.Mutex
	expected := make(map[string]string)

	for i := 0; i < miners; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", srv.listener.Addr().String())
			if err != nil {
				t.Errorf("dial server: %v", err)
				return
			}
			defer conn.Close()

			reader := bufio.NewReader(conn)

			<-start

			loginReq := map[string]interface{}{
				"id":      idx + 1,
				"jsonrpc": "2.0",
				"method":  "login",
				"params": map[string]interface{}{
					"login": fmt.Sprintf("0xminer%02d", idx),
					"pass":  fmt.Sprintf("worker-%d", idx),
					"agent": "xmrig/test",
				},
			}
			sendJSON(t, conn, loginReq)

			loginResp := readResponse(t, reader)
			resultMap, ok := loginResp.Result.(map[string]interface{})
			if !ok {
				t.Errorf("unexpected login result: %#v", loginResp.Result)
				return
			}
			jobMap, ok := resultMap["job"].(map[string]interface{})
			if !ok {
				t.Errorf("missing job in login response: %#v", resultMap)
				return
			}

			blob, ok := jobMap["blob"].(string)
			if !ok {
				t.Errorf("blob missing or not string: %#v", jobMap["blob"])
				return
			}
			if len(blob) < 72 {
				t.Errorf("blob too short: %d", len(blob))
				return
			}

			extraNonceBytes, err := hex.DecodeString(blob[64:72])
			if err != nil {
				t.Errorf("decode extra nonce: %v", err)
				return
			}
			extraNonce := binary.LittleEndian.Uint32(extraNonceBytes)

			minerNonce := uint32(idx + 1)
			submitReq := map[string]interface{}{
				"id":      miners + idx + 1,
				"jsonrpc": "2.0",
				"method":  "submit",
				"params": map[string]interface{}{
					"id":     resultMap["id"],
					"job_id": jobMap["job_id"],
					"nonce":  littleEndianNonceHex(minerNonce),
					"result": fmt.Sprintf("%064x", idx+1),
				},
			}
			sendJSON(t, conn, submitReq)

			submitResp := readResponse(t, reader)
			if submitResp.Error != nil {
				t.Errorf("submit returned error: %#v", submitResp.Error)
				return
			}

			expectedNonce := formatNonce((uint64(extraNonce) << 32) | uint64(minerNonce))
			expectedMix := "0x" + submitReq["params"].(map[string]interface{})["result"].(string)

			expectedMu.Lock()
			expected[expectedNonce] = expectedMix
			expectedMu.Unlock()
		}(i)
	}

	close(start)
	wg.Wait()

	submitMu.Lock()
	recorded := append([]submission(nil), submissions...)
	submitMu.Unlock()

	if len(recorded) != miners {
		t.Fatalf("expected %d submissions, got %d", miners, len(recorded))
	}

	for _, sub := range recorded {
		if sub.header != work[0] {
			t.Fatalf("unexpected header: got %s want %s", sub.header, work[0])
		}
		expectedMu.Lock()
		mix, ok := expected[sub.nonce]
		expectedMu.Unlock()
		if !ok {
			t.Fatalf("unexpected nonce submitted: %s", sub.nonce)
		}
		if sub.mix != mix {
			t.Fatalf("unexpected mixdigest: got %s want %s", sub.mix, mix)
		}
	}
}

func mustRaw(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(b)
}

func sendJSON(t *testing.T, conn net.Conn, payload interface{}) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("write payload: %v", err)
	}
}

func readResponse(t *testing.T, reader *bufio.Reader) StratumResponse {
	t.Helper()
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	var resp StratumResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func formatNonce(nonce uint64) string {
	return fmt.Sprintf("0x%016x", nonce)
}

func littleEndianNonceHex(v uint32) string {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	return hex.EncodeToString(buf[:])
}
