package ethash

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	math1 "math"
	"math/big"
	"time"
)

func NUCReward3(header *types.Header, state *state.StateDB, c consensus.ChainReader) *CoinbaseTxs {
	//***********************poc users reward*****************************************************//
	allPocers := GetAllPocers1(header, state, c)
	pocReward := GetRewardByType1(PocBlockReward, header, state, c)
	everyPocUserReward := pocReward
	allPocerWeight := allPocers.AllWeight()
	if allPocerWeight > 0 {
		everyPocUserReward = new(big.Int).Div(pocReward, big.NewInt(allPocerWeight))
	}
	for _, user := range allPocers {
		user.Reward = user.Reward.Add(user.Reward, everyPocUserReward)
		// fmt.Println("==========everyPocUserReward", everyPocUserReward, user.Weight)
		if !user.HasBind {
			//invalid
			user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
		}
		// fmt.Println("============ poc reward============", everyPocUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//***********************pow users reward*****************************************************//
	allPowers := GetAllPowers1(allPocers, header, state, c)
	powReward := GetRewardByType1(PowBlockReward, header, state, c)
	everyPowUserReward := powReward
	//所有的 pocer和 power 之和
	allPowerWeight := allPocers.AllWeight() + allPowers.AllWeight()
	if allPowerWeight > 0 {
		everyPowUserReward = new(big.Int).Div(powReward, big.NewInt(allPowerWeight))
	}
	for _, user := range allPowers {
		user.Reward = user.Reward.Add(user.Reward, everyPowUserReward)
		user.Reward = user.Reward.Mul(user.Reward, big.NewInt(user.Weight))
		if !user.HasBind {
			//invalid
			user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
		}
		// fmt.Println("============ pow reward============", everyPowUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//***********************pool users reward **************************************************//
	allPoolers := GetAllPoolers1(header, state, c)
	poolReward := GetRewardByType1(PoolBlockReward, header, state, c)
	everyPoolUserReward := poolReward
	//全网所有的pow 和pool
	allPoolerWeight := allPoolers.AllDifferentCount() + allPowers.AllDifferentCount()
	if allPoolerWeight > 0 {
		everyPoolUserReward = new(big.Int).Div(poolReward, big.NewInt(allPoolerWeight))
	}
	for _, user := range allPoolers {
		user.Reward = user.Reward.Add(user.Reward, everyPoolUserReward)
		user.Reward = user.Reward.Mul(user.Reward, big.NewInt(user.Weight))
	}

	//***********************pow mortage reward **********************************************//
	allTopPowRewardUsers := allPowers.GetRankMiningUsers(20) //前20名 质押power
	allTopPowAmount := allTopPowRewardUsers.AllMortage()     //前20名总共抵押额度
	topPowReward := big.NewInt(0).Add(big.NewInt(0), PowReward)
	if allTopPowAmount.Cmp(big.NewInt(0)) > 0 {
		topPowReward = topPowReward.Div(topPowReward, allTopPowAmount)
	}
	for _, user := range allTopPowRewardUsers {
		topReward := big.NewInt(1).Mul(big.NewInt(1), topPowReward)
		user.Reward = topReward.Mul(topPowReward, user.MortageBalance)
		topReward = topReward.Div(topReward, big.NewInt(2))
		user.Reward = user.Reward.Add(user.Reward, topReward)
		//reward poc users
		if len(user.Child) > 0 {
			everyPowPocerReward := big.NewInt(0).Div(user.Reward, big.NewInt(int64(len(user.Child))))
			for _, pocUser := range user.Child {
				allPocers[pocUser.Address].Reward = allPocers[pocUser.Address].Reward.
					Add(allPocers[pocUser.Address].Reward, everyPowPocerReward)
			}
		}

	}

	//***********************pool mortage reward*********************************************//
	allTopPoolRewardUsers := allPoolers.GetRankMiningUsers(5)
	allTopPoolAmount := allTopPoolRewardUsers.AllMortage()
	topPoolReward := big.NewInt(0).Add(big.NewInt(0), PoolReward)
	if allTopPoolAmount.Cmp(big.NewInt(0)) > 0 {
		topPoolReward = topPoolReward.Div(topPoolReward, allTopPoolAmount)
	}

	for _, user := range allTopPoolRewardUsers {
		topReward := big.NewInt(1).Mul(big.NewInt(1), user.MortageBalance)
		topReward = topReward.Div(topReward, big.NewInt(2))
		user.Reward = user.Reward.Add(user.Reward, topReward)
		//reward poc users
		if len(user.Child) > 0 {
			everyPoolPowerReward := big.NewInt(0).Div(user.Reward, big.NewInt(int64(len(user.Child))))
			for _, powUser := range user.Child {
				allPowers[powUser.Address].Reward = allPowers[powUser.Address].Reward.
					Add(allPowers[powUser.Address].Reward, everyPoolPowerReward)
			}
		}
	}

	//***********************all post mortage reward***********************************
	allPostRewardUsers := GetAllPosters1(header, state, c)
	everyPostReward := big.NewInt(0).Add(big.NewInt(0), PoSTReward)
	if allPostRewardUsers.AllDifferentCount() > 0 {
		everyPostReward = everyPostReward.Div(everyPostReward, big.NewInt(allPostRewardUsers.AllDifferentCount()))
	}
	for _, user := range allPostRewardUsers {
		user.Reward = user.Reward.Add(user.Reward, everyPostReward)
	}

	//***********************all top 5 post mortage reward***********************************
	allTop5PostRewardUsers := allPostRewardUsers.GetRankMiningUsers(5)
	everyTop5PostReward := big.NewInt(0).Add(big.NewInt(0), Top5PoSTReward)
	allTop5PostAmount := allTop5PostRewardUsers.AllMortage()
	for _, user := range allTop5PostRewardUsers {
		topReward := big.NewInt(0).Add(big.NewInt(0), everyTop5PostReward)
		topReward = topReward.Mul(topReward, user.MortageBalance)
		topReward = topReward.Div(topReward, allTop5PostAmount)
		user.Reward = user.Reward.Add(user.Reward, topReward)
	}

	//***********************all top 20 post mortage reward***********************************
	allTop20PostRewardUsers := allPostRewardUsers.GetRankMiningUsers(20)
	everyTop20PostReward := big.NewInt(0).Add(big.NewInt(0), Top20PoSTReward)
	allTop20PostAmount := allTop20PostRewardUsers.AllMortage()
	for _, user := range allTop20PostRewardUsers {
		topReward := big.NewInt(0).Add(big.NewInt(0), everyTop20PostReward)
		topReward = topReward.Mul(topReward, user.MortageBalance)
		topReward = topReward.Div(topReward, allTop20PostAmount)
		user.Reward = user.Reward.Add(user.Reward, topReward)
	}

	//***********************all top 100 post mortage reward***********************************
	allTop100PostRewardUsers := allPostRewardUsers.GetRankMiningUsers(100)
	everyTop100PostReward := big.NewInt(0).Add(big.NewInt(0), Top100PoSTReward)
	allTop100PostAmount := allTop100PostRewardUsers.AllMortage()
	for _, user := range allTop100PostRewardUsers {
		topReward := big.NewInt(0).Add(big.NewInt(0), everyTop100PostReward)
		topReward = topReward.Mul(topReward, user.MortageBalance)
		topReward = topReward.Div(topReward, allTop100PostAmount)
		user.Reward = user.Reward.Add(user.Reward, topReward)
	}
	//
	ctxs := MergeCoinbasetxs1(allPocers, allPowers, allPoolers, allTopPowRewardUsers,
		allTopPoolRewardUsers, allPostRewardUsers, allTop5PostRewardUsers,
		allTop20PostRewardUsers, allTop100PostRewardUsers)
	return ctxs
}

func MergeCoinbasetxs1(pocUsers, powUsers, poolUsers, powUsers1, poolUsers1,
	allPostRewardUsers, allTop5PostRewardUsers, allTop20PostRewardUsers,
	allTop100PostRewardUsers MiningUsers) *CoinbaseTxs {
	coinbaseTxs := &CoinbaseTxs{}
	for _, user := range pocUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PocReward.Add((*coinbaseTxs)[user.Address].PocReward, user.Reward)
	}
	powUsers.MergeReward(powUsers1)
	for _, user := range powUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PowReward.Add((*coinbaseTxs)[user.Address].PowReward, user.Reward)
	}
	// for _, pool := range poolUsers {
	// 	fmt.Println("===========pool", pool.Address.String(), pool.Reward)
	// }
	poolUsers.MergeReward(poolUsers1)
	for _, user := range poolUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PoolReward.Add((*coinbaseTxs)[user.Address].PoolReward, user.Reward)
		fmt.Println("poolUsers length", user.Address.String(), len(poolUsers), (*coinbaseTxs)[user.Address].PoolReward)
	}
	allPostRewardUsers.MergeReward(allTop5PostRewardUsers)
	allPostRewardUsers.MergeReward(allTop20PostRewardUsers)
	allPostRewardUsers.MergeReward(allTop100PostRewardUsers)
	for _, user := range allPostRewardUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PostReward.Add((*coinbaseTxs)[user.Address].PostReward, user.Reward)
	}
	return coinbaseTxs
}

