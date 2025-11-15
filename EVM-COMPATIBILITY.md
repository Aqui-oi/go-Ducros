# EVM Compatibility & Fork Schedule

## Ducros Network EVM Specification

**ChainID:** `9999`
**Network ID:** `9999`
**Consensus:** RandomX PoW (CPU-friendly, ASIC-resistant)
**Block Time:** ~13 seconds (LWMA difficulty adjustment)
**Gas Limit:** 8,000,000 (0x7a1200)
**Initial Base Fee:** 1 Gwei (0x3b9aca00)

---

## Activated Hard Forks

Ducros Network starts with modern EVM features enabled from genesis (block 0):

| Fork | Block Number | EIPs | Status |
|------|--------------|------|--------|
| **Homestead** | 0 | EIP-2, EIP-7, EIP-8 | ✅ Active |
| **EIP-150** | 0 | EIP-150 (Gas cost changes) | ✅ Active |
| **EIP-155** | 0 | EIP-155 (Replay protection) | ✅ Active |
| **EIP-158** | 0 | EIP-158 (State clearing) | ✅ Active |
| **Byzantium** | 0 | EIP-140, EIP-211, EIP-214, EIP-658, EIP-649 | ✅ Active |
| **Constantinople** | 0 | EIP-145, EIP-1014, EIP-1052, EIP-1283, EIP-1234 | ✅ Active |
| **Petersburg** | 0 | Constantinople - EIP-1283 | ✅ Active |
| **Istanbul** | 0 | EIP-152, EIP-1108, EIP-1344, EIP-1884, EIP-2028, EIP-2200 | ✅ Active |
| **MuirGlacier** | 0 | EIP-2384 (Difficulty bomb delay) | ✅ Active |
| **Berlin** | 0 | EIP-2565, EIP-2718, EIP-2929, EIP-2930 | ✅ Active |
| **London** | 0 | EIP-1559, EIP-3198, EIP-3529, EIP-3541 | ✅ Active |

### ❌ NOT Activated (PoS-specific forks)

| Fork | Reason | Status |
|------|--------|--------|
| **Shanghai** | Requires PoS (withdrawals) | ❌ Not supported |
| **Cancun** | Requires PoS (blob transactions) | ❌ Not supported |
| **Prague** | Requires PoS | ❌ Not supported |

---

## EIP-1559 (Fee Market)

**Status:** ✅ **Fully Active**

- **Base Fee:** Dynamic, starts at 1 Gwei
- **Max Fee Per Gas:** User-specified maximum
- **Priority Fee:** Tip to miners
- **Base Fee Burn:** ✅ Burned (deflationary)

### Base Fee Calculation

```go
baseFee[n+1] = baseFee[n] * (1 + (gasUsed - gasTarget) / gasTarget / 8)
```

- **Gas Target:** 50% of gas limit (4,000,000)
- **Max Change:** ±12.5% per block
- **Floor:** 7 wei (minimum base fee)

---

## Opcodes Reference

### Active Opcodes (London-compatible)

| Opcode | Mnemonic | Gas Cost | EIP | Description |
|--------|----------|----------|-----|-------------|
| 0x31 | BALANCE | 2600 (warm) / 100 (warm) | EIP-2929 | Get balance of account |
| 0x3A | GASPRICE | 2 | - | Get gas price |
| 0x3D | RETURNDATASIZE | 2 | EIP-211 | Size of returned data |
| 0x3E | RETURNDATACOPY | 3 + mem | EIP-211 | Copy returned data |
| 0x3F | EXTCODEHASH | 2600 (cold) / 100 (warm) | EIP-1052 | Get code hash |
| 0x44 | CHAINID | 2 | EIP-1344 | Get chain ID (9999) |
| 0x45 | SELFBALANCE | 5 | EIP-1884 | Get own balance |
| 0x46 | BASEFEE | 2 | EIP-3198 | Get current base fee |
| 0x48 | BASEFEE | 2 | EIP-3198 | Get base fee |
| 0x54 | SLOAD | 2100 (cold) / 100 (warm) | EIP-2929 | Load from storage |
| 0x55 | SSTORE | Variable | EIP-2200 | Store to storage |
| 0x5C | TLOAD | - | - | NOT SUPPORTED (post-London) |
| 0x5D | TSTORE | - | - | NOT SUPPORTED (post-London) |
| 0xF5 | CREATE2 | 32000 + mem | EIP-1014 | Create contract with salt |

