package ethash

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb/reward"
	"math/big"
	"sort"
)

type MiningUser struct {
	HasBind        bool
	Address        common.Address
	PoolAddress    common.Address
	Weight         int64
	Reward         *big.Int
	MortageBalance *big.Int
	CanPocBalance  *big.Int
	Child          MiningUsers
}

type MiningUsers map[common.Address]*MiningUser

const ZERO_ADDR = "0x0000000000000000000000000000000000000000"

func (this *MiningUser) AddChild(addr common.Address) {
	this.Child.Add(addr)
}

func (this *MiningUsers) Has(addr common.Address) bool {
	if _, ok := (*this)[addr]; ok {
		return true
	}
	return false
}

func (this *MiningUsers) MergeReward(mu MiningUsers) {
	for _, v := range mu {
		if !this.Has(v.Address) {
			(*this)[v.Address] = v
		} else {
			(*this)[v.Address].Reward.Add((*this)[v.Address].Reward, v.Reward)
		}
	}
}

func (this *MiningUsers) AddWeight(addr common.Address) {
	if !this.Has(addr) {
		return
	}
	(*this)[addr].Weight += 1
}

func (this *MiningUsers) AddWeightCount(addr common.Address, count int64) {
	if !this.Has(addr) {
		return
	}
	(*this)[addr].Weight += count
}
func (this *MiningUsers) SetBind(addr common.Address, hasBind bool) {
	if !this.Has(addr) {
		return
	}
	(*this)[addr].HasBind = hasBind
}
func (this *MiningUsers) SetBindAddr(addr common.Address, poolAddr common.Address) {
	if !this.Has(addr) {
		return
	}
	(*this)[addr].HasBind = true
	(*this)[addr].PoolAddress = poolAddr
}
func (this *MiningUsers) SetMortageBalance(addr common.Address, balance *big.Int) {
	if !this.Has(addr) {
		return
	}
	(*this)[addr].MortageBalance = balance
	//设置可以获得POC的最大收益
	(*this)[addr].CanPocBalance = balance.Mul(balance, big.NewInt(120))
	(*this)[addr].CanPocBalance = (*this)[addr].CanPocBalance.Div((*this)[addr].CanPocBalance, big.NewInt(100))
}

func (this *MiningUsers) Add(addr common.Address) {
	if this.Has(addr) {
		return
	}
	if addr.String() == "0x0000000000000000000000000000000000000000" {
		return
	}
	(*this)[addr] = &MiningUser{
		false,
		addr,
		common.HexToAddress(ZERO_ADDR),
		1,
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		MiningUsers{},
	}
}

func (this *MiningUsers) AddUser(user *MiningUser) {
	if this.Has(user.Address) {
		return
	}
	(*this)[user.Address] = user
}

func (this *MiningUsers) Remove(addr common.Address) {
	if !this.Has(addr) {
		return
	}
	delete(*this, addr)
}

func (this *MiningUsers) AllWeight() int64 {
	allCount := int64(0)
	for _, v := range *this {
		allCount += v.Weight
	}
	return allCount
}

func (this *MiningUsers) AllMortage() *big.Int {
	allMortage := big.NewInt(0)
	for _, v := range *this {
		allMortage = allMortage.Add(allMortage, v.MortageBalance)
	}
	return allMortage
}

func (this *MiningUsers) GetRankMiningUsers(rank int) MiningUsers {
	topMiningUsers := MiningUsers{}
	l := make([]MiningUser, 0)
	if len(*this) <= rank {
		return *this
	}
	for _, u := range *this {
		l = append(l, *u)
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].MortageBalance.Cmp(l[j].MortageBalance) >= 0
	})
	for k := 0; k < rank; k++ {
		topMiningUsers.AddUser(&l[k])
	}
	return topMiningUsers
}

func (this *MiningUsers) AllValidWeight() int64 {
	allCount := int64(0)
	for _, v := range *this {
		if v.HasBind {
			allCount += v.Weight
		}
	}
	return allCount
}

