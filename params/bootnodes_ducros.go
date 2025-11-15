// Copyright 2024 The go-ethereum Authors
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

// DucrosBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the Ducros network.
var DucrosBootnodes = []string{
	// Ducros Foundation Bootnodes
	// TODO: Replace with actual production bootnodes once deployed
	// Generate bootnode keys with: bootnode -genkey=bootnode.key
	// Get enode URL with: bootnode -nodekey=bootnode.key -writeaddress

	// Example format (replace with real enodes):
	// "enode://pubkey@bootnode1.ducros.network:30303",
	// "enode://pubkey@bootnode2.ducros.network:30303",
	// "enode://pubkey@bootnode3.ducros.network:30304",
}

// DucrosTestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the Ducros test network.
var DucrosTestnetBootnodes = []string{
	// Ducros Testnet Bootnodes
	// TODO: Add testnet bootnodes for development
}
