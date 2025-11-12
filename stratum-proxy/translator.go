package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// Translator converts between Stratum/Monero format and Ethereum/Ducros format

// WorkToJob converts Geth work package to Stratum job
func WorkToJob(work *WorkPackage, jobID string, algo string) (*Job, error) {
	// Parse block number
	var blockNum uint64
	fmt.Sscanf(work.BlockNumber, "0x%x", &blockNum)

	// Parse target to difficulty
	difficulty, err := TargetToDifficulty(work.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target: %w", err)
	}

	// Create blob from header hash
	// For RandomX, the blob is: SealHash (32 bytes) + reserved nonce space (8 bytes)
	blob := createBlob(work.HeaderHash)

	job := &Job{
		JobID:      jobID,
		Blob:       blob,
		Target:     DifficultyToStratumTarget(difficulty), // Use Stratum format (8 chars, little-endian)
		Algo:       algo,
		Height:     blockNum,
		SeedHash:   work.SeedHash,
		HeaderHash: work.HeaderHash,
		Difficulty: difficulty,
	}

	return job, nil
}

// createBlobRxEth creates an rx-eth-v1 format blob compatible with both xmrig and geth
// Format: headerHash(32) || extraNonce4(4) || const3(3) || nonce4_placeholder(4) = 43 bytes
// This format allows geth to reconstruct the exact same preimage for RandomX verification
// extraNonce4 will be stored in header.Extra for geth to retrieve during verification
func createBlobRxEth(headerHash string, extraNonce uint32) string {
	// Remove 0x prefix if present
	headerHash = strings.TrimPrefix(headerHash, "0x")

	// Ensure it's 32 bytes (64 hex chars)
	if len(headerHash) != 64 {
		headerHash = fmt.Sprintf("%064s", headerHash)
	}

	// Build rx-eth-v1 blob:
	var blob strings.Builder

	// 1. Header hash (32 bytes)
	blob.WriteString(headerHash)

	// 2. ExtraNonce (4 bytes, little-endian)
	blob.WriteString(fmt.Sprintf("%02x%02x%02x%02x",
		byte(extraNonce&0xFF),
		byte((extraNonce>>8)&0xFF),
		byte((extraNonce>>16)&0xFF),
		byte((extraNonce>>24)&0xFF)))

	// 3. Constant padding (3 bytes) - ensures nonce at offset 39
	blob.WriteString("000000")

	// 4. Nonce placeholder (4 bytes) - will be filled by miner
	blob.WriteString("00000000")

	// Total: 43 bytes (86 hex chars)
	return blob.String()
}

// ExtractNonceFromBlobRxEth extracts the miner nonce from rx-eth-v1 blob
// Blob structure:
// - HeaderHash: 32 bytes (offset 0-63 hex)
// - ExtraNonce: 4 bytes (offset 64-71 hex)
// - Const3: 3 bytes (offset 72-77 hex)
// - Nonce4: 4 bytes (offset 78-85 hex) <- THIS IS WHAT WE EXTRACT
func ExtractNonceFromBlobRxEth(blob string) (string, error) {
	// Remove any whitespace
	blob = strings.TrimSpace(blob)

	// Expected length: 86 hex chars (43 bytes)
	if len(blob) < 86 {
		return "", fmt.Errorf("invalid blob length: %d (expected at least 86)", len(blob))
	}

	// Extract nonce4 from rx-eth-v1 structure (4 bytes at offset 39)
	// In hex: offset 78-85 (4 bytes = 8 hex chars)
	nonceHex := blob[78:86]

	// Decode to verify it's valid hex
	nonceBytes, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", fmt.Errorf("invalid nonce hex: %w", err)
	}

	// Verify we have exactly 4 bytes
	if len(nonceBytes) != 4 {
		return "", fmt.Errorf("invalid nonce length: %d (expected 4)", len(nonceBytes))
	}

	// Return just the 4-byte nonce (will be combined with extraNonce later)
	return nonceHex, nil
}

// CalculateHash computes the RandomX hash for verification
// This is a simplified version - actual hashing should use RandomX
func CalculateHash(headerHash, nonce string) string {
	// In production, this would:
	// 1. Initialize RandomX with seed
	// 2. Compute RandomX(headerHash + nonce)
	// 3. Return the hash

	// For now, we let Geth do the verification via submitWork
	// Return a placeholder that indicates we need Geth to verify
	return "0x" + strings.Repeat("0", 64)
}

