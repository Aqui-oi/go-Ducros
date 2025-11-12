package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Server represents the Stratum proxy server
type Server struct {
	config      *ServerConfig
	rpcClient   *RPCClient
	listener    net.Listener
	miners      map[string]*Miner
	minersMu    sync.RWMutex
	currentWork *WorkPackage
	currentJob  *Job
	workMu      sync.RWMutex
	stats       *Stats
	jobCounter  uint64
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewServer creates a new Stratum server
func NewServer(config *ServerConfig) (*Server, error) {
	rpcClient := NewRPCClient(config.GethRPC)

	// Test connection
	if err := rpcClient.CheckConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Geth: %w", err)
	}

	return &Server{
		config:     config,
		rpcClient:  rpcClient,
		miners:     make(map[string]*Miner),
		stats:      NewStats(),
		stopCh:     make(chan struct{}),
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	// Start work updater
	s.wg.Add(1)
	go s.workUpdater()

	// Start stats reporter
	s.wg.Add(1)
	go s.statsReporter()

	// Accept connections
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

// acceptConnections accepts incoming miner connections
func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				log.Printf("❌ Accept error: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleMiner(conn)
	}
}

// handleMiner handles a single miner connection
func (s *Server) handleMiner(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	minerID := conn.RemoteAddr().String()
	log.Printf("🔌 New connection from %s", minerID)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Generate random 4-byte extraNonce for rx-eth-v1 format
	var extraNonceBytes [4]byte
	rand.Read(extraNonceBytes[:])
	extraNonce := binary.LittleEndian.Uint32(extraNonceBytes[:])

	miner := &Miner{
		ID:           minerID,
		Difficulty:   uint64(s.config.InitialDiff),
		ExtraNonce:   extraNonce,
		LastActivity: time.Now(),
	}

	// Register miner
	s.minersMu.Lock()
	s.miners[minerID] = miner
	s.minersMu.Unlock()

	defer func() {
		s.minersMu.Lock()
		delete(s.miners, minerID)
		s.minersMu.Unlock()
		log.Printf("👋 Miner %s disconnected", minerID)
	}()

	// Handle requests
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if s.config.Verbose {
				log.Printf("Read error from %s: %v", minerID, err)
			}
			return
		}

		// Parse request
		var req StratumRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("❌ Invalid JSON from %s: %v", minerID, err)
			continue
		}

		// Handle request
		response := s.handleRequest(miner, &req)

		// Send response
		responseJSON, _ := json.Marshal(response)
		responseJSON = append(responseJSON, '\n')

		// Debug: log the response we're sending
		if s.config.Verbose {
			log.Printf("📤 [%s] Response: %s", minerID, string(responseJSON))
		}

		if _, err := writer.Write(responseJSON); err != nil {
			log.Printf("Write error to %s: %v", minerID, err)
			return
		}
		writer.Flush()

		miner.LastActivity = time.Now()
	}
}

// handleRequest handles a Stratum request
func (s *Server) handleRequest(miner *Miner, req *StratumRequest) *StratumResponse {
	if s.config.Verbose {
		log.Printf("📩 [%s] %s %v", miner.ID, req.Method, req.Params)
	}

	switch req.Method {
	case "login":
		return s.handleLogin(miner, req)
	case "submit":
		return s.handleSubmit(miner, req)
	case "keepalived":
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Result:  map[string]interface{}{"status": "KEEPALIVED"},
		}
	default:
		log.Printf("⚠️  Unknown method from %s: %s", miner.ID, req.Method)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: fmt.Sprintf("Unknown method: %s", req.Method),
			},
		}
	}
}

