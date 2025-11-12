package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// RPCClient handles communication with Geth's JSON-RPC
type RPCClient struct {
	url        string
	httpClient *http.Client
}

// NewRPCClient creates a new RPC client
func NewRPCClient(url string) *RPCClient {
	return &RPCClient{
		url: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// call makes a JSON-RPC call
func (c *RPCClient) call(method string, params []interface{}) (json.RawMessage, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var response JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// GetWork fetches work from Geth (eth_getWork or randomx_getWork)
func (c *RPCClient) GetWork() (*WorkPackage, error) {
	// Try randomx_getWork first, fallback to eth_getWork
	result, err := c.call("randomx_getWork", []interface{}{})
	if err != nil {
		// Fallback to eth_getWork
		result, err = c.call("eth_getWork", []interface{}{})
		if err != nil {
			return nil, fmt.Errorf("getWork failed: %w", err)
		}
	}

	var work [4]string
	if err := json.Unmarshal(result, &work); err != nil {
		return nil, fmt.Errorf("failed to parse work: %w", err)
	}

	return &WorkPackage{
		HeaderHash:  work[0],
		SeedHash:    work[1],
		Target:      work[2],
		BlockNumber: work[3],
		ReceivedAt:  time.Now(),
	}, nil
}

// SubmitWork submits a solution to Geth
func (c *RPCClient) SubmitWork(nonce, headerHash, mixDigest string) (bool, error) {
	params := []interface{}{nonce, headerHash, mixDigest}

	// Try randomx_submitWork first, fallback to eth_submitWork
	result, err := c.call("randomx_submitWork", params)
	if err != nil {
		result, err = c.call("eth_submitWork", params)
		if err != nil {
			return false, fmt.Errorf("submitWork failed: %w", err)
		}
	}

	var accepted bool
	if err := json.Unmarshal(result, &accepted); err != nil {
		return false, fmt.Errorf("failed to parse submit result: %w", err)
	}

	return accepted, nil
}

// SubmitHashrate submits hashrate to Geth
func (c *RPCClient) SubmitHashrate(hashrate, id string) error {
	params := []interface{}{hashrate, id}

	_, err := c.call("eth_submitHashrate", params)
	if err != nil {
		return fmt.Errorf("submitHashrate failed: %w", err)
	}

	return nil
}

// GetBlockNumber gets the current block number
func (c *RPCClient) GetBlockNumber() (uint64, error) {
	result, err := c.call("eth_blockNumber", []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("blockNumber failed: %w", err)
	}

	var blockNumber string
	if err := json.Unmarshal(result, &blockNumber); err != nil {
		return 0, fmt.Errorf("failed to parse block number: %w", err)
	}

	var num uint64
	fmt.Sscanf(blockNumber, "0x%x", &num)
	return num, nil
}

// CheckConnection verifies the RPC connection is working
func (c *RPCClient) CheckConnection() error {
	_, err := c.call("eth_blockNumber", []interface{}{})
	if err != nil {
		return fmt.Errorf("connection check failed: %w", err)
	}
	log.Println("✅ RPC connection verified")
	return nil
}
