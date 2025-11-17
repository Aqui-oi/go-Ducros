<div align="center">

```
██████╗ ██╗   ██╗ ██████╗██████╗  ██████╗ ███████╗
██╔══██╗██║   ██║██╔════╝██╔══██╗██╔═══██╗██╔════╝
██║  ██║██║   ██║██║     ██████╔╝██║   ██║███████╗
██║  ██║██║   ██║██║     ██╔══██╗██║   ██║╚════██║
██████╔╝╚██████╔╝╚██████╗██║  ██║╚██████╔╝███████║
╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝
```

# 🚀 Ducros Mainnet

### **The Next-Generation CPU-Friendly Blockchain**

*Powered by RandomX Proof-of-Work • Built with Sustainability • EVM Compatible*

---

[![Go Report Card](https://goreportcard.com/badge/github.com/Aqui-oi/go-Ducros)](https://goreportcard.com/report/github.com/Aqui-oi/go-Ducros)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)
[![ChainID](https://img.shields.io/badge/ChainID-33669-orange)](https://chainlist.org)
[![RandomX](https://img.shields.io/badge/PoW-RandomX-green)](https://github.com/tevador/RandomX)
[![Block Time](https://img.shields.io/badge/Block_Time-13s-yellow)](https://github.com/Aqui-oi/go-Ducros)

[**Quick Start**](#-quick-start) • [**Documentation**](#-documentation) • [**Mining Guide**](#-mining-guide) • [**API Reference**](#-api-reference) • [**Community**](#-community--support)

</div>

---

## 🌟 What is Ducros?

**Ducros Mainnet** is a revolutionary blockchain that combines the best of Ethereum's smart contract capabilities with Monero's CPU-friendly RandomX mining algorithm. Built for fairness, sustainability, and decentralization.

### ✨ Key Highlights

<table>
<tr>
<td width="50%">

#### 🔨 **Fair Mining**
- **ASIC-Resistant** RandomX algorithm
- Mine with your **CPU** - no GPUs needed
- **AMD Ryzen optimized** for best hashrate
- Same algorithm as **Monero (XMR)**

</td>
<td width="50%">

#### 💰 **Sustainable Treasury**
- **5%** of rewards → Development fund
- **95%** of rewards → Miners
- **Hardcoded** in consensus layer
- **Transparent** & immutable

</td>
</tr>
<tr>
<td width="50%">

#### ⚡ **High Performance**
- **13-second** block time
- **LWMA** difficulty adjustment
- **EVM compatible** smart contracts
- **JSON-RPC** API support

</td>
<td width="50%">

#### 🔒 **Advanced Features**
- **Fee exemption** whitelist system
- **Zero-fee** transactions for whitelisted addresses
- **Ethereum tooling** compatibility
- **Web3** integration ready

</td>
</tr>
</table>

---

## 🎯 Network Information

<div align="center">

| Parameter | Value | Description |
|:---------:|:-----:|:-----------:|
| 🌐 **Network** | **Ducros Mainnet** | Official production network |
| 🔗 **Chain ID** | **33669** | Unique blockchain identifier |
| ⛏️ **Consensus** | **RandomX PoW** | CPU-optimized mining |
| ⏱️ **Block Time** | **~13 seconds** | LWMA-adjusted |
| 💎 **Block Reward** | **2 DCR** | 1.9 DCR miner + 0.1 DCR treasury |
| 📊 **Difficulty** | **LWMA** | Anti-timewarp protection |
| 💰 **Symbol** | **DCR** | Ducros |
| 🔧 **EVM** | **✅ Compatible** | Full Ethereum support |

</div>

---

## 🚀 Quick Start

### 📦 Installation (3 Steps)

```bash
# 1️⃣ Clone the repository
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros

# 2️⃣ Build geth
make geth

# 3️⃣ Initialize blockchain
./build/bin/geth init --datadir ./ducros-data genesis-production.json
```

### 🎮 Run Your Node

<table>
<tr>
<td width="50%">

#### 📡 **Full Node** (Validator)

```bash
./build/bin/geth \
  --datadir ./ducros-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx \
  --http.addr 0.0.0.0 \
  --http.port 8545
```

✅ Validates transactions
✅ Relays blocks
✅ Exposes RPC API

</td>
<td width="50%">

#### ⛏️ **Mining Node** (Earn DCR!)

```bash
./build/bin/geth \
  --datadir ./ducros-data \
  --networkid 33669 \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xYOUR_ADDRESS \
  --http --http.api eth,net,web3,miner
```

✅ Mines DCR
✅ Secures network
✅ Earns 1.9 DCR per block

</td>
</tr>
</table>

### 🐳 Docker Quick Start

```bash
# Pull and run in one command
docker run -d \
  --name ducros-node \
  -v $(pwd)/data:/root/.ducros \
  -p 8545:8545 -p 30303:30303 \
  ducros/mainnet:latest \
  --http --http.addr 0.0.0.0 --mine --miner.threads 4
```

---

## 💻 Hardware Requirements

<div align="center">

### ⚙️ Choose Your Setup

</div>

<table>
<thead>
<tr>
<th width="25%">🏷️ Type</th>
<th width="25%">⚡ Minimum</th>
<th width="25%">✅ Recommended</th>
<th width="25%">🏆 Mining Node</th>
</tr>
</thead>
<tbody>
<tr>
<td><b>CPU</b></td>
<td>2+ cores</td>
<td>4+ cores @ 2.5 GHz</td>
<td><b>6+ cores (AMD Ryzen)</b></td>
</tr>
<tr>
<td><b>RAM</b></td>
<td>4 GB</td>
<td>8 GB</td>
<td><b>16 GB</b></td>
</tr>
<tr>
<td><b>Storage</b></td>
<td>50 GB SSD</td>
<td>100 GB SSD</td>
<td><b>250 GB NVMe</b></td>
</tr>
<tr>
<td><b>Network</b></td>
<td>5 Mbps</td>
<td>10 Mbps</td>
<td><b>25+ Mbps</b></td>
</tr>
<tr>
<td><b>Use Case</b></td>
<td>Validator only</td>
<td>Full node</td>
<td><b>Competitive mining</b></td>
</tr>
<tr>
<td><b>Est. Cost</b></td>
<td>~$5-10/month</td>
<td>~$15-20/month</td>
<td><b>~$50-100/month</b></td>
</tr>
</tbody>
</table>

> 💡 **Pro Tip**: AMD Ryzen CPUs offer the best hashrate/$ ratio for RandomX mining!

---

## ⛏️ Mining Guide

### 🎯 Expected Hashrates

<div align="center">

| CPU Model | Cores | Hashrate | Monthly DCR* |
|:----------|:-----:|:--------:|:-----------:|
| Intel i5-9400 | 6 | 3-5 KH/s | ~150-250 DCR |
| **AMD Ryzen 5 3600** ⭐ | 6 | 6-9 KH/s | **~300-450 DCR** |
| **AMD Ryzen 7 5800X** ⭐⭐ | 8 | 10-15 KH/s | **~500-750 DCR** |
| **AMD Ryzen 9 5950X** 🏆 | 16 | 20-25 KH/s | **~1000-1250 DCR** |

<sub>*Estimates based on current difficulty. Actual results may vary.</sub>

</div>

### 🔧 Mining Options

<table>
<tr>
<td width="50%">

#### 🖥️ **Solo Mining (Built-in)**

```bash
# Start mining with 4 threads
./build/bin/geth \
  --datadir ./ducros-data \
  --networkid 33669 \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xYOUR_ADDRESS
```

**Pros:**
- ✅ Keep 100% of rewards
- ✅ No pool fees
- ✅ Full control

**Cons:**
- ⚠️ Higher variance
- ⚠️ Requires full node

</td>
<td width="50%">

#### 🎱 **Pool Mining (xmrig)**

```bash
# 1. Start stratum proxy
cd stratum-proxy
./stratum-proxy \
  --stratum 0.0.0.0:3333 \
  --geth http://localhost:8545

# 2. Connect xmrig
xmrig -o localhost:3333 \
  -u YOUR_ADDRESS \
  --algo rx/0 \
  -t 4
```

**Pros:**
- ✅ Consistent payouts
- ✅ Lower variance
- ✅ Professional mining software

**Cons:**
- ⚠️ Pool fees (if using pool)
- ⚠️ Need proxy setup

</td>
</tr>
</table>

### 💰 Block Rewards Breakdown

```
Total Block Reward: 2.000 DCR
├─ 95% → Miner:     1.900 DCR  💎
└─ 5%  → Treasury:  0.100 DCR  🏦

Blocks per day: ~6,646 (at 13s block time)
Daily emission: ~13,292 DCR
```

---

## 🛠️ Building from Source

### 📋 Prerequisites

<table>
<tr>
<td width="33%">

#### 🐹 **Golang**
Version **1.23** or later

```bash
# Ubuntu/Debian
sudo apt install golang-go

# macOS
brew install go

# Check version
go version
```

</td>
<td width="33%">

#### ⚙️ **C Compiler**
gcc or clang

```bash
# Ubuntu/Debian
sudo apt install build-essential

# macOS (installs Xcode tools)
xcode-select --install

# Check
gcc --version
```

</td>
<td width="33%">

#### 📦 **Git**
Latest version

```bash
# Ubuntu/Debian
sudo apt install git

# macOS
brew install git

# Check
git --version
```

</td>
</tr>
</table>

### 🔨 Build Commands

```bash
# Clone repository
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros

# Build geth only (recommended)
make geth

# Or build all utilities
make all

# Binaries will be in ./build/bin/
ls -lh ./build/bin/
```

### 🧪 Run Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./consensus/randomx/...

# Run with coverage
make test-coverage
```

---

## 🎮 Interactive Console

Launch the interactive JavaScript console to interact with your node:

```bash
./build/bin/geth --datadir ./ducros-data --networkid 33669 console
```

### 📝 Example Commands

```javascript
// Check block number
> eth.blockNumber
12345

// Check your balance
> eth.getBalance("0xYourAddress")
"1900000000000000000"  // 1.9 DCR

// Create new account
> personal.newAccount("password")
"0x1234567890abcdef..."

// Start mining with 4 threads
> miner.start(4)
true

// Check mining status
> eth.mining
true

// Get current hashrate
> eth.hashrate
5500  // ~5.5 KH/s

// Stop mining
> miner.stop()
true

// Send transaction
> eth.sendTransaction({
    from: "0xYourAddress",
    to: "0xRecipient",
    value: web3.toWei(10, "ether"),
    gas: 21000
  })

// Get transaction receipt
> eth.getTransactionReceipt("0xTxHash...")
```

---

## 🌐 API Reference

### 🔌 JSON-RPC Endpoints

<table>
<thead>
<tr>
<th width="30%">🚀 Enable RPC</th>
<th width="70%">Command</th>
</tr>
</thead>
<tbody>
<tr>
<td><b>HTTP</b></td>
<td>

```bash
--http --http.addr 0.0.0.0 --http.port 8545 --http.api eth,net,web3,randomx
```

</td>
</tr>
<tr>
<td><b>WebSocket</b></td>
<td>

```bash
--ws --ws.addr 0.0.0.0 --ws.port 8546 --ws.api eth,net,web3
```

</td>
</tr>
<tr>
<td><b>IPC</b></td>
<td>

Enabled by default at `~/.ducros/geth.ipc`

</td>
</tr>
</tbody>
</table>

### 📡 Available APIs

<div align="center">

| API | Description | Security Level |
|:---:|:------------|:--------------:|
| `eth` | Ethereum-compatible transactions & queries | 🟢 Safe |
| `net` | Network information & peer count | 🟢 Safe |
| `web3` | Web3 utilities | 🟢 Safe |
| `randomx` | RandomX mining stats & controls | 🟡 Moderate |
| `miner` | Mining operations (start/stop/setEtherbase) | 🟡 Moderate |
| `personal` | Account management | 🔴 Private |
| `admin` | Node administration | 🔴 Private |
| `debug` | Debugging & diagnostics | 🔴 Private |

</div>

> ⚠️ **Security Warning**: Never expose `personal`, `admin`, or `debug` APIs publicly!

### 🧪 Example API Calls

```bash
# Get latest block number
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Get account balance
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_getBalance",
    "params":["0xYourAddress","latest"],
    "id":1
  }'

# Get current hashrate
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}'

# Check mining status
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}'
```

---

## 🏦 Treasury System

### 💎 How It Works

Ducros includes a **built-in treasury** for sustainable blockchain development:

```
Every Block:
┌─────────────────────────────────────┐
│  Block Reward: 2.000 DCR            │
├─────────────────────────────────────┤
│  ├─ 95% → Miner:     1.900 DCR  💰 │
│  └─ 5%  → Treasury:  0.100 DCR  🏦 │
└─────────────────────────────────────┘

Daily Treasury Income: ~665 DCR
Monthly Treasury Income: ~20,000 DCR
Yearly Treasury Income: ~240,000 DCR
```

### 🔒 Key Features

- ✅ **Hardcoded** in consensus layer (`consensus/randomx/consensus.go`)
- ✅ **Immutable** - cannot be changed without network upgrade
- ✅ **Transparent** - all funds visible on-chain
- ✅ **Automatic** - no manual intervention needed

### 🔍 Verify Treasury

```bash
# Check treasury balance
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_getBalance",
    "params":["0xTREASURY_ADDRESS","latest"],
    "id":1
  }'
```

---

## 🔧 Configuration

### 📝 Using Config File

Instead of long command lines, use a TOML config file:

```bash
# Generate config from current flags
./build/bin/geth --your-flags dumpconfig > config.toml

# Run with config
./build/bin/geth --config config.toml
```

### 📄 Example config.toml

```toml
[Eth]
NetworkId = 33669
SyncMode = "full"

[Node]
DataDir = "./ducros-data"
HTTPHost = "0.0.0.0"
HTTPPort = 8545
HTTPModules = ["eth", "net", "web3", "randomx", "miner"]

[Node.P2P]
MaxPeers = 50
NoDiscovery = false

[Eth.Miner]
Etherbase = "0xYOUR_MINING_ADDRESS"
ExtraData = "Ducros Miner"

[Eth.TxPool]
PriceLimit = 1
```

---

## 🐳 Docker Deployment

### 🚀 Quick Deploy

```bash
# Build image
docker build -t ducros-node .

# Run validator node
docker run -d \
  --name ducros-validator \
  -v ducros-data:/root/.ducros \
  -p 8545:8545 \
  -p 30303:30303 \
  ducros-node \
  --http --http.addr 0.0.0.0

# Run mining node
docker run -d \
  --name ducros-miner \
  -v ducros-data:/root/.ducros \
  -p 8545:8545 \
  -p 30303:30303 \
  ducros-node \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xYOUR_ADDRESS
```

### 🔄 Docker Compose

```yaml
version: '3.8'
services:
  ducros-node:
    build: .
    container_name: ducros-mainnet
    ports:
      - "8545:8545"
      - "30303:30303"
    volumes:
      - ./data:/root/.ducros
    command: >
      --networkid 33669
      --http --http.addr 0.0.0.0
      --mine --miner.threads 4
      --miner.etherbase 0xYOUR_ADDRESS
    restart: unless-stopped
```

---

## 🛡️ Security Best Practices

<table>
<tr>
<td width="50%">

### ✅ **DO**

- ✅ Use **firewalls** to restrict RPC access
- ✅ Only expose **necessary APIs** (`eth,net,web3`)
- ✅ Use **strong passwords** for keystores
- ✅ **Backup** your keystore regularly
- ✅ Keep your node **updated**
- ✅ Monitor your node **logs**
- ✅ Use **HTTPS** for remote RPC
- ✅ Enable **rate limiting** if public

</td>
<td width="50%">

### ❌ **DON'T**

- ❌ **Never** expose `admin` API publicly
- ❌ **Never** expose `debug` API publicly
- ❌ **Never** expose `personal` API publicly
- ❌ **Don't** share your private keys
- ❌ **Don't** reuse passwords
- ❌ **Don't** run as root user
- ❌ **Don't** disable firewall
- ❌ **Don't** trust unverified binaries

</td>
</tr>
</table>

### 🚨 Reporting Vulnerabilities

Found a security issue? **Please report responsibly:**

1. ❌ **DO NOT** open a public GitHub issue
2. ✅ Email maintainers privately with details
3. ⏳ Wait for confirmation before public disclosure
4. 🏆 Receive credit in security acknowledgments

---

## 📚 Documentation

### 📖 Guides & Tutorials

| Document | Description |
|:---------|:------------|
| 📘 [**Treasury Implementation**](./TREASURY_IMPLEMENTATION.md) | Complete guide to the 95/5 treasury system |
| 🔧 [**RandomX Segfault Fix**](./RANDOMX_SEGFAULT_FIX.md) | Technical details on the threading fix |
| ⚙️ [**Difficulty Adjustment**](./DIFFICULTY_ADJUSTMENT.md) | How LWMA difficulty algorithm works |
| 🎱 [**Stratum Proxy Setup**](./START_STRATUM_PROXY.md) | Configure external mining with xmrig |
| 📊 [**Changelog**](./CHANGELOG_TREASURY.md) | Version history and updates |

### 🔗 External Resources

- 🌐 [**Official Website**](https://ducros.network) *(coming soon)*
- 📊 [**Block Explorer**](https://explorer.ducros.network) *(coming soon)*
- 💬 [**Community Forum**](https://forum.ducros.network) *(coming soon)*
- 📈 [**Network Stats**](https://stats.ducros.network) *(coming soon)*

---

## 🤝 Contributing

We ❤️ contributions from the community!

### 🎯 How to Contribute

```bash
# 1. Fork the repository
# 2. Create your feature branch
git checkout -b feature/amazing-feature

# 3. Commit your changes
git commit -m 'feat(consensus): add amazing feature'

# 4. Push to your branch
git push origin feature/amazing-feature

# 5. Open a Pull Request
```

### 📝 Commit Message Format

```
<type>(<scope>): <subject>

Examples:
feat(mining): add CPU affinity support
fix(consensus): resolve LWMA edge case
docs(readme): update hardware requirements
refactor(randomx): optimize dataset initialization
```

### 🧪 Code Standards

- ✅ Code must pass `gofmt`
- ✅ All tests must pass (`make test`)
- ✅ Document all exported functions
- ✅ Follow Go best practices
- ✅ Add tests for new features

---

## 📜 License

<div align="center">

### Open Source & Free Forever

</div>

<table>
<tr>
<td width="50%">

#### 📚 **Library Code**
All code outside `cmd/` directory

**License:** [GNU LGPL v3.0](./COPYING.LESSER)

*Allows linking from proprietary software*

</td>
<td width="50%">

#### 🛠️ **Binary Tools**
All code inside `cmd/` directory

**License:** [GNU GPL v3.0](./COPYING)

*Ensures all modifications remain open source*

</td>
</tr>
</table>

---

## 🙏 Acknowledgments

<div align="center">

**Ducros stands on the shoulders of giants**

</div>

| Project | Contribution |
|:--------|:-------------|
| [**go-ethereum**](https://github.com/ethereum/go-ethereum) | Core blockchain infrastructure & EVM |
| [**RandomX**](https://github.com/tevador/RandomX) | CPU-optimized PoW algorithm (Monero Research Lab) |
| [**LWMA**](https://github.com/zawy12/difficulty-algorithms) | Difficulty adjustment algorithm by Zawy |
| [**Ethereum Foundation**](https://ethereum.org) | Smart contract platform & tooling |

---

## 💬 Community & Support

<div align="center">

### Join the Ducros Community!

[![GitHub](https://img.shields.io/badge/GitHub-Aqui--oi/go--Ducros-181717?logo=github)](https://github.com/Aqui-oi/go-Ducros)
[![Issues](https://img.shields.io/github/issues/Aqui-oi/go-Ducros)](https://github.com/Aqui-oi/go-Ducros/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/Aqui-oi/go-Ducros)](https://github.com/Aqui-oi/go-Ducros/pulls)
[![Contributors](https://img.shields.io/github/contributors/Aqui-oi/go-Ducros)](https://github.com/Aqui-oi/go-Ducros/graphs/contributors)

</div>

### 🆘 Get Help

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/Aqui-oi/go-Ducros/issues/new?template=bug_report.md)
- 💡 **Feature Requests**: [GitHub Issues](https://github.com/Aqui-oi/go-Ducros/issues/new?template=feature_request.md)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Aqui-oi/go-Ducros/discussions)
- 📧 **Email**: support@ducros.network *(coming soon)*

### 🌍 Network Info

```
Network Name:    Ducros Mainnet
Chain ID:        33669
RPC Endpoint:    https://rpc.ducros.network (coming soon)
Explorer:        https://explorer.ducros.network (coming soon)
```

---

<div align="center">

## 🌟 Star Us!

**If you like Ducros, give us a ⭐ on GitHub!**

### Built with ❤️ for a Decentralized Future

```
┌─────────────────────────────────────────────┐
│  FAIR MINING • SUSTAINABLE TREASURY         │
│  CPU-FRIENDLY • EVM COMPATIBLE              │
│  COMMUNITY DRIVEN • OPEN SOURCE             │
└─────────────────────────────────────────────┘
```

**© 2025 Ducros Mainnet • Empowering True Decentralization**

[⬆️ Back to Top](#-ducros-mainnet)

</div>