//get all powers
func GetAllPowers1(pocUsers MiningUsers, header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := NewNUCCaller(header, c.(*core.BlockChain), state)
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
			if u.BindPoolAddr.String() != ZERO_ADDR {
				l.SetBind(u.UserAddr, true)
			}
			l.SetMortageBalance(u.UserAddr, u.MortageBalance)
			l.AddWeight(u.UserAddr) // 1+ poc count
			for _, poc := range u.PocAddrs {
				if pocUsers.Has(poc) {
					l.AddWeightCount(u.UserAddr, pocUsers[poc].Weight)
				}
			}
			for _, c := range u.PocAddrs {
				l[u.UserAddr].AddChild(c)
			}
		}
	}
	return l
}

//get all powers
func GetAllPocers1(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := NewNUCCaller(header, c.(*core.BlockChain), state)
	allCount, err := caller.PocerCount()
	l := MiningUsers{}
	if err != nil {
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
			l.Add(u.UserAddr)
			if u.BindPowAddr.String() != ZERO_ADDR {
				l.SetBind(u.UserAddr, true)
			}
			l.SetMortageBalance(u.UserAddr, u.MortageBalance)
			for _, r := range u.Records {
				expireTime := time.Unix(r.ExpireTime.Int64(), 0)
				if expireTime.Before(time.Now()) {
					l.AddWeight(u.UserAddr)
				}
			}
		}
	}
	return l
}

