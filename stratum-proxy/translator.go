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

// createBlob creates a Monero-style blob from Ethereum header hash
// Format: HeaderHash (32 bytes) + Nonce placeholder (8 bytes) = 40 bytes total
func createBlob(headerHash string) string {
	// Remove 0x prefix if present
	headerHash = strings.TrimPrefix(headerHash, "0x")

	// Ensure it's 32 bytes (64 hex chars)
	if len(headerHash) != 64 {
		// Pad or truncate
		headerHash = fmt.Sprintf("%064s", headerHash)
	}

	// Append 8 bytes (16 hex chars) of zeros for nonce placeholder
	blob := headerHash + "0000000000000000"

	return blob
}

// ExtractNonceFromBlob extracts the nonce from a submitted blob
// The nonce is in the last 8 bytes of the blob (bytes 32-39)
func ExtractNonceFromBlob(blob string) (string, error) {
	// Remove any whitespace
	blob = strings.TrimSpace(blob)

	// Expected length: 80 hex chars (40 bytes)
	if len(blob) != 80 {
		return "", fmt.Errorf("invalid blob length: %d (expected 80)", len(blob))
	}

	// Extract nonce from bytes 32-39 (hex chars 64-79)
	nonceHex := blob[64:80]

	// Monero/xmrig sends nonce in little-endian format within the blob
	// Ethereum expects the nonce as a hex string "0x" + 16 hex chars
	// Since both use little-endian, we can pass through directly
	nonceBytes, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", fmt.Errorf("invalid nonce hex: %w", err)
	}

	// Verify we have exactly 8 bytes
	if len(nonceBytes) != 8 {
		return "", fmt.Errorf("invalid nonce length: %d (expected 8)", len(nonceBytes))
	}

	// Format as 0x + 16 hex chars (8 bytes little-endian)
	// Note: Ethereum BlockNonce is [8]byte and we preserve byte order from blob
	nonce := "0x" + nonceHex

	return nonce, nil
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
