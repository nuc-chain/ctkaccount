package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"math"
	"math/big"
	"strings"
)

type NUCCaller struct {
	Abi    abi.ABI
	header *types.Header
	c      *core.BlockChain
	state  *state.StateDB
}

var NucRuleContractAddr = common.HexToAddress("0000000000000000000000000000000000000011")
var CallContractGuessGas = uint64(math.MaxUint64 / 2)

// Struct1 is an auto generated low-level Go binding around an user-defined struct.
type Struct1 struct {
	Records        []Struct0
	UserAddr       common.Address
	GetReward      *big.Int
	Index          *big.Int
	MortageBalance *big.Int
	BindPoolAddr   common.Address
}

func (this *Struct1) CanGetMaxReward() *big.Int {
	reward := this.MortageBalance.Mul(this.MortageBalance, big.NewInt(120))
	reward = reward.Div(reward, big.NewInt(100))
	return reward
}

// Struct0 is an auto generated low-level Go binding around an user-defined struct.
type Struct0 struct {
	CreateTime *big.Int
}

// Struct2 is an auto generated low-level Go binding around an user-defined struct.
type Struct2 struct {
	CreateTime   *big.Int
	Index        *big.Int
	BuyBalance   *big.Int
	BindPoolAddr common.Address
	PocAddrs     []common.Address
	UserAddr     common.Address
	Records      []Struct0
}

// Struct3 is an auto generated low-level Go binding around an user-defined struct.
type Struct3 struct {
	CreateTime *big.Int
	Index      *big.Int
	BuyBalance *big.Int
	PowAddrs   []common.Address
	PocAddrs   []common.Address
	UserAddr   common.Address
}

// TokenABI is the input ABI used to generate the binding from.
const TokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"GetPoCTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"times\",\"type\":\"uint256\"}],\"name\":\"InitPow\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPocers\",\"outputs\":[{\"components\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"getReward\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPowers\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"buyBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"poolUserAddr\",\"type\":\"address\"}],\"name\":\"PocBindPoolUser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"powAddresses\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"poolUserAddr\",\"type\":\"address\"}],\"name\":\"PowBindPoolUser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"frozenBalancers\",\"outputs\":[{\"name\":\"Start\",\"type\":\"uint256\"},{\"name\":\"End\",\"type\":\"uint256\"},{\"name\":\"Addr\",\"type\":\"address\"},{\"name\":\"Balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPooler\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"buyBalance\",\"type\":\"uint256\"},{\"name\":\"powAddrs\",\"type\":\"address[]\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"burnAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPocer\",\"outputs\":[{\"components\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"getReward\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pocTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"BuyPool\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PocExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner1\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetRewardRatio\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"poolTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"BuyPow\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"powTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPoolers\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"buyBalance\",\"type\":\"uint256\"},{\"name\":\"powAddrs\",\"type\":\"address[]\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetCurrentTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"GetLeftFrozenBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPowTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PowCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"BuyPoc\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reduceDuration\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PoolCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPower\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"buyBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"start\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PoolExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pocExpired\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"InitContract\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PowExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPoolTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PocCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"times\",\"type\":\"uint256\"}],\"name\":\"InitPoc\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"ChangeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"_start\",\"type\":\"uint256\"},{\"name\":\"_end\",\"type\":\"uint256\"},{\"name\":\"balance\",\"type\":\"uint256\"}],\"name\":\"SetFrozenBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"ChangeOwner1\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

func NewNUCCaller(header *types.Header, c *core.BlockChain, state *state.StateDB) *NUCCaller {
	parsed, err := abi.JSON(strings.NewReader(TokenABI))
	if err != nil {
		return nil
	}
	return &NUCCaller{
		Abi:    parsed,
		header: header,
		c:      c,
		state:  state,
	}
}

func (this *NUCCaller) AllPowers(offset, pageSize *big.Int) ([]Struct2, error) {
	method := "AllPowers"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		fmt.Println("+++++++==", err)
		return nil, err
	}
	out := new([]Struct2)
	// fmt.Println("===========input", hex.EncodeToString(input))
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, nil
}

func (this *NUCCaller) PowerCount() (*big.Int, error) {
	method := "PowCount"
	input, err := this.Abi.Pack(method)
	if err != nil {
		return nil, err
	}
	out := new(*big.Int)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, err
}

func (this *NUCCaller) AllPocers(offset, pageSize *big.Int) ([]Struct1, error) {
	method := "AllPocers"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := new([]Struct1)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, nil
}

func (this *NUCCaller) PocerCount() (*big.Int, error) {
	method := "PocCount"
	input, err := this.Abi.Pack(method)
	if err != nil {
		return nil, err
	}
	out := new(*big.Int)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	if len(output) <= 0 {
		return big.NewInt(0), nil
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, err
}

func (this *NUCCaller) AllPoolers(offset, pageSize *big.Int) ([]Struct3, error) {
	method := "AllPoolers"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := new([]Struct3)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, nil
}

func (this *NUCCaller) PoolerCount() (*big.Int, error) {
	method := "PoolCount"
	input, err := this.Abi.Pack(method)
	if err != nil {
		return nil, err
	}
	out := new(*big.Int)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, err
}

func (this *NUCCaller) GetRewardRatio() (*big.Int, error) {
	method := "GetRewardRatio"
	input, err := this.Abi.Pack(method)
	if err != nil {
		return nil, err
	}
	out := new(*big.Int)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, err
}

func (this *NUCCaller) CallContract(input []byte) (res []byte, err error) {
	failed := false
	msg := types.NewMessage(NucRuleContractAddr, &NucRuleContractAddr, 0, big.NewInt(0), CallContractGuessGas, big.NewInt(0), input, false)
	context := core.NewEVMContext(msg, this.header, this.c, nil)
	evm := vm.NewEVM(context, this.state, this.c.Config(), *this.c.GetVMConfig())
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	if evm != nil {
		res, _, failed, err = core.ApplyMessage(evm, msg, gp)
		if !failed && err == nil {
			return res, nil
		}
	}
	fmt.Println(hex.EncodeToString(input), "==========coinbase call contract error", failed, err)
	return nil, errors.New(hex.EncodeToString(input) + "==========coinbase call contract error")
}