//get all powers
func GetAllPoolers1(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := NewNUCCaller(header, c.(*core.BlockChain), state)
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
			l.SetMortageBalance(u.UserAddr, u.MortageBalance)
			l.AddWeight(u.UserAddr) // 1 + allBindPow count
			if len(u.PowAddrs) > 0 {
				l.SetBind(u.UserAddr, true)
				l.AddWeightCount(u.UserAddr, int64(len(u.PowAddrs)))
			}
			for _, c := range u.PowAddrs {
				l[u.UserAddr].AddChild(c)
			}
		}
	}
	return l
}

//get all powers
func GetAllPosters1(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	caller := NewNUCCaller(header, c.(*core.BlockChain), state)
	allCount, err := caller.PosterCount()
	l := MiningUsers{}
	if err != nil {
		return l
	}
	pageSize := 1000
	pageCount := math1.Ceil(float64(allCount.Uint64()) / float64(pageSize))
	for page := 0; page < int(pageCount); page++ {
		users, err := caller.AllPosters(big.NewInt(int64(page*pageSize)), big.NewInt(int64(pageSize)))
		if err != nil {
			return l
		}
		for _, u := range users {
			l.Add(u.UserAddr)
			l.SetMortageBalance(u.UserAddr, u.MortageBalance)
		}
	}
	return l
}

func GetRewardByType1(reward *big.Int, header *types.Header, state *state.StateDB, c consensus.ChainReader) *big.Int {
	// ========================== get reward reduce ratio ===========================
	caller := NewNUCCaller(header, c.(*core.BlockChain), state)
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
