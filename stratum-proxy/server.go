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
	config           *ServerConfig
	rpcClient        *RPCClient
	listener         net.Listener
	miners           map[string]*Miner
	minersMu         sync.RWMutex
	currentWork      *WorkPackage
	currentJob       *Job
	workMu           sync.RWMutex
	stats            *Stats
	jobCounter       uint64
	connectionCount  int           // Current number of connections
	connectionCountMu sync.Mutex   // Protects connectionCount
	stopCh           chan struct{}
	wg               sync.WaitGroup
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

		// Check connection limit (DoS protection)
		if s.config.MaxConnections > 0 {
			s.connectionCountMu.Lock()
			if s.connectionCount >= s.config.MaxConnections {
				s.connectionCountMu.Unlock()
				log.Printf("🚫 Connection limit reached (%d), rejecting %s",
					s.config.MaxConnections, conn.RemoteAddr())
				conn.Close()
				continue
			}
			s.connectionCount++
			s.connectionCountMu.Unlock()
		}

		s.wg.Add(1)
		go s.handleMiner(conn)
	}
}

// handleMiner handles a single miner connection
func (s *Server) handleMiner(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Decrement connection count on exit
	defer func() {
		if s.config.MaxConnections > 0 {
			s.connectionCountMu.Lock()
			s.connectionCount--
			s.connectionCountMu.Unlock()
		}
	}()

	minerID := conn.RemoteAddr().String()
	log.Printf("🔌 New connection from %s", minerID)

	// Set absolute deadline for connection (1 hour max)
	// Prevents zombie connections from staying forever
	// Each read will also have its own 5-minute deadline (set below)
	conn.SetDeadline(time.Now().Add(1 * time.Hour))

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Generate random 4-byte extraNonce for rx-eth-v1 format
	var extraNonceBytes [4]byte
	rand.Read(extraNonceBytes[:])
	extraNonce := binary.LittleEndian.Uint32(extraNonceBytes[:])

	// Create JSON encoder for pushing notifications
	jsonWriter := json.NewEncoder(writer)

	miner := &Miner{
		ID:             minerID,
		Writer:         jsonWriter,
		BufferedWriter: writer, // Store for Flush() after notifications
		Difficulty:     uint64(s.config.InitialDiff),
		ExtraNonce:     extraNonce,
		LastActivity:   time.Now(),
		LastShareTime:  time.Now(),
		ShareTimes:     make([]time.Time, 0, 100),
	}

	// Register miner
	s.minersMu.Lock()
	s.miners[minerID] = miner
	s.minersMu.Unlock()

	defer func() {
		// Clean up miner resources before removal
		miner.mu.Lock()
		miner.ShareTimes = nil // Release slice memory
		miner.CurrentJob = nil // Release job reference
		miner.mu.Unlock()

		miner.writerMu.Lock()
		miner.Writer = nil         // Release encoder
		miner.BufferedWriter = nil // Release buffer
		miner.writerMu.Unlock()

		// Remove from miners map
		s.minersMu.Lock()
		delete(s.miners, minerID)
		s.minersMu.Unlock()

		if s.config.Verbose {
			log.Printf("👋 Miner %s disconnected (cleaned up)", minerID)
		} else {
			log.Printf("👋 Miner %s disconnected", minerID)
		}
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

		// CRITICAL: Protect write with mutex to avoid race with pushJob
		miner.writerMu.Lock()
		if _, err := writer.Write(responseJSON); err != nil {
			miner.writerMu.Unlock()
			log.Printf("Write error to %s: %v", minerID, err)
			return
		}
		writer.Flush()
		miner.writerMu.Unlock()

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
	blob, err := createBlobRxEth(job.HeaderHash, miner.ExtraNonce)
	if err != nil {
		log.Printf("❌ Failed to create blob for %s: %v", miner.ID, err)
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: fmt.Sprintf("Invalid work data: %v", err),
			},
		}
	}

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

	// Check if miner is banned
	miner.mu.RLock()
	if miner.Banned {
		banReason := miner.BanReason
		miner.mu.RUnlock()
		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Banned: " + banReason,
			},
		}
	}
	miner.mu.RUnlock()

	// Rate limiting check (DoS protection)
	// NOTE: We check rate BEFORE validating share to prevent spam
	// But we DON'T update timestamp yet - only after valid share
	if s.config.ShareRateLimit > 0 {
		now := time.Now()
		miner.mu.Lock()
		timeSinceLastShare := now.Sub(miner.LastShareSubmitTime).Seconds()
		minInterval := 1.0 / s.config.ShareRateLimit // Minimum seconds between shares

		if timeSinceLastShare < minInterval && !miner.LastShareSubmitTime.IsZero() {
			miner.mu.Unlock()
			if s.config.Verbose {
				log.Printf("⚠️  Rate limit exceeded for %s (%.3fs < %.3fs)",
					miner.ID, timeSinceLastShare, minInterval)
			}
			return &StratumResponse{
				ID:      req.ID,
				JSONRPC: "2.0",
				Error: &StratumError{
					Code:    -1,
					Message: "Rate limit exceeded",
				},
			}
		}
		// Don't update timestamp yet - wait until share is validated
		miner.mu.Unlock()
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

		// Update invalid share count with lock
		miner.mu.Lock()
		miner.SharesInvalid++
		miner.mu.Unlock()

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

		// Update invalid share count with lock
		miner.mu.Lock()
		miner.SharesInvalid++
		miner.mu.Unlock()

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

		// Update invalid share count with lock
		miner.mu.Lock()
		miner.SharesInvalid++
		miner.mu.Unlock()

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

	// Pool validation: check share difficulty WITHOUT submitting to geth
	// This allows us to validate low-difficulty shares locally
	shareValid, shareDiff := ValidateShare(resultStr, miner.Difficulty)

	if s.config.Verbose {
		log.Printf("🔍 Share validation: hash=%s shareDiff=%d minerDiff=%d valid=%v",
			resultStr[:18]+"...", shareDiff, miner.Difficulty, shareValid)
	}

	if !shareValid {
		log.Printf("❌ Share below difficulty from %s (got %d, need %d)",
			miner.ID, shareDiff, miner.Difficulty)

		// Update invalid share stats with locking
		miner.mu.Lock()
		miner.SharesInvalid++
		miner.SharesInvalidStreak++

		// Check ban threshold
		if s.config.MaxInvalidStreak > 0 && miner.SharesInvalidStreak >= s.config.MaxInvalidStreak {
			miner.Banned = true
			miner.BanReason = fmt.Sprintf("Exceeded max invalid shares (%d consecutive)", miner.SharesInvalidStreak)
			log.Printf("🚫 BANNED miner %s: %s", miner.ID, miner.BanReason)
		}
		banned := miner.Banned
		miner.mu.Unlock()

		s.stats.RecordShare(false)

		// Disconnect banned miners
		if banned {
			return &StratumResponse{
				ID:      req.ID,
				JSONRPC: "2.0",
				Error: &StratumError{
					Code:    -1,
					Message: "Banned: " + miner.BanReason,
				},
			}
		}

		return &StratumResponse{
			ID:      req.ID,
			JSONRPC: "2.0",
			Error: &StratumError{
				Code:    -1,
				Message: "Share difficulty too low",
			},
		}
	}

	// Share is valid for the miner's difficulty
	log.Printf("✅ Valid share from %s (diff: %d, actual: %d)", miner.ID, miner.Difficulty, shareDiff)

	// Update miner stats with proper locking
	miner.mu.Lock()
	miner.SharesValid++
	miner.SharesInvalidStreak = 0 // Reset invalid streak on valid share
	miner.TotalDifficulty += miner.Difficulty // Track contribution for pool payouts

	// Update rate limit timestamp ONLY on valid share (prevents rate limit bypass)
	if s.config.ShareRateLimit > 0 {
		miner.LastShareSubmitTime = time.Now()
	}

	now := time.Now()

	// Track share times for hashrate calculation (rolling window of last 100 shares)
	miner.ShareTimes = append(miner.ShareTimes, now)
	if len(miner.ShareTimes) > 100 {
		miner.ShareTimes = miner.ShareTimes[1:]
	}

	// Calculate hashrate from share times
	if len(miner.ShareTimes) >= 2 {
		// Time span for shares
		timeSpan := miner.ShareTimes[len(miner.ShareTimes)-1].Sub(miner.ShareTimes[0]).Seconds()
		if timeSpan > 0 {
			// Hashrate = (difficulty * num_shares) / time
			numShares := float64(len(miner.ShareTimes))
			miner.Hashrate = (float64(miner.Difficulty) * numShares) / timeSpan
		}
	}

	// Adjust difficulty based on configured window
	oldDiff := miner.Difficulty
	difficultyChanged := false
	varDiffWindow := s.config.VarDiffWindow
	if varDiffWindow == 0 {
		varDiffWindow = 10 // Default fallback
	}

	if miner.SharesValid%varDiffWindow == 0 && len(miner.ShareTimes) >= int(varDiffWindow) {
		// Calculate share rate based on configured window
		lastN := miner.ShareTimes[len(miner.ShareTimes)-int(varDiffWindow):]
		timeSpan := now.Sub(lastN[0]).Minutes()
		if timeSpan > 0 {
			shareRate := float64(varDiffWindow) / timeSpan // shares per minute
			// Convert target from seconds to shares/minute
			targetRate := 60.0 / s.config.VarDiffTarget // shares per minute
			miner.Difficulty = AdjustDifficulty(miner.Difficulty, shareRate, targetRate)

			if miner.Difficulty != oldDiff {
				difficultyChanged = true
				log.Printf("📊 Adjusted difficulty for %s: %d → %d (hashrate: %.2f H/s, target: %.1fs/share)",
					miner.ID, oldDiff, miner.Difficulty, miner.Hashrate, s.config.VarDiffTarget)
			}
		}
	}

	currentJob := miner.CurrentJob
	miner.LastShareTime = now
	miner.mu.Unlock()

	// If difficulty changed, push new job with updated target
	if difficultyChanged && currentJob != nil {
		s.pushJob(miner, currentJob)
	}

	s.stats.RecordShare(true)

	// Check if this share meets NETWORK difficulty (potential block)
	// Get network difficulty from current work
	s.workMu.RLock()
	networkDifficulty := uint64(0)
	if s.currentWork != nil {
		networkDifficulty, _ = TargetToDifficulty(s.currentWork.Target)
	}
	s.workMu.RUnlock()

	isBlock := shareDiff >= networkDifficulty && networkDifficulty > 0

	if isBlock {
		log.Printf("🎉 BLOCK CANDIDATE from %s! (diff: %d >= %d)", miner.ID, shareDiff, networkDifficulty)

		// Submit to geth for block validation
		accepted, err := s.rpcClient.SubmitWork(
			nonceHex,
			miner.CurrentJob.HeaderHash,
			resultStr, // Use result as mixDigest
		)

		if err != nil {
			log.Printf("⚠️  Block submission error for %s: %v", miner.ID, err)
			// Don't fail the share just because geth submission failed
			// The share itself is valid for pool purposes
		} else if accepted {
			log.Printf("🎉🎉🎉 BLOCK ACCEPTED by network from %s!", miner.ID)
			s.stats.RecordBlock()

			// Update miner's block count
			miner.mu.Lock()
			miner.BlocksFound++
			minerAddress := miner.Address
			minerBlocks := miner.BlocksFound
			miner.mu.Unlock()

			// Pool fee information
			if s.config.PoolAddress != "" {
				log.Printf("💰 Block mined to pool address: %s", s.config.PoolAddress)
				log.Printf("💰 Pool fee: %.2f%% (operator keeps fee, pays out based on shares)", s.config.PoolFee)
				log.Printf("💰 Miner %s contributed this block (address: %s, total blocks: %d)",
					miner.ID, minerAddress, minerBlocks)
				// TODO: Implement automated payout system based on share contributions
				// For now, operator must manually calculate payouts using share logs
			} else {
				log.Printf("💰 Block reward goes to miner address: %s", minerAddress)
			}
		} else {
			log.Printf("⚠️  Block rejected by network from %s (but share is valid)", miner.ID)
		}
	}

	// Return success to miner (share was valid for their difficulty)
	return &StratumResponse{
		ID:      req.ID,
		JSONRPC: "2.0",
		Result:  map[string]interface{}{"status": "OK"},
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

// broadcastJob sends new job to all connected miners via mining.notify
func (s *Server) broadcastJob(job *Job) {
	s.minersMu.RLock()
	miners := make([]*Miner, 0, len(s.miners))
	for _, m := range s.miners {
		miners = append(miners, m)
	}
	s.minersMu.RUnlock()

	successCount := 0
	for _, miner := range miners {
		// Skip banned miners
		miner.mu.RLock()
		banned := miner.Banned
		miner.mu.RUnlock()

		if banned {
			continue
		}

		// Push job returns error internally, counted as failed
		s.pushJob(miner, job)
		successCount++
	}

	if s.config.Verbose {
		log.Printf("📢 Broadcasted job %s to %d/%d miners", job.JobID, successCount, len(miners))
	}
}

// pushJob pushes a new job to a specific miner using mining.notify
func (s *Server) pushJob(miner *Miner, job *Job) {
	// Lock miner state to read values and update current job
	miner.mu.Lock()
	miner.CurrentJob = job
	extraNonce := miner.ExtraNonce
	difficulty := miner.Difficulty
	minerID := miner.ID
	miner.mu.Unlock()

	// Create rx-eth-v1 blob with miner's extraNonce
	blob, err := createBlobRxEth(job.HeaderHash, extraNonce)
	if err != nil {
		log.Printf("❌ Failed to create blob for push to %s: %v", minerID, err)
		return // Don't send invalid job
	}

	// Create job response for xmrig RandomX
	// CRITICAL: Use miner's CURRENT difficulty (not network difficulty)
	jobResponse := JobResponse{
		JobID:     job.JobID,
		Algo:      "rx/0",
		SeedHash:  strings.TrimPrefix(job.SeedHash, "0x"),
		Height:    job.Height,
		Blob:      blob,
		Target:    DifficultyToStratumTarget(difficulty),
		CleanJobs: true, // Discard old jobs
	}

	// Create JSON-RPC notification (no ID for notifications)
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "job",
		"params":  jobResponse,
	}

	// Send notification - CRITICAL: Use dedicated writer mutex to avoid race with handleMiner
	miner.writerMu.Lock()
	defer miner.writerMu.Unlock()

	if miner.Writer != nil {
		if err := miner.Writer.Encode(notification); err != nil {
			if s.config.Verbose {
				log.Printf("⚠️  Failed to push job to %s: %v", minerID, err)
			}
			return
		}

		// CRITICAL: Flush buffer so notification is actually sent!
		if miner.BufferedWriter != nil {
			if err := miner.BufferedWriter.Flush(); err != nil {
				if s.config.Verbose {
					log.Printf("⚠️  Failed to flush buffer for %s: %v", minerID, err)
				}
				return
			}
		}

		if s.config.Verbose {
			log.Printf("📤 Pushed job %s to %s (diff: %d)", job.JobID, minerID, difficulty)
		}
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
	var totalContribution uint64

	// Collect miner stats
	type MinerStats struct {
		ID              string
		Address         string
		Hashrate        float64
		SharesValid     uint64
		TotalDifficulty uint64
		BlocksFound     uint64
	}
	var minerStats []MinerStats

	for _, miner := range s.miners {
		miner.mu.RLock()
		lastActivity := miner.LastActivity
		hashrate := miner.Hashrate
		stats := MinerStats{
			ID:              miner.ID,
			Address:         miner.Address,
			Hashrate:        hashrate,
			SharesValid:     miner.SharesValid,
			TotalDifficulty: miner.TotalDifficulty,
			BlocksFound:     miner.BlocksFound,
		}
		miner.mu.RUnlock()

		if time.Since(lastActivity) < 2*time.Minute {
			activeMiners++
			totalHashrate += hashrate
		}
		totalContribution += stats.TotalDifficulty
		minerStats = append(minerStats, stats)
	}
	s.minersMu.RUnlock()

	// Update stats BEFORE reading them to ensure fresh values
	s.stats.UpdateMiners(totalMiners, activeMiners)
	s.stats.UpdateHashrate(totalHashrate)

	// Now get the fresh stats
	total, active, shares, valid, invalid, blocks, hashrate, uptime := s.stats.GetStats()

	log.Printf("📊 Stats: Miners=%d/%d Shares=%d/%d/%d Blocks=%d Hashrate=%.2f H/s Uptime=%s",
		active, total, valid, invalid, shares, blocks, hashrate, uptime.Round(time.Second))

	// Pool fee stats (if pool mode enabled)
	if s.config.PoolAddress != "" && totalContribution > 0 && s.config.Verbose {
		log.Printf("💰 Pool Contributions:")
		for _, ms := range minerStats {
			if ms.SharesValid > 0 {
				percentage := (float64(ms.TotalDifficulty) / float64(totalContribution)) * 100.0
				log.Printf("   %s: %.2f%% (%d shares, %d blocks, %.2f H/s)",
					ms.Address, percentage, ms.SharesValid, ms.BlocksFound, ms.Hashrate)
			}
		}
	}
}