### Gas Costs (EIP-2929 - Berlin)

- **Cold account access:** 2600 gas
- **Warm account access:** 100 gas
- **Cold SLOAD:** 2100 gas
- **Warm SLOAD:** 100 gas

---

## Transaction Types

### Type 0: Legacy Transactions
```json
{
  "nonce": "0x0",
  "gasPrice": "0x3b9aca00",
  "gasLimit": "0x5208",
  "to": "0x...",
  "value": "0x0",
  "data": "0x",
  "v": "0x4e43", // chainId * 2 + 35 + {0,1}
  "r": "0x...",
  "s": "0x..."
}
```

### Type 1: Access List Transactions (EIP-2930)
```json
{
  "chainId": "0x270f",
  "nonce": "0x0",
  "gasPrice": "0x3b9aca00",
  "gasLimit": "0x5208",
  "to": "0x...",
  "value": "0x0",
  "data": "0x",
  "accessList": [
    {
      "address": "0x...",
      "storageKeys": ["0x..."]
    }
  ],
  "v": "0x0",
  "r": "0x...",
  "s": "0x..."
}
```

### Type 2: EIP-1559 Transactions
```json
{
  "chainId": "0x270f",
  "nonce": "0x0",
  "maxPriorityFeePerGas": "0x3b9aca00", // 1 Gwei tip
  "maxFeePerGas": "0x77359400",         // 2 Gwei max
  "gasLimit": "0x5208",
  "to": "0x...",
  "value": "0x0",
  "data": "0x",
  "accessList": [],
  "v": "0x0",
  "r": "0x...",
  "s": "0x..."
}
```

---

## JSON-RPC API Compatibility

### Fully Supported Methods

#### `eth_*` Namespace
- ✅ `eth_chainId` → Returns `0x270f` (9999)
- ✅ `eth_blockNumber`
- ✅ `eth_getBalance`
- ✅ `eth_getTransactionCount`
- ✅ `eth_getCode`
- ✅ `eth_call`
- ✅ `eth_estimateGas`
- ✅ `eth_gasPrice`
- ✅ `eth_maxPriorityFeePerGas`
- ✅ `eth_feeHistory`
- ✅ `eth_getBlockByNumber`
- ✅ `eth_getBlockByHash`
- ✅ `eth_getTransactionByHash`
- ✅ `eth_getTransactionReceipt`
- ✅ `eth_sendRawTransaction`
- ✅ `eth_getLogs`

#### `net_*` Namespace
- ✅ `net_version` → Returns `"9999"`
- ✅ `net_listening`
- ✅ `net_peerCount`

#### `web3_*` Namespace
- ✅ `web3_clientVersion`
- ✅ `web3_sha3`

#### Mining (RandomX-specific)
- ✅ `eth_mining`
- ✅ `eth_hashrate`
- ✅ `eth_getWork` → Returns `[headerHash, seedHash, target, blockNumber]`
- ✅ `eth_submitWork` → Accepts `[nonce, headerHash, mixDigest]`
- ✅ `eth_submitHashrate`
- ✅ `randomx_getWork` (custom) → Same as eth_getWork but explicit
- ✅ `randomx_submitWork` (custom)

### NOT Supported (PoS-specific)

- ❌ `eth_getProof` (requires state proofs)
- ❌ Beacon chain RPCs
- ❌ Blob transaction methods (EIP-4844)
- ❌ Withdrawal methods (EIP-4895)

---

## Wallet Compatibility

### MetaMask Configuration

```javascript
{
  "chainId": "0x270f", // 9999 in hex
  "chainName": "Ducros Network",
  "nativeCurrency": {
    "name": "Ducros",
    "symbol": "DUC",
    "decimals": 18
  },
  "rpcUrls": ["https://rpc.ducros.network"], // Replace with actual RPC
  "blockExplorerUrls": ["https://explorer.ducros.network"] // Replace with actual explorer
}
```

### Web3.js Example

