package ethash

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

// Struct3 is an auto generated low-level Go binding around an user-defined struct.
type Struct3 struct {
	Records        []Struct2
	UserAddr       common.Address
	Index          *big.Int
	MortageBalance *big.Int
	BindPowAddr    common.Address
}

// Struct2 is an auto generated low-level Go binding around an user-defined struct.
type Struct2 struct {
	CreateTime *big.Int
	ExpireTime *big.Int
}

// Struct4 is an auto generated low-level Go binding around an user-defined struct.
type Struct4 struct {
	CreateTime     *big.Int
	Index          *big.Int
	MortageBalance *big.Int
	UserAddr       common.Address
}

// Struct1 is an auto generated low-level Go binding around an user-defined struct.
type Struct1 struct {
	CreateTime     *big.Int
	Index          *big.Int
	MortageBalance *big.Int
	BindPoolAddr   common.Address
	PocAddrs       []common.Address
	UserAddr       common.Address
}

// Struct0 is an auto generated low-level Go binding around an user-defined struct.
type Struct0 struct {
	CreateTime     *big.Int
	Index          *big.Int
	MortageBalance *big.Int
	PowAddrs       []common.Address
	UserAddr       common.Address
}

// TokenABI is the input ABI used to generate the binding from.
const TokenABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"n\",\"type\":\"uint256\"}],\"name\":\"BenchPoolTest\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"n\",\"type\":\"uint256\"}],\"name\":\"BenchPostTest\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"n\",\"type\":\"uint256\"}],\"name\":\"BenchPowTest\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"poolUserAddr\",\"type\":\"address\"}],\"name\":\"BindPoolUser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"powUserAddr\",\"type\":\"address\"}],\"name\":\"BindPowUser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"burnAmount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"BuyPoc\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"BuyPool\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"BuyPow\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"CancelPoolMortage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"CancelPostMortage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"CancelPowMortage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"ChangeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"ChangeOwner1\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"times\",\"type\":\"uint256\"}],\"name\":\"InitPoc\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\"},{\"name\":\"_decimals\",\"type\":\"uint8\"}],\"name\":\"InitToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"val\",\"type\":\"uint256\"}],\"name\":\"InitUserNUC\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"MortagePool\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"MortagePost\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"MortagePow\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferNUC\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPocers\",\"outputs\":[{\"components\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"expireTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPowAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPoolers\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"powAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPosters\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"offset\",\"type\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\"}],\"name\":\"AllPowers\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"atLeastMortageAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"atLeastPostMortageAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"BURN_ADDRESS\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"burnAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetCurrentTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetMortageFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetMortageTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPocer\",\"outputs\":[{\"components\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"expireTime\",\"type\":\"uint256\"}],\"name\":\"records\",\"type\":\"tuple[]\"},{\"name\":\"userAddr\",\"type\":\"address\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPowAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPoCTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPooler\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"powAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPoolTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"GetPostBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPoster\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPostMortageTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"GetPower\",\"outputs\":[{\"components\":[{\"name\":\"createTime\",\"type\":\"uint256\"},{\"name\":\"Index\",\"type\":\"uint256\"},{\"name\":\"mortageBalance\",\"type\":\"uint256\"},{\"name\":\"bindPoolAddr\",\"type\":\"address\"},{\"name\":\"pocAddrs\",\"type\":\"address[]\"},{\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetPowTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetRewardRatio\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"mortageAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"mortageFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner1\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PocCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PocExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pocExpired\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pocTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PocValid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PoolCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PoolExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"poolTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PoolValid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PostCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PostExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PostUsers\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"powAddresses\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"PowCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PowExisted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"powTicket\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"PowValid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reduceDuration\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"start\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"topPoolMortageCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"topPowMortageCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

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

func (this *NUCCaller) AllPowers(offset, pageSize *big.Int) ([]Struct1, error) {
	method := "AllPowers"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		fmt.Println("+++++++==", err)
		return nil, err
	}
	out := new([]Struct1)
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

func (this *NUCCaller) AllPocers(offset, pageSize *big.Int) ([]Struct3, error) {
	method := "AllPocers"
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
	err = this.Abi.Unpack(out, method, output)
	return *out, err
}

func (this *NUCCaller) AllPoolers(offset, pageSize *big.Int) ([]Struct0, error) {
	method := "AllPoolers"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := new([]Struct0)
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

func (this *NUCCaller) AllPosters(offset, pageSize *big.Int) ([]Struct4, error) {
	method := "AllPosters"
	input, err := this.Abi.Pack(method, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := new([]Struct4)
	output, err := this.CallContract(input)
	if err != nil {
		return nil, err
	}
	err = this.Abi.Unpack(out, method, output)
	return *out, nil
}

func (this *NUCCaller) PosterCount() (*big.Int, error) {
	method := "PostCount"
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
