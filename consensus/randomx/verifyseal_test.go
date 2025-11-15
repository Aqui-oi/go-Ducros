// Copyright 2024 The go-ethereum Authors
package randomx

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestVerifySealFake(t *testing.T) {
	engine := NewFaker()
	defer engine.Close()

	header := &types.Header{
		ParentHash: common.HexToHash("0x1234"),
		Number:     big.NewInt(1),
		Difficulty: big.NewInt(1000),
		Time:       uint64(time.Now().Unix()),
		Coinbase:   common.HexToAddress("0xabcd"),
		GasLimit:   5000000,
		GasUsed:    0,
		Nonce:      types.EncodeNonce(0),
		MixDigest:  common.Hash{},
	}

	if err := engine.verifyPoW(header); err != nil {
		t.Errorf("Fake engine should accept any header, got error: %v", err)
	}
}

func TestSealHash(t *testing.T) {
	engine := New(&Config{PowMode: ModeNormal, LightMode: true})
	defer engine.Close()

	header := &types.Header{
		ParentHash:  common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		UncleHash:   types.EmptyUncleHash,
		Coinbase:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Root:        common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		TxHash:      types.EmptyTxsHash,
		ReceiptHash: types.EmptyReceiptsHash,
		Bloom:       types.Bloom{},
		Difficulty:  big.NewInt(131072),
		Number:      big.NewInt(1),
		GasLimit:    5000000,
		GasUsed:     0,
		Time:        1234567890,
		Extra:       []byte("go-ducros"),
		MixDigest:   common.Hash{},
		Nonce:       types.EncodeNonce(0),
	}

	sealHash1 := engine.SealHash(header)
	sealHash2 := engine.SealHash(header)
	if sealHash1 != sealHash2 {
		t.Error("SealHash should be deterministic")
	}

	// Changing nonce should NOT affect seal hash
	header.Nonce = types.EncodeNonce(12345)
	sealHash3 := engine.SealHash(header)
	if sealHash1 != sealHash3 {
		t.Error("SealHash should not include nonce")
	}

	// Changing other fields SHOULD affect seal hash
	header.Number = big.NewInt(2)
	sealHash4 := engine.SealHash(header)
	if sealHash1 == sealHash4 {
		t.Error("SealHash should change when block number changes")
	}

	t.Logf("SealHash test passed: %x", sealHash1)
}

func TestVerifyRandomX(t *testing.T) {
	tests := []struct {
		name       string
		hash       common.Hash
		difficulty *big.Int
		shouldPass bool
	}{
		{
			name:       "Hash below target (valid)",
			hash:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			difficulty: big.NewInt(1000000),
			shouldPass: true,
		},
		{
			name:       "Hash above target (invalid)",
			hash:       common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			difficulty: big.NewInt(1000000000),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyRandomX(tt.hash, tt.difficulty)
			if result != tt.shouldPass {
				t.Errorf("verifyRandomX(%x, %v) = %v, want %v",
					tt.hash, tt.difficulty, result, tt.shouldPass)
			}
		})
	}
}
