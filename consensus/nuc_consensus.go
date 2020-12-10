// Copyright 2019 The nuc Team

package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

const (
	BlockVersion = 1
)

func CheckNUCVersion(version uint32) bool {
	if version == BlockVersion {
		return true
	}
	return false
}

// get NUCDifficulty by tx count
// currentCount > 0 is block verify
// currentCount <= 0 is mining new block
func GetNUCDifficultyByTxCount(node_diff big.Int, chain ChainReader, headerHash common.Hash, number uint64, minerAddr common.Address, minerTxCount uint64) (nucdiff *big.Int, minerRecentTxCount uint64) {
	currentDiff := &big.Int{}
	currentDiff.Set(&node_diff)
	//need calculate the block count
	needReduceDiffTxCount := uint64(10)
	minerRecentTxCount = minerTxCount
	if minerRecentTxCount <= 0 {
		//if minerTxCount is header verify
		minerRecentTxCount = GetMinerRecentTxCount(chain, headerHash, number, minerAddr)
	}
	if minerRecentTxCount >= needReduceDiffTxCount {
		diff := currentDiff.Div(currentDiff, big.NewInt(2))
		return diff, minerRecentTxCount
	}
	return currentDiff, minerRecentTxCount
}

// get NUCDifficulty by tx count
// currentCount > 0 is block verify
// currentCount <= 0 is mining new block
func GetMinerRecentTxCount(chain ChainReader, headerHash common.Hash, number uint64, minerAddr common.Address) uint64 {
	//need calculate the block count
	needReduceDiffTxCount := uint64(10)
	needCalcBlocksCount := 5
	minerRecentTxCount := uint64(0)
	i := 0
	for {
		if minerRecentTxCount >= needReduceDiffTxCount {
			break
		}
		//if block hash arrived the block count
		if i > needCalcBlocksCount {
			break
		}

		b := chain.GetBlock(headerHash, number)
		if b == nil {
			log.Warn("block not exist!", number, headerHash.String())
			break
		}
		minerRecentTxCount += uint64(types.MinerTxCount(chain.ChainConfig(), minerAddr, b.Transactions()))
		// find parent
		headerHash = b.Header().ParentHash
		number = b.Header().Number.Uint64() - 1
		i++
	}
	return minerRecentTxCount
}

// get NUCDifficulty by miner money
// if balance more than 1000 NUC
// nucdifficulty = difficulty / 2
func GetNUCDifficultyByMinerAccount(node_diff big.Int, minerAddr common.Address, chain ChainReader, parentHash common.Hash, number uint64) *big.Int {
	currentDiff := &big.Int{}
	currentDiff.Set(&node_diff)
	parentBlock := chain.GetBlock(parentHash, number)
	if parentBlock == nil {
		return currentDiff
	}
	//need get the address balance
	stateDb, err := chain.StateAt(parentBlock.Root())
	if err != nil {
		return currentDiff
	}
	needReduceDiffBalance := uint64(1000)
	oneNUC := big.NewInt(1000000000000000000)
	balance := stateDb.GetBalance(minerAddr)
	balance = balance.Div(balance, oneNUC)
	if balance.Uint64() > needReduceDiffBalance {
		diff := currentDiff.Div(currentDiff, big.NewInt(2))
		return diff
	}
	return currentDiff
}
