// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
	"runtime"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

// Ethash proof-of-work protocol constants.
var (
	DefaultCoinbaseAddr = common.BytesToAddress([]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	})
	FrontierBlockReward       = big.NewInt(1e+18).Mul(big.NewInt(1e+18), big.NewInt(12)) // Block reward in wei for successfully mining a block
	PowBlockReward            = big.NewInt(2e+18)                                        // Block reward for nuc pow 2NUC
	PocBlockReward            = big.NewInt(1e+18).Mul(big.NewInt(1e+18), big.NewInt(10)) // Block reward for 10 nuc poc
	PoolBlockReward           = big.NewInt(1e+18)                                        // Block reward for nuc pool
	PowReward                 = big.NewInt(2e+18)                                        // Block reward for nuc pool
	PoolReward                = big.NewInt(2e+18)                                        // Block reward for nuc pool
	PoSTReward                = big.NewInt(1e+18)                                        // Block reward for nuc pool
	Top5PoSTReward            = big.NewInt(1e+18)                                        // Block reward for nuc pool
	Top20PoSTReward           = big.NewInt(1e+18)                                        // Block reward for nuc pool
	Top100PoSTReward          = big.NewInt(1e+18)                                        // Block reward for nuc pool
	ByzantiumBlockReward      = big.NewInt(3e+18)                                        // Block reward in wei for successfully mining a block upward from Byzantium
	ConstantinopleBlockReward = big.NewInt(2e+18)                                        // Block reward in wei for successfully mining a block upward from Constantinople
	maxUncles                 = 2                                                        // Maximum number of uncles allowed in a single block
	allowedFutureBlockTime    = 60 * time.Second                                         // Max time from current time allowed for blocks, before they're considered future blocks
	NucRuleContractAddr       = common.HexToAddress("0000000000000000000000000000000000000011")
	// calcDifficultyConstantinople is the difficulty adjustment algorithm for Constantinople.
	// It returns the difficulty that a new block should have when created at time given the
	// parent block's time and difficulty. The calculation uses the Byzantium rules, but with
	// bomb offset 5M.
	// Specification EIP-1234: https://eips.ethereum.org/EIPS/eip-1234
	calcDifficultyConstantinople = makeDifficultyCalculator(big.NewInt(5000000))

	// calcDifficultyByzantium is the difficulty adjustment algorithm. It returns
	// the difficulty that a new block should have when created at time given the
	// parent block's time and difficulty. The calculation uses the Byzantium rules.
	// Specification EIP-649: https://eips.ethereum.org/EIPS/eip-649
	calcDifficultyByzantium = makeDifficultyCalculator(big.NewInt(3000000))

	CallContractGuessGas = uint64(math.MaxUint64 / 2)
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errZeroBlockTime     = errors.New("timestamp equals parent's")
	errTooManyUncles     = errors.New("too many uncles")
	errDuplicateUncle    = errors.New("duplicate uncle")
	errUncleIsAncestor   = errors.New("uncle is ancestor")
	errDanglingUncle     = errors.New("uncle's parent is not ancestor")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")
)

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (ethash *Ethash) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
func (ethash *Ethash) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake {
		return nil
	}
	// Short circuit if the header is known, or its parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return ethash.verifyHeader(chain, header, parent, false, seal)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (ethash *Ethash) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake || len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = ethash.verifyHeaderWorker(chain, headers, seals, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