func (this *MiningUsers) AllDifferentCount() int64 {
	return int64(len(*this))
}

func NUCReward2(header *types.Header, state *state.StateDB, c consensus.ChainReader) *CoinbaseTxs {
	//***********************poc users
	allPocers := GetAllPocers(header, state, c)
	pocReward := GetRewardByType(PocBlockReward, header, state, c)
	everyPocUserReward := pocReward
	allPocerWeight := allPocers.AllWeight()
	if allPocerWeight > 0 {
		everyPocUserReward = new(big.Int).Div(pocReward, big.NewInt(allPocerWeight))
	}
	for _, user := range allPocers {
		user.Reward = user.Reward.Add(user.Reward, everyPocUserReward)
		user.Reward = user.Reward.Mul(user.Reward, big.NewInt(user.Weight))
		if !user.HasBind {
			//invalid
			user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
		}
		// fmt.Println("============ poc reward============", everyPocUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//***********************pow users
	allPowers := GetAllPowers(header, state, c)
	powReward := GetRewardByType(PowBlockReward, header, state, c)
	everyPowUserReward := powReward
	allPowerWeight := allPocers.AllWeight()
	if allPowerWeight+allPocerWeight > 0 {
		everyPowUserReward = new(big.Int).Div(powReward, big.NewInt(allPowerWeight+allPocerWeight))
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
	//***********************pool users***********************************
	allPoolers := GetAllPoolers(header, state, c)
	poolReward := GetRewardByType(PoolBlockReward, header, state, c)
	everyPoolUserReward := poolReward
	allPoolerWeight := allPoolers.AllWeight()
	if allPowerWeight+allPoolerWeight > 0 {
		everyPoolUserReward = new(big.Int).Div(poolReward, big.NewInt(allPowerWeight+allPoolerWeight))
	}
	for _, user := range allPoolers {
		user.Reward = user.Reward.Add(user.Reward, everyPoolUserReward)
		user.Reward = user.Reward.Mul(user.Reward, big.NewInt(user.Weight))
		// fmt.Println("============ pool reward============", everyPoolUserReward, user.Weight, user.Address.String(), user.Reward)
	}
	//***********************pow mortage reward ***********************************
	allTopPowRewardUsers := GetTopPowReward(header, state, c)
	allTopPowAmount := GetContractValue(common.FromHex("0x682c73e3"), header, c.(*core.BlockChain), state)
	topPowReward := big.NewInt(0).Add(big.NewInt(0), PowReward)
	if allTopPowAmount.Cmp(big.NewInt(0)) > 0 {
		topPowReward = topPowReward.Div(topPowReward, allTopPowAmount)
	}
	for _, user := range allTopPowRewardUsers {
		input := "0x44aa7c3b000000000000000000000000"
		input += user.Address.String()[2:]
		currentMortageCount := GetContractValue(common.FromHex(input), header, c.(*core.BlockChain), state)
		user.Reward = user.Reward.Mul(topPowReward, currentMortageCount)
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
	}
	allTopPowPocRewardUsers := GetTopPowPocReward(header, state, c)
	for _, user := range allTopPowPocRewardUsers {
		allWeight := allTopPowPocRewardUsers.AllWeight()
		user.Reward = user.Reward.Mul(topPowReward, big.NewInt(user.Weight))
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(allWeight))
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
	}

	//***********************pool mortage reward***********************************
	allTopPoolRewardUsers := GetTopPoolReward(header, state, c)
	allTopPoolAmount := GetContractValue(common.FromHex("0xcfa14566"), header, c.(*core.BlockChain), state)
	topPoolReward := big.NewInt(0).Add(big.NewInt(0), PoolReward)
	if allTopPoolAmount.Cmp(big.NewInt(0)) > 0 {
		topPoolReward = topPoolReward.Div(topPoolReward, allTopPoolAmount)
	}

	for _, user := range allTopPoolRewardUsers {
		input := "0x2a100e87000000000000000000000000"
		input += user.Address.String()[2:]
		currentMortageCount := GetContractValue(common.FromHex(input), header, c.(*core.BlockChain), state)
		user.Reward = user.Reward.Mul(topPoolReward, currentMortageCount)
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
	}
	allTopPoolPowRewardUsers := GetTopPoolPowReward(header, state, c)
	for _, user := range allTopPoolPowRewardUsers {
		allWeight := allTopPoolPowRewardUsers.AllWeight()
		user.Reward = user.Reward.Mul(topPoolReward, big.NewInt(user.Weight))
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(allWeight))
		user.Reward = user.Reward.Div(user.Reward, big.NewInt(2))
	}

	//***********************all post mortage reward***********************************
	allPostRewardUsers := GetPostUsers(header, state, c)
	allPostReward := big.NewInt(0).Add(big.NewInt(0), PoSTReward)
	if allPostRewardUsers.AllWeight() > 0 {
		allPostReward = allPostReward.Div(allPostReward, big.NewInt(allPostRewardUsers.AllWeight()))
	}
	for _, user := range allPostRewardUsers {
		user.Reward = user.Reward.Add(user.Reward, big.NewInt(2))
	}

	//***********************all top 5 post mortage reward***********************************
	allTop5PostRewardUsers := GetPostUsers(header, state, c)
	allTop5PostReward := big.NewInt(0).Add(big.NewInt(0), Top5PoSTReward)
	allTop5PostAmount := GetContractValue(common.FromHex("0xcddbd64d0000000000000000000000000000000000000000000000000000000000000005"), header, c.(*core.BlockChain), state)
	if allTop5PostAmount.Cmp(big.NewInt(0)) > 0 {
		allTop5PostReward = allTop5PostReward.Div(allTop5PostReward, allTop5PostAmount)
	}
	for _, user := range allTop5PostRewardUsers {
		input := "0xd4bdc612000000000000000000000000"
		input += user.Address.String()[2:]
		currentMortageBalance := GetContractValue(common.FromHex(input), header, c.(*core.BlockChain), state)
		user.Reward = user.Reward.Mul(allTop5PostReward, currentMortageBalance)
	}

	//***********************all top 20 post mortage reward***********************************
	allTop20PostRewardUsers := GetPostUsers(header, state, c)
	allTop20PostReward := big.NewInt(0).Add(big.NewInt(0), Top20PoSTReward)
	allTop20PostAmount := GetContractValue(common.FromHex("0xcddbd64d0000000000000000000000000000000000000000000000000000000000000014"), header, c.(*core.BlockChain), state)
	if allTop5PostAmount.Cmp(big.NewInt(0)) > 0 {
		allTop20PostReward = allTop20PostReward.Div(allTop20PostReward, allTop20PostAmount)
	}
	for _, user := range allTop20PostRewardUsers {
		input := "0xd4bdc612000000000000000000000000"
		input += user.Address.String()[2:]
		currentMortageBalance := GetContractValue(common.FromHex(input), header, c.(*core.BlockChain), state)
		user.Reward = user.Reward.Mul(Top20PoSTReward, currentMortageBalance)
	}

	//***********************all top 100 post mortage reward***********************************
	allTop100PostRewardUsers := GetPostUsers(header, state, c)
	allTop100PostReward := big.NewInt(0).Add(big.NewInt(0), Top100PoSTReward)
	allTop100PostAmount := GetContractValue(common.FromHex("0xcddbd64d0000000000000000000000000000000000000000000000000000000000000064"), header, c.(*core.BlockChain), state)
	if allTop100PostAmount.Cmp(big.NewInt(0)) > 0 {
		allTop100PostReward = allTop100PostReward.Div(allTop100PostReward, allTop100PostAmount)
	}
	for _, user := range allTop100PostRewardUsers {
		input := "0xd4bdc612000000000000000000000000"
		input += user.Address.String()[2:]
		currentMortageBalance := GetContractValue(common.FromHex(input), header, c.(*core.BlockChain), state)
		user.Reward = user.Reward.Mul(topPoolReward, currentMortageBalance)
	}

	//
	ctxs := MergeCoinbasetxs(allPocers, allPowers, allPoolers, allTopPowRewardUsers, allTopPowPocRewardUsers,
		allTopPoolRewardUsers, allTopPoolPowRewardUsers, allPostRewardUsers, allTop5PostRewardUsers,
		allTop20PostRewardUsers, allTop100PostRewardUsers)
	return ctxs
}

