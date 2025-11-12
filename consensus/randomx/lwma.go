// Copyright 2024 The go-ethereum Authors
// LWMA (Linearly Weighted Moving Average) Difficulty Algorithm
// Optimized for CPU-minable RandomX chains

package randomx

import (
	"math/big"

	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

const (
	LWMAWindowSize              = 60
	LWMATargetBlockTime         = 13
	LWMAMinDifficulty           = 1
	LWMAMaxAdjustmentUp         = 2
	LWMAMaxAdjustmentDown       = 2
	LWMATimestampMaxFutureDrift = 15
	LWMATimestampMaxPastDrift   = 91
)

// CalcDifficultyLWMA calculates difficulty using LWMA-3 algorithm
func CalcDifficultyLWMA(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	if parent.Number.Uint64() < LWMAWindowSize {
		return big.NewInt(LWMAMinDifficulty)
	}

	var (
		blockTimes            = make([]uint64, LWMAWindowSize)
		difficulties          = make([]*big.Int, LWMAWindowSize)
		weightedSolveTimeSum  = big.NewInt(0)
		weightSum             = big.NewInt(0)
		weightedDifficultySum = big.NewInt(0)
		currentBlock          = parent
	)

	// Collect last N blocks
	for i := LWMAWindowSize - 1; i >= 0; i-- {
		if currentBlock == nil || currentBlock.Number.Uint64() == 0 {
			return big.NewInt(LWMAMinDifficulty)
		}
		blockTimes[i] = currentBlock.Time
		difficulties[i] = new(big.Int).Set(currentBlock.Difficulty)
		if i > 0 {
			currentBlock = chain.GetHeader(currentBlock.ParentHash, currentBlock.Number.Uint64()-1)
		}
	}

	// Calculate LWMA
	for i := 0; i < LWMAWindowSize-1; i++ {
		solveTime := blockTimes[i+1] - blockTimes[i]
		if solveTime == 0 {
			solveTime = 1
		}
		if solveTime > 6*LWMATargetBlockTime {
			solveTime = 6 * LWMATargetBlockTime
		}

		weight := int64(i + 1)
		bigWeight := big.NewInt(weight)
		bigSolveTime := big.NewInt(int64(solveTime))

		temp1 := new(big.Int).Mul(bigSolveTime, bigWeight)
		weightedSolveTimeSum.Add(weightedSolveTimeSum, temp1)
		weightSum.Add(weightSum, bigWeight)
		temp2 := new(big.Int).Mul(temp1, difficulties[i])
		weightedDifficultySum.Add(weightedDifficultySum, temp2)
	}

	nextDifficulty := new(big.Int).Div(weightedDifficultySum, weightedSolveTimeSum)

	minDiff := big.NewInt(LWMAMinDifficulty)
	if nextDifficulty.Cmp(minDiff) < 0 {
		nextDifficulty.Set(minDiff)
	}

	maxIncrease := new(big.Int).Mul(parent.Difficulty, big.NewInt(LWMAMaxAdjustmentUp))
	if nextDifficulty.Cmp(maxIncrease) > 0 {
		nextDifficulty.Set(maxIncrease)
	}

	maxDecrease := new(big.Int).Div(parent.Difficulty, big.NewInt(LWMAMaxAdjustmentDown))
	if nextDifficulty.Cmp(maxDecrease) < 0 {
		nextDifficulty.Set(maxDecrease)
	}

	return nextDifficulty
}

// ShouldUseLWMA determines whether to use LWMA
func ShouldUseLWMA(config *params.ChainConfig, blockNumber *big.Int) bool {
	if config.RandomX != nil && config.RandomX.LWMAActivationBlock != nil {
		return blockNumber.Cmp(config.RandomX.LWMAActivationBlock) >= 0
	}
	if config.RandomX != nil {
		return true
	}
	return false
}
