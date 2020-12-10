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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/reward"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type diffTest struct {
	ParentTimestamp    uint64
	ParentDifficulty   *big.Int
	CurrentTimestamp   uint64
	CurrentBlocknumber *big.Int
	CurrentDifficulty  *big.Int
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	config := &params.ChainConfig{HomesteadBlock: big.NewInt(1150000)}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
			Number:        number,
			Time:          test.ParentTimestamp,
			Difficulty:    test.ParentDifficulty,
			NUCDifficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
	}
}

func TestDecodeCoinbaseTXS(t *testing.T) {
	coinbaseTxs := "0xe53617f9e0f76e8b63b9741f5f670c6d81c86e96000000000000000006f05b59d3b200000000000000000000bb906570a8d83970368cacee22958f4b78138124000000000000000006f05b59d3b200000000000000000000"
	b := common.FromHex(coinbaseTxs)
	start := 0
	var value *big.Int
	var l uint32
	for {
		typ := b[start : start+1][0]
		start += 1
		switch typ {
		case 0:
			fmt.Println("type:PoW")
		case 1:
			fmt.Println("type:PoC")
		case 2:
			fmt.Println("type:Pool")
		}
		value = new(big.Int).SetBytes(b[start : start+8]) // value 8 bytes
		start += 8
		l = binary.LittleEndian.Uint32(b[start : start+4]) // length 4 bytes
		start += 4
		for i := 0; i < int(l); i++ {
			addr := common.BytesToAddress(b[start : start+20])
			start += 20
			fmt.Println("address:", addr.String(), value)
		}
		if typ == 2 {
			break
		}
	}
}

func TestBigFloatValue(t *testing.T) {
	powReward := PowReward.Div(PowReward, big.NewInt(100))
	b := FormatUint64Bytes2(powReward)
	newV := new(big.Int).SetBytes(b)
	fmt.Println(newV)
	newF := new(big.Float).SetInt(newV)
	newF.Quo(newF, big.NewFloat(1e+18))
	fmt.Println(newF)
}

func TestDecodeCoinbaseTXS2(t *testing.T) {
	data := "0x01830c3cf0e571588578b192552d52f89747ecf14563918244f400000de0b6b3a76400000000000000000000"
	ctxs := DecodeFromBytes(data)
	fmt.Println("allCount", len(*ctxs))
	for addr, v := range *ctxs {
		fmt.Println(addr.String())
		fmt.Println("PocReward", v.PocReward)
		fmt.Println("PowReward", v.PowReward)
		fmt.Println("PoolReward", v.PoolReward)
	}
}

func TestDecodeCoinbaseTXS3(t *testing.T) {
	data := reward.GetRewardsByNumber(big.NewInt(110))
	ctxs := DecodeFromBytes(common.ToHex(data))
	fmt.Println("all Addr Count", len(*ctxs))
	i := 0
	for addr, v := range *ctxs {
		if i == len(*ctxs)-1 {
			fmt.Println(addr.String())
			fmt.Println("PocReward", v.PocReward)
			fmt.Println("PowReward", v.PowReward)
			fmt.Println("PoolReward", v.PoolReward)
			fmt.Println("PostReward", v.PostReward)
		}
		if addr.String() == "0xe53617F9e0f76E8B63B9741F5f670C6d81c86e96" {
			fmt.Println(addr.String())
			fmt.Println("PocReward", v.PocReward)
			fmt.Println("PowReward", v.PowReward)
			fmt.Println("PoolReward", v.PoolReward)
			fmt.Println("PostReward", v.PostReward)
		}
		i++
	}
}

func TestBigBytes(t *testing.T) {
	powReward := new(big.Int).Set(PowBlockReward)
	ratio := int64(25)
	powReward = powReward.Div(powReward, big.NewInt(ratio))
	powFee := big.NewInt(0)
	// pow get 60% tx fee
	powFee = powFee.Mul(powFee, big.NewInt(60))
	powFee = powFee.Div(powFee, big.NewInt(100))
	powReward = powReward.Add(powReward, powFee)

	fmt.Println(powFee.String(), powReward.Bytes())
}