// handleLogin handles miner login
func (s *Server) handleLogin(miner *Miner, req *StratumRequest) *StratumResponse {
	// Try params as object first (xmrig format)
	paramsObj, err := req.GetParamsObject()
	var loginData map[string]interface{}

	if err == nil {
		// Params is an object directly (xmrig style)
		loginData = paramsObj
	} else {
		// Try params as array (standard stratum style)
		paramsArray, err := req.GetParamsArray()
		if err != nil || len(paramsArray) < 1 {
			return &StratumResponse{
				ID:      req.ID,
				JSONRPC: "2.0",
				Error: &StratumError{
					Code:    -1,
					Message: "Missing login parameters",
				},
			}
		}

		// Get first element from array
		if obj, ok := paramsArray[0].(map[string]interface{}); ok {
			loginData = obj
		} else if login, ok := paramsArray[0].(string); ok {
			// Simple string login
			miner.Address = login
			miner.WorkerName = login
			loginData = nil
		}
	}

	// Parse login data
	if loginData != nil {
		if login, ok := loginData["login"].(string); ok {
			miner.Address = login
		}
		if pass, ok := loginData["pass"].(string); ok {
			miner.WorkerName = pass
		}
		if agent, ok := loginData["agent"].(string); ok {
			miner.Agent = agent
		}
	}

	log.Printf("✅ Miner %s logged in: %s (%s)", miner.ID, miner.Address, miner.Agent)

	// Get current job
	s.workMu.RLock()
	job := s.currentJob
	s.workMu.RUnlock()

	if job == nil {
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "No work available",
			},
		}
	}

	// Send job to miner
	miner.CurrentJob = job

	// Create rx-eth-v1 blob with miner's extraNonce
	// Format: headerHash(32) || extraNonce(4) || const3(3) || nonce4(4)
	blob := createBlobRxEth(job.HeaderHash, miner.ExtraNonce)

	// Create complete job response for xmrig RandomX compatibility
	// Use MINER's difficulty (not blockchain difficulty) for target
	jobResponse := JobResponse{
		JobID:     job.JobID,
		Algo:      "rx/0",
		SeedHash:  strings.TrimPrefix(job.SeedHash, "0x"), // Remove 0x prefix
		Height:    job.Height,
		Blob:      blob, // Use rx-eth-v1 format with miner's extraNonce
		Target:    DifficultyToStratumTarget(miner.Difficulty),
		CleanJobs: true,
	}

	result := map[string]interface{}{
		"id":         miner.ID,
		"job":        jobResponse,
		"status":     "OK",
		"extensions": []string{"keepalive", "algo"},
	}

	return &StratumResponse{
		ID:      req.ID,
		JSONRPC: "2.0",
		Result:  result,
		// No Error field for success (omitempty)
	}
}

// handleSubmit handles share submission
func (s *Server) handleSubmit(miner *Miner, req *StratumRequest) *StratumResponse {
	// Try params as object first (xmrig format)
	submitData, err := req.GetParamsObject()

	if err != nil {
		// Try params as array (standard stratum style)
		paramsArray, err := req.GetParamsArray()
		if err != nil || len(paramsArray) < 1 {
			return &StratumResponse{
				ID:      req.ID,
				JSONRPC: "2.0",
				Error: &StratumError{
					Code:    -1,
					Message: "Missing submit parameters",
				},
			}
		}

		// Get first element from array
		if obj, ok := paramsArray[0].(map[string]interface{}); ok {
			submitData = obj
		} else {
			return &StratumResponse{
				ID:      req.ID,
				JSONRPC: "2.0",
				Error: &StratumError{
					Code:    -1,
					Message: "Invalid submit format",
				},
			}
		}
	}

	jobID, _ := submitData["job_id"].(string)
	nonceStr, _ := submitData["nonce"].(string)
	resultStr, _ := submitData["result"].(string)

	if s.config.Verbose {
		log.Printf("📤 Share from %s: job=%s nonce=%s result=%s",
			miner.ID, jobID, nonceStr, resultStr)
	}

	// Validate job
	if miner.CurrentJob == nil || miner.CurrentJob.JobID != jobID {
		log.Printf("❌ Stale share from %s", miner.ID)
		miner.SharesInvalid++
		s.stats.RecordShare(false)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Stale share",
			},
		}
	}

	// Extract miner nonce4 (4 bytes LE) and combine with extraNonce
	// xmrig sends nonce as 8 hex chars (4 bytes LE)
	if len(nonceStr) != 8 {
		log.Printf("❌ Invalid nonce length from %s: %d (expected 8)", miner.ID, len(nonceStr))
		miner.SharesInvalid++
		s.stats.RecordShare(false)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Invalid nonce length",
			},
		}
	}

	// Parse miner's 4-byte nonce (little-endian)
	minerNonceBytes, err := hex.DecodeString(nonceStr)
	if err != nil {
		log.Printf("❌ Invalid nonce hex from %s: %v", miner.ID, err)
		miner.SharesInvalid++
		s.stats.RecordShare(false)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Invalid nonce hex",
			},
		}
	}

	minerNonce4 := binary.LittleEndian.Uint32(minerNonceBytes)

	// Combine extraNonce (high 32 bits) and minerNonce (low 32 bits)
	// nonce64 = (extraNonce << 32) | minerNonce4
	nonce64 := (uint64(miner.ExtraNonce) << 32) | uint64(minerNonce4)
	nonceHex := fmt.Sprintf("0x%016x", nonce64)

	if s.config.Verbose {
		log.Printf("🔢 Nonce: extraNonce=%08x minerNonce=%08x combined=%016x",
			miner.ExtraNonce, minerNonce4, nonce64)
	}

	// Ensure result hash has 0x prefix for Ethereum
	if !strings.HasPrefix(resultStr, "0x") {
		resultStr = "0x" + resultStr
	}

	// Submit to Geth
	accepted, err := s.rpcClient.SubmitWork(
		nonceHex,
		miner.CurrentJob.HeaderHash,
		resultStr, // Use result as mixDigest
	)

	if err != nil {
		log.Printf("❌ Submit error for %s: %v", miner.ID, err)
		miner.SharesInvalid++
		s.stats.RecordShare(false)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: fmt.Sprintf("Submit failed: %v", err),
			},
		}
	}

	if accepted {
		log.Printf("✅ Valid share from %s (diff: %d)", miner.ID, miner.Difficulty)
		miner.SharesValid++
		s.stats.RecordShare(true)

		// Check if it's a block
		// TODO: Implement proper block detection
		// For now, assume high difficulty shares might be blocks
		if miner.Difficulty > 1000000 {
			log.Printf("🎉 BLOCK FOUND by %s!", miner.ID)
			s.stats.RecordBlock()
		}

		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Result:  map[string]interface{}{"status": "OK"},
		}
	} else {
		log.Printf("❌ Invalid share from %s", miner.ID)
		miner.SharesInvalid++
		s.stats.RecordShare(false)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Invalid share",
			},
		}
	}
}