func (ethash *Ethash) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, index int) error {
	var parent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.Uint64()) != nil {
		return nil // known block
	}
	return ethash.verifyHeader(chain, headers[index], parent, false, seals[index])
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of the stock Ethereum ethash engine.
func (ethash *Ethash) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake {
		return nil
	}
	// Verify that there are at most 2 uncles included in this block
	if len(block.Uncles()) > maxUncles {
		return errTooManyUncles
	}
	if len(block.Uncles()) == 0 {
		return nil
	}
	// Gather the set of past uncles and ancestors
	uncles, ancestors := mapset.NewSet(), make(map[common.Hash]*types.Header)

	number, parent := block.NumberU64()-1, block.ParentHash()
	for i := 0; i < 7; i++ {
		ancestor := chain.GetBlock(parent, number)
		if ancestor == nil {
			break
		}
		ancestors[ancestor.Hash()] = ancestor.Header()
		for _, uncle := range ancestor.Uncles() {
			uncles.Add(uncle.Hash())
		}
		parent, number = ancestor.ParentHash(), number-1
	}
	ancestors[block.Hash()] = block.Header()
	uncles.Add(block.Hash())

	// Verify each of the uncles that it's recent, but not an ancestor
	for _, uncle := range block.Uncles() {
		// Make sure every uncle is rewarded only once
		hash := uncle.Hash()
		if uncles.Contains(hash) {
			return errDuplicateUncle
		}
		uncles.Add(hash)

		// Make sure the uncle has a valid ancestry
		if ancestors[hash] != nil {
			return errUncleIsAncestor
		}
		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			return errDanglingUncle
		}
		if err := ethash.verifyHeader(chain, uncle, ancestors[uncle.ParentHash], true, true); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
// See YP section 4.3.4. "Block Header Validity"
func (ethash *Ethash) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, uncle bool, seal bool) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if !uncle {
		if header.Time > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
			return consensus.ErrFutureBlock
		}
	}
	if header.Time <= parent.Time {
		return errZeroBlockTime
	}
	// Verify the block's difficulty based in its timestamp and parent's difficulty
	expected := ethash.CalcDifficulty(chain, header.Time, parent)
	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}

	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := ethash.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyDAOHeaderExtraData(chain.Config(), header); err != nil {
		return err
	}
	if err := misc.VerifyForkHashes(chain.Config(), header, uncle); err != nil {
		return err
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (ethash *Ethash) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return CalcDifficulty(chain.Config(), time, parent)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
	return calcDifficultyFrontier(time, parent)
	// next := new(big.Int).Add(parent.Number, big1)
	// switch {
	// case config.IsConstantinople(next):
	// return calcDifficultyConstantinople(time, parent)
	// case config.IsByzantium(next):
	// return calcDifficultyByzantium(time, parent)
	// case config.IsHomestead(next):
	// 	return calcDifficultyHomestead(time, parent)
	// default:
	// 	return calcDifficultyFrontier(time, parent)
	// }
}

// Some weird constants to avoid constant memory allocs for them.
var (
	expDiffPeriod = big.NewInt(100000)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
	big9          = big.NewInt(9)
	big10         = big.NewInt(10)
	bigMinus99    = big.NewInt(-99)
)

// makeDifficultyCalculator creates a difficultyCalculator with the given bomb-delay.
// the difficulty is calculated with Byzantium rules, which differs from Homestead in
// how uncles affect the calculation
func makeDifficultyCalculator(bombDelay *big.Int) func(time uint64, parent *types.Header) *big.Int {
	// Note, the calculations below looks at the parent number, which is 1 below
	// the block number. Thus we remove one from the delay given
	bombDelayFromParent := new(big.Int).Sub(bombDelay, big1)
	return func(time uint64, parent *types.Header) *big.Int {
		// https://github.com/ethereum/EIPs/issues/100.
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
		//        ) + 2^(periodCount - 2)

		bigTime := new(big.Int).SetUint64(time)
		bigParentTime := new(big.Int).SetUint64(parent.Time)

		// holds intermediate values to make the algo easier to read & audit
		x := new(big.Int)
		y := new(big.Int)

		// (2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9
		x.Sub(bigTime, bigParentTime)
		x.Div(x, big9)
		if parent.UncleHash == types.EmptyUncleHash {
			x.Sub(big1, x)
		} else {
			x.Sub(big2, x)
		}
		// max((2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9, -99)
		if x.Cmp(bigMinus99) < 0 {
			x.Set(bigMinus99)
		}
		// parent_diff + (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
		y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
		x.Mul(y, x)
		x.Add(parent.Difficulty, x)

		// minimum difficulty can ever be (before exponential factor)
		if x.Cmp(params.MinimumDifficulty) < 0 {
			x.Set(params.MinimumDifficulty)
		}
		// calculate a fake block number for the ice-age delay
		// Specification: https://eips.ethereum.org/EIPS/eip-1234
		fakeBlockNumber := new(big.Int)
		if parent.Number.Cmp(bombDelayFromParent) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, bombDelayFromParent)
		}
		// for the exponential factor
		periodCount := fakeBlockNumber
		periodCount.Div(periodCount, expDiffPeriod)

		// the exponential factor, commonly referred to as "the bomb"
		// diff = diff + 2^(periodCount - 2)
		if periodCount.Cmp(big1) > 0 {
			y.Sub(periodCount, big2)
			y.Exp(big2, y, nil)
			x.Add(x, y)
		}
		return x
	}
}