func GetRewardByType(reward *big.Int, header *types.Header, state *state.StateDB, c consensus.ChainReader) *big.Int {
	// ========================== get reward reduce ratio ===========================
	data := common.FromHex("0x76b8dde1")
	ratio := GetReduceRatio(data, header, c.(*core.BlockChain), state)
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

//get all powers
func GetAllPowers(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x93510d62")   //PowUsers
	validInput := common.FromHex("0xf7c4de01") //ValidPowUsers
	return GetUsersByType(allInput, validInput, header, state, c)
}

//get all pocers
func GetAllPocers(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x4d4df113")   //PocUsers
	validInput := common.FromHex("0x6957fbaf") //ValidPocUsers
	return GetUsersByType(allInput, validInput, header, state, c)
}

func GetTopPowReward(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x13c7bbce") //GetPowRank
	return GetMiningUsersTop(allInput, header, state, c)
}

func GetTopPowPocReward(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x2b8c9583") //GetPowRankPoc
	return GetMiningUsers(allInput, header, state, c)
}

func GetTopPoolReward(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0xd6badb3f") //GetPoolRank
	return GetMiningUsersTop(allInput, header, state, c)
}

func GetPostUsers(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0xfa3f5e26") //PostUsers
	return GetMiningUsers(allInput, header, state, c)
}

func GetTopPoolPowReward(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x02896c3a") //GetPoolRankPow
	return GetMiningUsers(allInput, header, state, c)
}

//get all poolers
func GetAllPoolers(header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allInput := common.FromHex("0x3dd544e3") //PoolUsers
	allUsers := GetMiningUsers(allInput, header, state, c)
	return allUsers
}

//get all valid pocers
func GetUsersByType(allInput, validInput []byte, header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	allUsers := GetMiningUsers(allInput, header, state, c)
	validUsers := GetMiningUsers(validInput, header, state, c)
	for _, user := range allUsers {
		if validUsers.Has(user.Address) {
			allUsers.SetBind(user.Address, true)
		}
	}
	return allUsers
}

func GetMiningUsers(contractType []byte, header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	users := MiningUsers{}
	addrs := CallContract(contractType, header, c.(*core.BlockChain), state)
	for _, power := range addrs {
		if power.String() == "0x0000000000000000000000000000000000000000" {
			continue
		}
		if !users.Has(power) {
			users.Add(power)
		} else {
			users.AddWeight(power)
		}
	}
	return users
}

func GetMiningUsersTop(contractType []byte, header *types.Header, state *state.StateDB, c consensus.ChainReader) MiningUsers {
	users := MiningUsers{}
	addrs := CallContractArr(contractType, header, c.(*core.BlockChain), state)
	for _, power := range addrs {
		if power.String() == "0x0000000000000000000000000000000000000000" {
			continue
		}
		if !users.Has(power) {
			users.Add(power)
		} else {
			users.AddWeight(power)
		}
	}
	return users
}

func GetContractValue(input []byte, header *types.Header, c *core.BlockChain, state *state.StateDB) *big.Int {
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
				return big.NewInt(0)
			}
			r := bitutil.NewUint256FromString(hex.EncodeToString(res))
			if r.BigInt().Uint64() <= 0 {
				return big.NewInt(0)
			}
			return r.BigInt()
		}
	}
	return big.NewInt(0)
}