// workUpdater fetches new work from Geth periodically
func (s *Server) workUpdater() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.updateWork()
		}
	}
}

// updateWork fetches new work and distributes to miners
func (s *Server) updateWork() {
	work, err := s.rpcClient.GetWork()
	if err != nil {
		if s.config.Verbose {
			log.Printf("⚠️  Failed to get work: %v", err)
		}
		return
	}

	// Check if work changed
	s.workMu.RLock()
	currentHeader := ""
	if s.currentWork != nil {
		currentHeader = s.currentWork.HeaderHash
	}
	s.workMu.RUnlock()

	if work.HeaderHash == currentHeader {
		// No new work
		return
	}

	// Create new job
	s.jobCounter++
	jobID := fmt.Sprintf("%d", s.jobCounter)

	job, err := WorkToJob(work, jobID, s.config.Algorithm)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	// Update current work
	s.workMu.Lock()
	s.currentWork = work
	s.currentJob = job
	s.workMu.Unlock()

	log.Printf("📦 New job %s: block %d, seed %s",
		jobID, job.Height, job.SeedHash[:16]+"...")

	// Broadcast to all miners
	s.broadcastJob(job)
}

// broadcastJob sends new job to all connected miners
func (s *Server) broadcastJob(job *Job) {
	s.minersMu.RLock()
	defer s.minersMu.RUnlock()

	for _, miner := range s.miners {
		miner.CurrentJob = job
		// TODO: Send job notification to miner
		// This would require keeping connection writers
	}
}

// statsReporter logs statistics periodically
func (s *Server) statsReporter() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.printStats()
		}
	}
}

// printStats prints current statistics
func (s *Server) printStats() {
	s.minersMu.RLock()
	totalMiners := len(s.miners)
	activeMiners := 0
	var totalHashrate float64

	for _, miner := range s.miners {
		if time.Since(miner.LastActivity) < 2*time.Minute {
			activeMiners++
			totalHashrate += miner.Hashrate
		}
	}
	s.minersMu.RUnlock()

	total, _, shares, valid, invalid, blocks, _, uptime := s.stats.GetStats()
	s.stats.UpdateMiners(totalMiners, activeMiners)
	s.stats.UpdateHashrate(totalHashrate)

	log.Printf("📊 Stats: Miners=%d/%d Shares=%d/%d/%d Blocks=%d Hashrate=%.2f H/s Uptime=%s",
		activeMiners, total, valid, invalid, shares, blocks, totalHashrate, uptime.Round(time.Second))
}