// calcDifficultyHomestead is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the Homestead rules.
func calcDifficultyHomestead(time uint64, parent *types.Header) *big.Int {
	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	//        ) + 2^(periodCount - 2)

	bigTime := new(big.Int).SetUint64(time)
	bigParentTime := new(big.Int).SetUint64(parent.Time)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// 1 - (block_timestamp - parent_timestamp) // 10
	x.Sub(bigTime, bigParentTime)
	x.Div(x, big10)
	x.Sub(big1, x)

	// max(1 - (block_timestamp - parent_timestamp) // 10, -99)
	if x.Cmp(bigMinus99) < 0 {
		x.Set(bigMinus99)
	}
	// (parent_diff + parent_diff // 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(parent.Difficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}
	// for the exponential factor
	periodCount := new(big.Int).Add(parent.Number, big1)
	periodCount.Div(periodCount, expDiffPeriod)

	// the exponential factor, commonly referred to as "the bomb"
	// diff = diff + 2^(periodCount - 2)
	if periodCount.Cmp(big1) > 0 {
		y.Sub(periodCount, big2)
		y.Exp(big2, y, nil)
		x.Add(x, y)
	}
	return x
}

// calcDifficultyFrontier is the difficulty adjustment algorithm. It returns the
// difficulty that a new block should have when created at time given the parent
// block's time and difficulty. The calculation uses the Frontier rules.
func calcDifficultyFrontier(time uint64, parent *types.Header) *big.Int {
	diff := new(big.Int)
	adjust := new(big.Int).Div(parent.Difficulty, params.DifficultyBoundDivisor)
	bigTime := new(big.Int)
	bigParentTime := new(big.Int)

	bigTime.SetUint64(time)
	bigParentTime.SetUint64(parent.Time)
	// offset := bigTime.Sub(bigTime, bigParentTime)
	// fmt.Println("difficulty ajustment ===========================", offset.Uint64(), params.DurationLimit, adjust, parent.Difficulty)
	if bigTime.Sub(bigTime, bigParentTime).Cmp(params.DurationLimit) < 0 {
		diff.Add(parent.Difficulty, adjust)
	} else {
		diff.Sub(parent.Difficulty, adjust)
	}
	if diff.Cmp(params.MinimumDifficulty) < 0 {
		diff.Set(params.MinimumDifficulty)
	}

	periodCount := new(big.Int).Add(parent.Number, big1)
	periodCount.Div(periodCount, expDiffPeriod)
	if periodCount.Cmp(big1) > 0 {
		// diff = diff + 2^(periodCount - 2)
		expDiff := periodCount.Sub(periodCount, big2)
		expDiff.Exp(big2, expDiff, nil)
		diff.Add(diff, expDiff)
		diff = math.BigMax(diff, params.MinimumDifficulty)
	}
	return diff
}

// VerifySeal implements consensus.Engine, checking whether the given block satisfies
// the PoW difficulty requirements.
func (ethash *Ethash) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return ethash.verifySeal(chain, header, false)
}