type CoinbaseUserReward struct {
	PoolReward *big.Int
	PocReward  *big.Int
	PowReward  *big.Int
	PostReward *big.Int
}

type CoinbaseTxs map[common.Address]*CoinbaseUserReward

func (this *CoinbaseTxs) Has(addr common.Address) bool {
	if _, ok := (*this)[addr]; ok {
		return true
	}
	return false
}
func (this *CoinbaseTxs) Add(addr common.Address) {
	if this.Has(addr) {
		return
	}
	(*this)[addr] = &CoinbaseUserReward{
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
	}
	return
}
func (this *CoinbaseTxs) Encode() []byte {
	b := make([]byte, 0)
	for addr, v := range *this {
		// if v.PoolReward.Cmp(big.NewInt(0)) > 0 {
		// 	// fmt.Println("===================encode v.PoolReward, v.PowReward",
		// 	// 	addr.String(), v.PoolReward, v.PowReward)
		// }
		// fmt.Println("===================encode v.PoolReward, v.PowReward",
		// 	addr.String(), v.PoolReward, v.PowReward)
		b = append(b, addr.Bytes()...)                    //20 bytes
		b = append(b, FormatUint64Bytes(v.PocReward)...)  // 8bytes
		b = append(b, FormatUint64Bytes(v.PowReward)...)  // 8bytes
		b = append(b, FormatUint64Bytes(v.PoolReward)...) // 8bytes
	}
	return b
}
func (this *CoinbaseTxs) Save(number *big.Int) error {
	rewardRecords := this.Encode()
	db := reward.NewRecordDB()
	// fmt.Println("rewardRecords", hex.EncodeToString(rewardRecords), number.Uint64())
	return db.Put(number.Bytes(), rewardRecords)
}
func DecodeFromBytes(coinbaseTxs string) *CoinbaseTxs {
	b := common.FromHex(coinbaseTxs)
	coTxs := &CoinbaseTxs{}
	//every user has 52 bytes
	for i := 0; i < len(b); {
		addr := common.BytesToAddress(b[i : i+20])
		coTxs.Add(addr)
		i += 20
		(*coTxs)[addr].PocReward.SetBytes(b[i : i+8])
		i += 8
		(*coTxs)[addr].PowReward.SetBytes(b[i : i+8])
		i += 8
		(*coTxs)[addr].PoolReward.SetBytes(b[i : i+8])
		i += 8
	}
	return coTxs
}

