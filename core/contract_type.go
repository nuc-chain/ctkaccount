package core

import (
	`bytes`
	`github.com/ethereum/go-ethereum/common`
	`github.com/ethereum/go-ethereum/core/vm`
	`math/big`
)

var (
	ContractTypeDefault = []byte{0x0,0x0}
	ContractTypeIPFS = []byte{0x0,0x1}
)

type ContractObject interface {
	Create(caller vm.ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error)
	SetEVM(evm *vm.EVM)
}

type ContractBase struct {
	Evm *vm.EVM
}

func (this *ContractBase)SetEVM(evm *vm.EVM) {
	this.Evm = evm
}

type ContractDefault struct {
	ContractBase
}
//default contract
func (this *ContractDefault)Create(caller vm.ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	return this.Evm.Create(caller,code,gas,value)
}


type ContractIPFS struct {
	ContractBase
}
//ipfs contract
func (this *ContractIPFS)Create(caller vm.ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	return this.Evm.Create3(caller,code,gas,value)
}

func GetContractInstance( data []byte ) (obj ContractObject) {
	if bytes.HasPrefix(data, ContractTypeIPFS){
		return &ContractIPFS{}
	} else if bytes.HasPrefix(data, ContractTypeDefault){
		return &ContractDefault{}
	}
	return nil
}