// verifySeal checks whether a block satisfies the PoW difficulty requirements,
// either using the usual ethash cache for it, or alternatively using a full DAG
// to make remote mining fast.
func (ethash *Ethash) verifySeal(chain consensus.ChainReader, header *types.Header, fulldag bool) error {
	// If we're running a fake PoW, accept any seal as valid
	if ethash.config.PowMode == ModeFake || ethash.config.PowMode == ModeFullFake {
		time.Sleep(ethash.fakeDelay)
		if ethash.fakeFail == header.Number.Uint64() {
			return errInvalidPoW
		}
		return nil
	}
	// If we're running a shared PoW, delegate verification to it
	if ethash.shared != nil {
		return ethash.shared.verifySeal(chain, header, fulldag)
	}
	// Ensure that we have a valid difficulty for the block
	if header.NUCDifficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}
	// Recompute the digest and PoW values
	number := header.Number.Uint64()

	var (
		digest []byte
		result []byte
	)
	// If fast-but-heavy PoW verification was requested, use an ethash dataset
	if fulldag {
		dataset := ethash.dataset(number, true)
		if dataset.generated() {
			digest, result = hashimotoFull(dataset.dataset, ethash.SealHash(header).Bytes(), header.Nonce.Uint64())

			// Datasets are unmapped in a finalizer. Ensure that the dataset stays alive
			// until after the call to hashimotoFull so it's not unmapped while being used.
			runtime.KeepAlive(dataset)
		} else {
			// Dataset not yet generated, don't hang, use a cache instead
			fulldag = false
		}
	}
	// If slow-but-light PoW verification was requested (or DAG not yet ready), use an ethash cache
	if !fulldag {
		cache := ethash.cache(number)

		size := datasetSize(number)
		if ethash.config.PowMode == ModeTest {
			size = 32 * 1024
		}
		digest, result = hashimotoLight(size, cache.cache, ethash.SealHash(header).Bytes(), header.Nonce.Uint64())

		// Caches are unmapped in a finalizer. Ensure that the cache stays alive
		// until after the call to hashimotoLight so it's not unmapped while being used.
		runtime.KeepAlive(cache)
	}
	// Verify the calculated values against the ones provided in the header
	if !bytes.Equal(header.MixDigest[:], digest) {
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(two256, header.NUCDifficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
func (ethash *Ethash) Prepare(chain consensus.ChainReader, header *types.Header) error {
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Difficulty = ethash.CalcDifficulty(chain, header.Time, parent)
	header.NUCDifficulty = header.Difficulty
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state on the header
func (ethash *Ethash) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	// Accumulate any block and uncle rewards and commit the final state root
	accumulateRewards(chain, state, header, uncles, txs)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
}

// FinalizeAndAssemble implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state and assembling the block.
func (ethash *Ethash) FinalizeAndAssemble(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	accumulateRewards(chain, state, header, uncles, txs)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, uncles, receipts), nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (ethash *Ethash) SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	rlp.Encode(hasher, []interface{}{
		header.Version,
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.NUCDifficulty,
		header.CoinbaseTxs,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	})
	hasher.Sum(hash[:0])
	return hash
}

// Some weird constants to avoid constant memory allocs for them.
var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

const (
	NUC_POW  = 0
	NUC_POC  = 1
	NUC_POOL = 2
)

func accumulateRewards(c consensus.ChainReader, state *state.StateDB, header *types.Header,
	uncles []*types.Header, txs []*types.Transaction) {
	blockReward := FrontierBlockReward
	fee := big.NewInt(0)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, big8)
		state.AddBalance(uncle.Coinbase, r)

		r.Div(blockReward, big32)
		fee.Add(fee, r)
	}
	for _, tx := range txs {
		fee.Add(fee, new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas())))
		// 70% fee to pow and pool
		fee = fee.Mul(fee, big.NewInt(70))
		fee = fee.Div(fee, big.NewInt(100))
	}
	teamFee := big.NewInt(0).Mul(fee, big.NewInt(5))
	teamFee = teamFee.Div(teamFee, big.NewInt(100))
	powFee := big.NewInt(0).Mul(fee, big.NewInt(95))
	powFee = powFee.Div(powFee, big.NewInt(100))

	allBlockReward := big.NewInt(0)
	ctxs := NUCReward4(header, state, c)
	if len(*ctxs) > 0 {
		powFee = powFee.Div(powFee, big.NewInt(int64(len(*ctxs))))
	}
	for addr, user := range *ctxs {
		powReward := big.NewInt(0)
		poolReward := big.NewInt(0)
		pocReward := big.NewInt(0)
		if user.PowReward.Cmp(big.NewInt(0)) > 0 {
			//如果是pow
			powReward = big.NewInt(0).Add(big.NewInt(0), user.PowReward)
		}
		if user.PoolReward.Cmp(big.NewInt(0)) > 0 {
			//如果是pool
			poolReward = big.NewInt(0).Add(big.NewInt(0), user.PoolReward)
		}
		if user.PocReward.Cmp(big.NewInt(0)) > 0 {
			//如果是poc 则记录总的pocReward
			pocReward = big.NewInt(0).Add(big.NewInt(0), user.PocReward)
			state.AddAllPocBalance(addr, pocReward)
		}
		//all reward = poc reward + pow reward + pool reward + post reward
		reward := new(big.Int).Add(pocReward, powReward)
		reward = reward.Add(reward, poolReward)
		reward = reward.Add(reward, powFee)
		if reward.Cmp(big.NewInt(0)) > 0 {
			state.AddBalance(addr, reward)
		}
		allBlockReward.Add(allBlockReward, reward)
	}
	if allBlockReward.Cmp(blockReward) < 0 {
		leftReward := big.NewInt(0).Sub(blockReward, allBlockReward)
		state.AddBalance(DefaultCoinbaseAddr, leftReward)
	}
	if teamFee.Cmp(big.NewInt(0)) > 0 {
		state.AddBalance(DefaultCoinbaseAddr, teamFee)
	}
	header.CoinbaseTxs = ctxs.Encode()
	_ = ctxs.Save(header.Number)
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewardsold(c consensus.ChainReader, state *state.StateDB, header *types.Header, uncles []*types.Header, txs []*types.Transaction) {
	header.Coinbase = GetCoinbase(header, c.(*core.BlockChain), state)
	// Select the correct block reward based on chain progression
	blockReward := FrontierBlockReward
	// if c.Config().IsByzantium(header.Number) {
	// 	blockReward = ByzantiumBlockReward
	// }
	// if c.Config().IsConstantinople(header.Number) {
	// 	blockReward = ConstantinopleBlockReward
	// }
	// Accumulate the rewards for the miner and any included uncles
	fee := big.NewInt(0)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, big8)
		state.AddBalance(uncle.Coinbase, r)

		r.Div(blockReward, big32)
		fee.Add(fee, r)
	}
	for _, tx := range txs {
		fee.Add(fee, new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas())))
		// 70% fee to pow and pool
		fee = fee.Mul(fee, big.NewInt(70))
		fee = fee.Div(fee, big.NewInt(100))
	}

	// ========================== get reward reduce ratio ===========================
	data := common.FromHex("0x76b8dde1")
	ratio := GetReduceRatio(data, header, c.(*core.BlockChain), state)
	if ratio == 25 && header.Number.Cmp(big.NewInt(10000)) <= 0 {
		ratio = 0
	}
	coinbaseTxs := make([]byte, 0)
	// ========================== get pow users ===========================
	coinbaseTxs = append(coinbaseTxs, []byte{NUC_POW}...) // 1 byte
	ratioBig := math.BigPow(2, int64(ratio))
	powReward := new(big.Int).Set(PowBlockReward)
	if ratio > 1 { // every 2 years reduce half
		powReward = powReward.Div(powReward, ratioBig)
	}
	powFee := fee
	// pow get 60% tx fee
	powFee = powFee.Mul(powFee, big.NewInt(60))
	powFee = powFee.Div(powFee, big.NewInt(100))
	powReward = powReward.Add(powReward, powFee)
	data = common.FromHex("0xcca1aa47") //pow Count
	powCount := CallPowCount(data, header, c.(*core.BlockChain), state)
	//pow users
	data = common.FromHex("0x93510d62")

	pows := CallContract(data, header, c.(*core.BlockChain), state)
	l := len(pows)

	if powCount > 1 {
		powReward = powReward.Div(powReward, big.NewInt(powCount))
	}
	if powCount == 0 {
		pows = append(pows, DefaultCoinbaseAddr)
	}
	if powCount > int64(l) {
		for j := 0; j < int(powCount-int64(l)); j++ {
			pows = append(pows, DefaultCoinbaseAddr)
		}
	}
	if len(powReward.Bytes()) != 8 {
		for i := 0; i < 8-len(powReward.Bytes()); i++ {
			coinbaseTxs = append(coinbaseTxs, []byte{0}...)
		}
	}
	coinbaseTxs = append(coinbaseTxs, powReward.Bytes()...) // 8 bytes

	b := make([]byte, 4)
	if powCount == 0 {
		powCount = 1
	}
	binary.LittleEndian.PutUint32(b, uint32(powCount))
	coinbaseTxs = append(coinbaseTxs, b...) // 4 bytes
	for _, addr := range pows {
		state.AddBalance(addr, powReward)
		coinbaseTxs = append(coinbaseTxs, addr.Bytes()...)
	}
	// =========================      get poc users ===========================

	coinbaseTxs = append(coinbaseTxs, []byte{NUC_POC}...)
	pocReward := new(big.Int).Set(PocBlockReward)
	if ratio > 1 { // every 2 years reduce half
		pocReward = pocReward.Div(pocReward, ratioBig)
	}

	data = common.FromHex("0x4d4df113")
	CallContract(data, header, c.(*core.BlockChain), state)
	pocs := CallContract(data, header, c.(*core.BlockChain), state)

	l = len(pocs)
	if l > 1 {
		pocReward = pocReward.Div(pocReward, big.NewInt(int64(l)))
	}
	if len(pocReward.Bytes()) != 8 {
		for i := 0; i < 8-len(pocReward.Bytes()); i++ {
			coinbaseTxs = append(coinbaseTxs, []byte{0}...)
		}
	}
	coinbaseTxs = append(coinbaseTxs, pocReward.Bytes()...) // 8 bytes

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(l))
	coinbaseTxs = append(coinbaseTxs, b...) // 4 bytes

	for _, addr := range pocs {
		state.AddBalance(addr, pocReward)
		coinbaseTxs = append(coinbaseTxs, addr.Bytes()...)
	}

	// ==========================        get pool users  ========================
	coinbaseTxs = append(coinbaseTxs, []byte{NUC_POOL}...)
	data = common.FromHex("0x3dd544e3")
	poolReward := new(big.Int).Set(PoolBlockReward)
	if ratio > 1 { // every 2 years reduce half
		poolReward = poolReward.Div(poolReward, ratioBig)
	}

	poolFee := fee
	// pow get 40% tx fee
	poolFee = poolFee.Mul(poolFee, big.NewInt(40))
	poolFee = poolFee.Div(poolFee, big.NewInt(100))
	poolReward = poolReward.Add(poolReward, poolFee)

	pools := CallContract(data, header, c.(*core.BlockChain), state)
	l = len(pools)
	if powCount > 1 {
		poolReward = poolReward.Div(poolReward, big.NewInt(powCount))
	}
	if powCount == 0 {
		pools = append(pools, DefaultCoinbaseAddr)
	}
	if powCount > int64(l) {
		for j := 0; j < int(powCount-int64(l)); j++ {
			pools = append(pools, DefaultCoinbaseAddr)
		}
	}
	if len(poolReward.Bytes()) != 8 {
		for i := 0; i < 8-len(poolReward.Bytes()); i++ {
			coinbaseTxs = append(coinbaseTxs, []byte{0}...)
		}
	}

	coinbaseTxs = append(coinbaseTxs, poolReward.Bytes()...) // 8 bytes
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(powCount))
	coinbaseTxs = append(coinbaseTxs, b...) // 4 bytes
	for _, addr := range pools {
		state.AddBalance(addr, poolReward)
		coinbaseTxs = append(coinbaseTxs, addr.Bytes()...)
	}
	header.CoinbaseTxs = coinbaseTxs

	// state.AddBalance(header.Coinbase, blockReward)
}