```javascript
const Web3 = require('web3');
const web3 = new Web3('https://rpc.ducros.network');

// Check chain ID
const chainId = await web3.eth.getChainId();
console.log(chainId); // 9999

// Send EIP-1559 transaction
const tx = {
  from: '0x...',
  to: '0x...',
  value: web3.utils.toWei('1', 'ether'),
  maxPriorityFeePerGas: web3.utils.toWei('1', 'gwei'),
  maxFeePerGas: web3.utils.toWei('2', 'gwei'),
};

const receipt = await web3.eth.sendTransaction(tx);
```

### Ethers.js Example

```javascript
const { ethers } = require('ethers');
const provider = new ethers.JsonRpcProvider('https://rpc.ducros.network');

// Get network info
const network = await provider.getNetwork();
console.log(network.chainId); // 9999n

// Send transaction with EIP-1559
const wallet = new ethers.Wallet(privateKey, provider);
const tx = await wallet.sendTransaction({
  to: '0x...',
  value: ethers.parseEther('1.0'),
  maxPriorityFeePerGas: ethers.parseUnits('1', 'gwei'),
  maxFeePerGas: ethers.parseUnits('2', 'gwei'),
});
```

---

## Smart Contract Compatibility

### Solidity Version Support

✅ **Solidity 0.4.x - 0.8.x** fully supported

All London-compatible Solidity features work:
- ✅ `block.basefee` (EIP-3198)
- ✅ `block.chainid` (EIP-1344)
- ✅ `create2` (EIP-1014)
- ✅ `extcodehash` (EIP-1052)
- ✅ Access lists (EIP-2930)
- ❌ `PUSH0` opcode (Shanghai) - NOT supported

### Vyper Support

✅ **Vyper 0.2.x - 0.3.x** supported (London-compatible)

---

## Precompiled Contracts

All standard Ethereum precompiles are supported:

| Address | Name | EIP | Status |
|---------|------|-----|--------|
| 0x01 | ecRecover | - | ✅ Active |
| 0x02 | SHA256 | - | ✅ Active |
| 0x03 | RIPEMD160 | - | ✅ Active |
| 0x04 | Identity | - | ✅ Active |
| 0x05 | ModExp | EIP-2565 | ✅ Active |
| 0x06 | BN256Add | EIP-196 | ✅ Active |
| 0x07 | BN256ScalarMul | EIP-196 | ✅ Active |
| 0x08 | BN256Pairing | EIP-197 | ✅ Active |
| 0x09 | Blake2F | EIP-152 | ✅ Active |

---

## Testing & Verification

### Verify EVM Compatibility

```bash
# Check chain ID
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545

# Check base fee
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://localhost:8545 | jq '.result.baseFeePerGas'

# Test EIP-1559 transaction
cast send --legacy=false \
  --max-fee 2gwei \
  --max-priority-fee 1gwei \
  0x0000000000000000000000000000000000000001 \
  --value 0.1ether \
  --rpc-url http://localhost:8545
```

---

## Migration Guide

### From Ethereum Mainnet

**Compatible:**
- ✅ All London-compatible smart contracts
- ✅ ERC-20, ERC-721, ERC-1155 tokens
- ✅ Uniswap V2/V3 (London-compatible versions)
- ✅ Most DeFi protocols (pre-Shanghai)

**NOT Compatible:**
- ❌ Shanghai+ contracts using `PUSH0`
- ❌ Contracts expecting PoS validators
- ❌ Blob transactions (EIP-4844)

### Deployment Checklist

1. ✅ Verify Solidity version ≤ 0.8.19 (London)
2. ✅ Test with `--evm-version london`
3. ✅ Use ChainID `9999` in signatures
4. ✅ Account for RandomX block times (~13s vs 12s Ethereum)
5. ✅ Test EIP-1559 fee estimation

---

## Future Upgrades

Ducros Network may activate future EIPs via coordinated hard forks:

**Potential Future EIPs:**
- EIP-3855 (PUSH0 opcode) - Shanghai
- EIP-3860 (Init code size limit) - Shanghai
- Custom Ducros improvement proposals (DIPs)

All upgrades require community consensus and coordinated activation block.

---

## Support

- **Documentation:** https://docs.ducros.network
- **RPC Endpoint:** https://rpc.ducros.network
- **Explorer:** https://explorer.ducros.network
- **GitHub:** https://github.com/Aqui-oi/go-Ducros

---

**Last Updated:** 2025-11-12
**Network Version:** London (EIP-1559)
**Consensus:** RandomX PoW
