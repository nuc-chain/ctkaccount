package ethash

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	v1 "github.com/ethereum/go-ethereum/consensus/ethash/nuc_token/v1"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	math1 "math"
	"math/big"
)

func NUCReward4(header *types.Header, state *state.StateDB, c consensus.ChainReader) *CoinbaseTxs {
	//******** pool users//
	allPoolers := GetAllPoolers3(header, state, c)
	//***********************poc users reward logic*****************************************************//
	allPocers := GetAllPocers3(header, state, c)
	pocReward := GetRewardByType3(PocBlockReward, header, state, c)
	everyPocUserReward := big.NewInt(0).Add(big.NewInt(0), pocReward)
	allPocerWeight := allPocers.AllWeight()
	if allPocerWeight > 0 {
		everyPocUserReward = new(big.Int).Div(pocReward, big.NewInt(allPocerWeight))
	}
	for _, user := range allPocers {
		pocR := big.NewInt(0).Add(big.NewInt(0), everyPocUserReward)
		pocR = pocR.Mul(pocR, big.NewInt(user.Weight))
		// fmt.Println("==========everyPocUserReward", everyPocUserReward, user.Weight)
		if !user.HasBind {
			//未绑定获取 50%
			pocR = pocR.Div(pocR, big.NewInt(2))
		} else {
			poolR := new(big.Int).Mul(pocR, big.NewInt(10))
			poolR = poolR.Div(poolR, big.NewInt(100))
			pocR = pocR.Sub(pocR, poolR)
			//绑定矿池节点
			poolUser, ok := allPoolers[user.PoolAddress]
			if ok {
				poolUser.Reward = poolUser.Reward.Add(poolUser.Reward, poolR)
			}
		}

		user.Reward = user.Reward.Add(user.Reward, pocR)
		// fmt.Println("============ poc reward============", everyPocUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//***********************pow users reward*****************************************************//
	allPowers := GetAllPowers3(header, state, c)
	powReward := GetRewardByType3(PowBlockReward, header, state, c)
	everyPowUserReward := big.NewInt(0).Add(big.NewInt(0), powReward)
	//所有的 power 之和
	allPowerWeight := allPowers.AllWeight()
	if allPowerWeight > 0 {
		everyPowUserReward = new(big.Int).Div(powReward, big.NewInt(allPowerWeight))
	}
	for _, user := range allPowers {
		powR := big.NewInt(0).Add(big.NewInt(0), everyPowUserReward)
		powR = powR.Mul(powR, big.NewInt(user.Weight))
		if !user.HasBind {
			//未绑定获取 50%
			powR = powR.Div(powR, big.NewInt(2))
		} else {
			poolR := new(big.Int).Mul(powR, big.NewInt(10))
			poolR = poolR.Div(poolR, big.NewInt(100))
			powR = powR.Sub(powR, poolR)
			//绑定矿池节点
			poolUser, ok := allPoolers[user.PoolAddress]
			if ok {
				poolUser.Reward = poolUser.Reward.Add(poolUser.Reward, poolR)
			}
		}
		user.Reward = user.Reward.Add(user.Reward, powR)
		// fmt.Println("============ pow reward============", everyPowUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//
	ctxs := MergeCoinbasetxs3(allPocers, allPowers, allPoolers)
	return ctxs
}

func MergeCoinbasetxs3(pocUsers, powUsers, poolUsers MiningUsers) *CoinbaseTxs {
	coinbaseTxs := &CoinbaseTxs{}
	for _, user := range pocUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PocReward.Add((*coinbaseTxs)[user.Address].PocReward, user.Reward)
	}
	for _, user := range powUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PowReward.Add((*coinbaseTxs)[user.Address].PowReward, user.Reward)
	}
	for _, user := range poolUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PoolReward.Add((*coinbaseTxs)[user.Address].PoolReward, user.Reward)
		fmt.Println("poolUsers length", user.Address.String(), len(poolUsers), (*coinbaseTxs)[user.Address].PoolReward)
	}
	return coinbaseTxs
}

//get all pocers
func GetAllPocers3(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := v1.NewNUCCaller(header, c.(*core.BlockChain), state)
	allCount, err := caller.PocerCount()
	l := MiningUsers{}
	if err != nil {
		fmt.Println("PocCount", err)
		return l
	}
	pageSize := 1000
	pageCount := math1.Ceil(float64(allCount.Uint64()) / float64(pageSize))
	for page := 0; page < int(pageCount); page++ {
		users, err := caller.AllPocers(big.NewInt(int64(page*pageSize)), big.NewInt(int64(pageSize)))
		if err != nil {
			return l
		}
		for _, u := range users {
			maxReward := u.CanGetMaxReward()
			hasGetReward := state.GetAllPocBalance(u.UserAddr)

			if maxReward.Cmp(hasGetReward) <= 0 { //如果收益达到120% 则自动收益停止
				continue
			}
			l.Add(u.UserAddr)
			if u.BindPoolAddr.String() != ZERO_ADDR {
				l.SetBindAddr(u.UserAddr, u.BindPoolAddr)
			}
			l.SetMortageBalance(u.UserAddr, u.MortageBalance)
			l.AddWeightCount(u.UserAddr, int64(len(u.Records))-1)
		}
	}
	return l
}

//get all powers
func GetAllPowers3(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := v1.NewNUCCaller(header, c.(*core.BlockChain), state)
	allCount, err := caller.PowerCount()
	l := MiningUsers{}
	if err != nil {
		return l
	}
	pageSize := 1000
	pageCount := math1.Ceil(float64(allCount.Uint64()) / float64(pageSize))
	for page := 0; page < int(pageCount); page++ {
		users, err := caller.AllPowers(big.NewInt(int64(page*pageSize)), big.NewInt(int64(pageSize)))
		if err != nil {
			return l
		}
		for _, u := range users {
			l.Add(u.UserAddr)
			l.AddWeightCount(u.UserAddr, int64(len(u.Records))-1)
			if u.BindPoolAddr.String() != ZERO_ADDR {
				l.SetBindAddr(u.UserAddr, u.BindPoolAddr)
			}
		}
	}
	return l
}

//get all poolers
func GetAllPoolers3(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := v1.NewNUCCaller(header, c.(*core.BlockChain), state)
	allCount, err := caller.PoolerCount()
	l := MiningUsers{}
	if err != nil {
		return l
	}
	pageSize := 1000
	pageCount := math1.Ceil(float64(allCount.Uint64()) / float64(pageSize))
	for page := 0; page < int(pageCount); page++ {
		users, err := caller.AllPoolers(big.NewInt(int64(page*pageSize)), big.NewInt(int64(pageSize)))
		if err != nil {
			return l
		}
		for _, u := range users {
			l.Add(u.UserAddr)
		}
	}
	return l
}

func GetRewardByType3(reward *big.Int, header *types.Header, state *state.StateDB, c consensus.ChainReader) *big.Int {
	// ========================== get reward reduce ratio ===========================
	caller := v1.NewNUCCaller(header, c.(*core.BlockChain), state)
	ratioBigData, err := caller.GetRewardRatio()
	if err != nil {
		return big.NewInt(0)
	}
	ratio := ratioBigData.Uint64()
	if ratio == 25 && header.Number.Cmp(big.NewInt(10000)) <= 0 {
		ratio = 0
	}
	ratioBig := math.BigPow(2, int64(ratio))
	newReward := big.NewInt(0).Add(big.NewInt(0), reward)
	if ratio > 1 { // every 2 years reduce half
		newReward = newReward.Div(reward, ratioBig)
	}
	return newReward
}