func CallContract(input []byte, header *types.Header, c *core.BlockChain, state *state.StateDB) []common.Address {
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), input, false)

	context := core.NewEVMContext(msg, header, c, nil)
	evm := vm.NewEVM(context, state, c.Config(), *c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err := core.ApplyMessage(evm, msg, gp)
		if failed || err != nil {
			fmt.Println(hex.EncodeToString(input), "==========coinbase call contract error", failed, err)
		} else {
			return GetAddressesFromContractHex(hex.EncodeToString(res))
		}
	}
	return []common.Address{}
}

func CallContractArr(input []byte, header *types.Header, c *core.BlockChain, state *state.StateDB) []common.Address {
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), input, false)

	context := core.NewEVMContext(msg, header, c, nil)
	evm := vm.NewEVM(context, state, c.Config(), *c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err := core.ApplyMessage(evm, msg, gp)
		if failed || err != nil {
			fmt.Println(hex.EncodeToString(input), "==========coinbase call contract error", failed, err)
		} else {
			return GetArrayAddressesFromContractHex(hex.EncodeToString(res))
		}
	}
	return []common.Address{}
}

func CallPowCount(input []byte, header *types.Header, c *core.BlockChain, state *state.StateDB) int64 {
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), input, false)

	context := core.NewEVMContext(msg, header, c, nil)
	evm := vm.NewEVM(context, state, c.Config(), *c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err := core.ApplyMessage(evm, msg, gp)
		if failed || err != nil {
			fmt.Println(hex.EncodeToString(input), "==========coinbase call contract error", failed, err)
		} else {
			bigN := new(big.Int)
			bigN.SetBytes(res)
			return int64(bigN.Uint64())
		}
	}
	return 0
}