func MergeCoinbasetxs(pocUsers, powUsers, poolUsers, powUsers1, pocUsers1, powUsers2, poolUsers1, allPostRewardUsers, allTop5PostRewardUsers, allTop20PostRewardUsers, allTop100PostRewardUsers MiningUsers) *CoinbaseTxs {
	pocUsers.MergeReward(pocUsers1)
	coinbaseTxs := &CoinbaseTxs{}
	for _, user := range pocUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PocReward.Add((*coinbaseTxs)[user.Address].PocReward, user.Reward)
	}
	powUsers.MergeReward(powUsers1)
	powUsers.MergeReward(powUsers2)
	for _, user := range powUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PowReward.Add((*coinbaseTxs)[user.Address].PowReward, user.Reward)
	}
	poolUsers.MergeReward(poolUsers1)
	for _, user := range poolUsers {
		if !coinbaseTxs.Has(user.Address) {
			coinbaseTxs.Add(user.Address)
		}
		(*coinbaseTxs)[user.Address].PoolReward.Add((*coinbaseTxs)[user.Address].PoolReward, user.Reward)
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

func FormatUint64Bytes(reward *big.Int) []byte {
	b := make([]byte, 0)
	l := len(reward.Bytes())
	if l < 8 {
		for i := 0; i < 8-l; i++ {
			b = append(b, []byte{0}...)
		}
	}
	b = append(b, reward.Bytes()...)
	return b[:8]
}

func FormatUint64Bytes2(reward *big.Int) []byte {
	b := make([]byte, 0)
	l := len(reward.Bytes())
	if l != 8 {
		for i := 0; i < 8-l; i++ {
			b = append(b, []byte{0}...)
		}
	}
	b = append(b, reward.Bytes()...)
	return b
}
