// Copyright 2015 The go-ethereum Authors
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

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var MainnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	"enode://02b75ee2d0a345ee487929277922844758e18086e52138b443c1192acaa950a75db96320223afa229d9aa9203f4dd53db9c6ee375550faf0b67197d17614b29a@47.75.149.85:30303", // bootnode-nuc-001
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://02b75ee2d0a345ee487929277922844758e18086e52138b443c1192acaa950a75db96320223afa229d9aa9203f4dd53db9c6ee375550faf0b67197d17614b29a@47.75.149.85:30303", // bootnode-nuc-001
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://02b75ee2d0a345ee487929277922844758e18086e52138b443c1192acaa950a75db96320223afa229d9aa9203f4dd53db9c6ee375550faf0b67197d17614b29a@47.75.149.85:30303", // bootnode-nuc-001
}

// GoerliBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Görli test network.
var GoerliBootnodes = []string{
	"enode://02b75ee2d0a345ee487929277922844758e18086e52138b443c1192acaa950a75db96320223afa229d9aa9203f4dd53db9c6ee375550faf0b67197d17614b29a@47.75.149.85:30303", // bootnode-nuc-001
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://02b75ee2d0a345ee487929277922844758e18086e52138b443c1192acaa950a75db96320223afa229d9aa9203f4dd53db9c6ee375550faf0b67197d17614b29a@47.75.149.85:30303", // bootnode-nuc-001
}