func GetCoinbase(header *types.Header, c *core.BlockChain, state *state.StateDB) common.Address {
	//coinbase 0x8da5cb5b
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), common.FromHex("0x8da5cb5b"), false)
	context := core.NewEVMContext(msg, header, c, nil)
	evm := vm.NewEVM(context, state, c.Config(), *c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err := core.ApplyMessage(evm, msg, gp)
		if failed || err != nil {
			fmt.Println("coinbase get ==========coinbase call contract error", failed, err)
		} else {
			addr := common.BytesToAddress(res)
			if addr.String() == "0x0000000000000000000000000000000000000000" {
				return DefaultCoinbaseAddr
			}
			return addr
		}
	}
	return DefaultCoinbaseAddr
}

func GetReduceRatio(input []byte, header *types.Header, c *core.BlockChain, state *state.StateDB) uint64 {
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), input, false)

	context := core.NewEVMContext(msg, header, c, nil)
	evm := vm.NewEVM(context, state, c.Config(), *c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err := core.ApplyMessage(evm, msg, gp)
		if failed || err != nil {
			fmt.Println(hex.EncodeToString(input), "==========coinbase call contract error", failed, err)
		} else {
			// fmt.Println(hex.EncodeToString(res), "hex.EncodeToString(res)")
			if res == nil || len(res) < 1 {
				return 1
			}
			r := bitutil.NewUint256FromString(hex.EncodeToString(res))
			// fmt.Println("ratio:", r.BigInt().Uint64())
			if r.BigInt().Uint64() <= 0 {
				return 0
			}
			return r.BigInt().Uint64()
		}
	}
	return 0
}

func GetAddressesFromContractHex(h string) []common.Address {
	if len(h) <= 128 {
		return []common.Address{}
	}
	h = h[128:]
	if len(h)%64 != 0 {
		return []common.Address{}
	}
	res := make([]common.Address, 0)
	for {
		address := h[0:64]
		address = address[24:]
		res = append(res, common.HexToAddress(address))
		if len(h) == 64 {
			break
		}
		h = h[64:]
	}
	return res
}

func GetArrayAddressesFromContractHex(h string) []common.Address {
	if len(h)%64 != 0 {
		return []common.Address{}
	}
	res := make([]common.Address, 0)
	for {
		address := h[0:64]
		address = address[24:]
		res = append(res, common.HexToAddress(address))
		if len(h) == 64 {
			break
		}
		h = h[64:]
	}
	return res
}
