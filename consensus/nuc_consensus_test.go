// Copyright 2019 The nuc Team

package consensus

import (
	"log"
	"math/big"
	"testing"
)

func TestDiffDecrease(t *testing.T) {
	a := big.NewInt(65536)
	b := a.Div(a, big.NewInt(2))
	log.Println(a, b)
}
