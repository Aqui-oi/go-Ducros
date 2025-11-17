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
	// Ducros Mainnet Bootstrap Nodes
	//
	// IMPORTANT: Replace these with your actual bootnode enodes before mainnet launch!
	//
	// How to generate bootnodes:
	// 1. Generate key: bootnode -genkey=bootnode.key
	// 2. Get enode:    bootnode -nodekey=bootnode.key -writeaddress
	// 3. Deploy on stable servers with public IPs
	// 4. Add enode URLs below
	//
	// Format: "enode://[public_key]@[IP_or_domain]:[port]"
	//
	// TODO: Add your production bootnodes here:
	// "enode://YOUR_BOOTNODE1_PUBKEY@bootnode1.ducros.network:30303",
	// "enode://YOUR_BOOTNODE2_PUBKEY@bootnode2.ducros.network:30303",
	// "enode://YOUR_BOOTNODE3_PUBKEY@bootnode3.ducros.network:30303",
	// "enode://YOUR_BOOTNODE4_PUBKEY@your-ip-address:30303",
	// "enode://YOUR_BOOTNODE5_PUBKEY@your-ip-address:30304",
}

// DucrosTestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the Ducros test network.
var DucrosTestnetBootnodes = []string{
	// Ducros Testnet Bootnodes
	// TODO: Add testnet bootnodes for development
}
