package main

import (
	"bufio"
	"encoding/json"
	"sync"
	"time"
)

// Stratum protocol types

// StratumRequest represents a Stratum JSON-RPC request
// Params can be either an array or an object depending on the method
type StratumRequest struct {
	ID     interface{}      `json:"id"`
	Method string           `json:"method"`
	Params json.RawMessage  `json:"params"` // Can be array or object
}

// GetParamsArray returns params as array (for standard stratum methods)
func (r *StratumRequest) GetParamsArray() ([]interface{}, error) {
	var params []interface{}
	if err := json.Unmarshal(r.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
}

// GetParamsObject returns params as object (for xmrig login method)
func (r *StratumRequest) GetParamsObject() (map[string]interface{}, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(r.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
}

// StratumResponse represents a Stratum JSON-RPC response
type StratumResponse struct {
	ID      interface{} `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"` // Only include on errors
}

// StratumError represents a Stratum error
type StratumError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Job represents a mining job for Stratum clients
type Job struct {
	JobID       string    `json:"job_id"`
	Blob        string    `json:"blob"`          // Block template (hex)
	Target      string    `json:"target"`        // Difficulty target (hex)
	Algo        string    `json:"algo,omitempty"`          // "rx/0" for RandomX (optional)
	Height      uint64    `json:"height,omitempty"`        // Block number (optional)
	SeedHash    string    `json:"seed_hash,omitempty"`     // RandomX seed (optional)
	HeaderHash  string    `json:"-"`             // Internal: header hash for verification
	Difficulty  uint64    `json:"-"`             // Internal: actual difficulty
	CreatedAt   time.Time `json:"-"`             // Internal: job creation time
}

// JobResponse is the complete job format for xmrig RandomX
type JobResponse struct {
	JobID     string `json:"job_id"`
	Algo      string `json:"algo"`       // "rx/0" for RandomX
	SeedHash  string `json:"seed_hash"`  // RandomX epoch seed (64 hex)
	Height    uint64 `json:"height"`     // Block height
	Blob      string `json:"blob"`       // Block template
	Target    string `json:"target"`     // Difficulty target (8 hex, LE)
	CleanJobs bool   `json:"clean_jobs"` // true = discard previous jobs
}

// Miner represents a connected miner
type Miner struct {
	ID            string                  // Unique miner ID
	Writer        *json.Encoder           // JSON encoder for pushing notifications
	BufferedWriter *bufio.Writer          // Underlying buffered writer (for Flush)
	Agent         string                  // Miner software (e.g., "xmrig/6.18.0")
	WorkerName    string                  // Worker name
	Address       string                  // Payout address
	Difficulty    uint64                  // Current difficulty
	CurrentJob    *Job                    // Current mining job
	ExtraNonce    uint32                  // 4-byte session-specific nonce for rx-eth-v1
	LastActivity  time.Time               // Last seen
	LastShareTime time.Time               // Time of last share submission
	ShareTimes    []time.Time             // Rolling window of share times (for hashrate)
	SharesValid   uint64                  // Valid shares submitted
	SharesInvalid uint64                  // Invalid shares
	SharesInvalidStreak uint64            // Consecutive invalid shares
	Hashrate      float64                  // Estimated hashrate (H/s)
	Banned        bool                     // Whether miner is banned
	BanReason     string                   // Reason for ban
	LastShareSubmitTime time.Time         // Last share submission time (for rate limiting)
	TotalDifficulty uint64                // Total difficulty contributed (for pool payouts)
	BlocksFound   uint64                  // Number of blocks found by this miner
	mu            sync.RWMutex            // Protects miner state
	writerMu      sync.Mutex              // Protects Writer and BufferedWriter from concurrent writes
}

// Share represents a submitted share
type Share struct {
	MinerID    string
	JobID      string
	Nonce      string  // Hex encoded nonce (8 bytes)
	Result     string  // Hex encoded hash (32 bytes)
	Difficulty uint64
	Timestamp  time.Time
}

// ServerConfig holds the proxy server configuration
type ServerConfig struct {
	ListenAddr         string
	GethRPC            string
	InitialDiff        float64
	PoolAddress        string
	PoolFee            float64
	Verbose            bool
	Algorithm          string
	VarDiffTarget      float64  // Target time between shares (seconds)
	VarDiffWindow      uint64   // Number of shares for vardiff calculation
	MaxInvalidStreak   uint64   // Max invalid shares before ban
	MaxConnections     int      // Max concurrent miner connections (0 = unlimited)
	ShareRateLimit     float64  // Max shares per second per miner (0 = unlimited)
}

// WorkPackage represents work from Geth (eth_getWork)
type WorkPackage struct {
	HeaderHash  string    // [0] - SealHash
	SeedHash    string    // [1] - Epoch seed
	Target      string    // [2] - Difficulty target
	BlockNumber string    // [3] - Block number (hex)
	ReceivedAt  time.Time
}

// Stats holds server statistics
type Stats struct {
	StartTime      time.Time
	TotalMiners    int
	ActiveMiners   int
	TotalShares    uint64
	ValidShares    uint64
	InvalidShares  uint64
	BlocksFound    uint64
	TotalHashrate  float64
	mu             sync.RWMutex
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
	}
}

// RecordShare records a share submission
func (s *Stats) RecordShare(valid bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalShares++
	if valid {
		s.ValidShares++
	} else {
		s.InvalidShares++
	}
}

// RecordBlock records a found block
func (s *Stats) RecordBlock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksFound++
}

// UpdateHashrate updates the network hashrate
func (s *Stats) UpdateHashrate(hashrate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalHashrate = hashrate
}

// UpdateMiners updates miner counts
func (s *Stats) UpdateMiners(total, active int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalMiners = total
	s.ActiveMiners = active
}

// GetStats returns current stats (thread-safe)
func (s *Stats) GetStats() (total, active int, shares, valid, invalid, blocks uint64, hashrate float64, uptime time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.TotalMiners, s.ActiveMiners, s.TotalShares, s.ValidShares,
	       s.InvalidShares, s.BlocksFound, s.TotalHashrate, time.Since(s.StartTime)
}
