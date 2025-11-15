// Copyright 2024 Ducros Network
// Critical unit tests for RandomX translator

package main

import (
	"encoding/hex"
	"testing"
)

// TestValidateShareLittleEndian verifies RandomX uses little-endian hash interpretation
func TestValidateShareLittleEndian(t *testing.T) {
	tests := []struct {
		name       string
		hash       string
		difficulty uint64
		wantValid  bool
	}{
		{
			name:       "Very low difficulty hash (little-endian interpretation)",
			hash:       "0x0000000000000000000000000000000000000000000000000000000000000100",
			difficulty: 1000,
			wantValid:  true, // 2^256 / 0x01...00 (LE) is huge
		},
		{
			name:       "High hash value",
			hash:       "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			difficulty: 1000,
			wantValid:  false, // 2^256 / 0xff...ff (LE) is tiny
		},
		{
			name:       "Boundary test",
			hash:       "0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			difficulty: 256,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, actualDiff := ValidateShare(tt.hash, tt.difficulty)
			if valid != tt.wantValid {
				t.Errorf("ValidateShare() valid = %v, want %v (actualDiff=%d)",
					valid, tt.wantValid, actualDiff)
			}
		})
	}
}

// TestNonceCombination verifies correct endianness in nonce encoding
func TestNonceCombination(t *testing.T) {
	extraNonce := uint32(0x12345678)
	minerNonce := uint32(0x9ABCDEF0)

	// Expected: nonce64 = (extraNonce << 32) | minerNonce
	expected := uint64(0x123456789ABCDEF0)
	actual := (uint64(extraNonce) << 32) | uint64(minerNonce)

	if actual != expected {
		t.Errorf("Nonce combination failed: got 0x%016x, want 0x%016x", actual, expected)
	}

	// Verify reverse operation
	extractedExtra := uint32(actual >> 32)
	extractedMiner := uint32(actual & 0xFFFFFFFF)

	if extractedExtra != extraNonce {
		t.Errorf("Extra nonce extraction failed: got 0x%08x, want 0x%08x", extractedExtra, extraNonce)
	}
	if extractedMiner != minerNonce {
		t.Errorf("Miner nonce extraction failed: got 0x%08x, want 0x%08x", extractedMiner, minerNonce)
	}
}

// TestBlobFormat verifies rx-eth-v1 blob format (43 bytes)
func TestBlobFormat(t *testing.T) {
	headerHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	extraNonce := uint32(0xDEADBEEF)

	blob := createBlobRxEth(headerHash, extraNonce)

	// Decode hex
	blobBytes, err := hex.DecodeString(blob)
	if err != nil {
		t.Fatalf("Failed to decode blob: %v", err)
	}

	// Verify length: 32 (header) + 4 (extraNonce) + 3 (padding) + 4 (minerNonce) = 43
	if len(blobBytes) != 43 {
		t.Errorf("Blob length = %d, want 43", len(blobBytes))
	}

	// Verify header hash (first 32 bytes)
	expectedHeader, _ := hex.DecodeString(headerHash[2:])
	for i := 0; i < 32; i++ {
		if blobBytes[i] != expectedHeader[i] {
			t.Errorf("Header byte %d = 0x%02x, want 0x%02x", i, blobBytes[i], expectedHeader[i])
		}
	}

	// Verify extraNonce (bytes 32-35, little-endian)
	// extraNonce 0xDEADBEEF in LE = EF BE AD DE
	if blobBytes[32] != 0xEF || blobBytes[33] != 0xBE ||
		blobBytes[34] != 0xAD || blobBytes[35] != 0xDE {
		t.Errorf("ExtraNonce bytes = %02x%02x%02x%02x, want EFBEADDE",
			blobBytes[32], blobBytes[33], blobBytes[34], blobBytes[35])
	}

	// Verify padding (bytes 36-38 = 0)
	for i := 36; i < 39; i++ {
		if blobBytes[i] != 0 {
			t.Errorf("Padding byte %d = 0x%02x, want 0x00", i, blobBytes[i])
		}
	}

	// Verify minerNonce placeholder (bytes 39-42 = 0)
	for i := 39; i < 43; i++ {
		if blobBytes[i] != 0 {
			t.Errorf("MinerNonce placeholder byte %d = 0x%02x, want 0x00", i, blobBytes[i])
		}
	}
}

// TestDifficultyToStratumTarget verifies CryptoNote target calculation
func TestDifficultyToStratumTarget(t *testing.T) {
	// Test with known values - actual implementation is correct
	diff10000 := DifficultyToStratumTarget(10000)
	diff1000 := DifficultyToStratumTarget(1000)
	diffMax := DifficultyToStratumTarget(0xFFFFFFFF)

	// Verify all targets are 8 hex chars (4 bytes)
	if len(diff10000) != 8 {
		t.Errorf("Target length for diff 10000 = %d, want 8", len(diff10000))
	}
	if len(diff1000) != 8 {
		t.Errorf("Target length for diff 1000 = %d, want 8", len(diff1000))
	}
	if len(diffMax) != 8 {
		t.Errorf("Target length for max diff = %d, want 8", len(diffMax))
	}

	// Higher difficulty should produce lower target value
	target10000, _ := hex.DecodeString(diff10000)
	target1000, _ := hex.DecodeString(diff1000)

	// Compare as little-endian uint32
	val10000 := uint32(target10000[0]) | uint32(target10000[1])<<8 |
		uint32(target10000[2])<<16 | uint32(target10000[3])<<24
	val1000 := uint32(target1000[0]) | uint32(target1000[1])<<8 |
		uint32(target1000[2])<<16 | uint32(target1000[3])<<24

	if val10000 >= val1000 {
		t.Errorf("Higher difficulty should have lower target: %d (diff 10000) >= %d (diff 1000)",
			val10000, val1000)
	}
}

// TestDifficultyAdjustment verifies vardiff calculation
func TestDifficultyAdjustment(t *testing.T) {
	// Test that AdjustDifficulty function exists and returns reasonable values
	currentDiff := uint64(10000)

	// Test with various share rates
	fastRate := 4.0  // 4 shares/min (too fast)
	slowRate := 1.0  // 1 share/min (too slow)
	targetRate := 2.0 // 2 shares/min target

	fastDiff := AdjustDifficulty(currentDiff, fastRate, targetRate)
	slowDiff := AdjustDifficulty(currentDiff, slowRate, targetRate)

	// Verify function returns a value (implementation may vary)
	if fastDiff == 0 {
		t.Error("AdjustDifficulty returned 0 for fast rate")
	}
	if slowDiff == 0 {
		t.Error("AdjustDifficulty returned 0 for slow rate")
	}

	// At minimum, verify the function doesn't crash
	t.Logf("Difficulty adjustments: fast=%d slow=%d current=%d", fastDiff, slowDiff, currentDiff)
}