// TargetToDifficulty converts a hex target to difficulty
func TargetToDifficulty(targetHex string) (uint64, error) {
	// Remove 0x prefix
	targetHex = strings.TrimPrefix(targetHex, "0x")

	// Parse target as big.Int
	target := new(big.Int)
	target.SetString(targetHex, 16)

	// Difficulty = 2^256 / target
	maxTarget := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	difficulty := new(big.Int).Div(maxTarget, target)

	return difficulty.Uint64(), nil
}

// DifficultyToTarget converts difficulty to hex target (32 bytes for Ethereum)
func DifficultyToTarget(difficulty uint64) string {
	// target = 2^256 / difficulty
	maxTarget := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	target := new(big.Int).Div(maxTarget, big.NewInt(int64(difficulty)))

	// Convert to hex string (32 bytes = 64 hex chars)
	targetHex := fmt.Sprintf("%064x", target)

	return targetHex
}

// DifficultyToStratumTarget converts difficulty to Stratum/xmrig target format
// Returns 4-byte little-endian hex string (8 characters)
func DifficultyToStratumTarget(difficulty uint64) string {
	// For Stratum/CryptoNote: target = 0xFFFFFFFF / difficulty
	// This gives a 32-bit target value
	maxTarget := uint64(0xFFFFFFFF)

	var target uint32
	if difficulty > maxTarget {
		// If difficulty is too high, use minimum target
		target = 1
	} else if difficulty == 0 {
		// Avoid division by zero
		target = 0xFFFFFFFF
	} else {
		target = uint32(maxTarget / difficulty)
	}

	// Convert to little-endian hex string (4 bytes = 8 hex chars)
	// Little-endian: least significant byte first
	b0 := byte(target & 0xFF)
	b1 := byte((target >> 8) & 0xFF)
	b2 := byte((target >> 16) & 0xFF)
	b3 := byte((target >> 24) & 0xFF)

	return fmt.Sprintf("%02x%02x%02x%02x", b0, b1, b2, b3)
}

// ValidateShare checks if a share meets the required difficulty
func ValidateShare(resultHash string, targetDiff uint64) (bool, uint64) {
	// Remove 0x prefix
	resultHash = strings.TrimPrefix(resultHash, "0x")

	// Parse hash as big.Int
	hashInt := new(big.Int)
	hashInt.SetString(resultHash, 16)

	// Calculate difficulty achieved
	// difficulty = 2^256 / hash
	maxTarget := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	achievedDiff := new(big.Int).Div(maxTarget, hashInt)

	actualDiff := achievedDiff.Uint64()

	// Check if achieved difficulty meets target
	valid := actualDiff >= targetDiff

	return valid, actualDiff
}

// FormatNonceForEthereum formats nonce for Ethereum (0x + 16 hex = 8 bytes)
func FormatNonceForEthereum(nonce uint64) string {
	return fmt.Sprintf("0x%016x", nonce)
}

// FormatHashForEthereum formats hash for Ethereum (0x + 64 hex = 32 bytes)
func FormatHashForEthereum(hash string) string {
	hash = strings.TrimPrefix(hash, "0x")
	return "0x" + strings.ToLower(hash)
}

// AdjustDifficulty adjusts miner difficulty based on share submission rate
func AdjustDifficulty(currentDiff uint64, shareRate float64, targetRate float64) uint64 {
	// Target: ~1 share every 30 seconds per miner
	// shareRate = shares per minute
	// targetRate = 2.0 (2 shares per minute = 1 per 30s)

	ratio := shareRate / targetRate

	var newDiff uint64
	if ratio > 2.0 {
		// Too many shares, increase difficulty
		newDiff = uint64(float64(currentDiff) * 1.5)
	} else if ratio < 0.5 {
		// Too few shares, decrease difficulty
		newDiff = uint64(float64(currentDiff) * 0.75)
	} else {
		// Good rate, keep current
		newDiff = currentDiff
	}

	// Enforce limits
	minDiff := uint64(1000)
	maxDiff := uint64(1000000000)

	if newDiff < minDiff {
		newDiff = minDiff
	}
	if newDiff > maxDiff {
		newDiff = maxDiff
	}

	return newDiff
}

// EstimateHashrate estimates hashrate from difficulty and time
func EstimateHashrate(difficulty uint64, timeSeconds float64) float64 {
	// Hashrate (H/s) = Difficulty / Time
	if timeSeconds <= 0 {
		return 0
	}
	return float64(difficulty) / timeSeconds
}
