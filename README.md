# Go Ducros

Golang execution layer implementation of the Ducros protocol with RandomX Proof-of-Work.

[![Go Report Card](https://goreportcard.com/badge/github.com/Aqui-oi/go-Ducros)](https://goreportcard.com/report/github.com/Aqui-oi/go-Ducros)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

**Ducros Mainnet** is a CPU-friendly blockchain powered by RandomX Proof-of-Work with an integrated treasury system for sustainable development.

## Key Features

- 🔨 **RandomX PoW**: ASIC-resistant, CPU-optimized mining algorithm (same as Monero)
- 💰 **Treasury System**: 5% of block rewards fund development (95% to miners)
- ⚡ **Fast Blocks**: 13-second block time with LWMA difficulty adjustment
- 🔒 **Fee Exemption**: Whitelist system for zero-fee transactions
- 🌐 **EVM Compatible**: Full Ethereum smart contract support

---

## Building the source

### Prerequisites

Building `geth` requires:
- **Go** (version 1.23 or later)
- **C compiler** (gcc or clang)
- **Git**

You can install them using your favourite package manager:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y build-essential git golang-go
```

**macOS:**
```bash
brew install go git
```

### Build Instructions

Once dependencies are installed, clone the repository and build:

```bash
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros
make geth
```

Or, to build the full suite of utilities:

```bash
make all
```

The compiled binaries will be available in `./build/bin/`.

---

## Executables

The go-Ducros project comes with several wrappers/executables found in the `cmd` directory:

| Command | Description |
|---------|-------------|
| **geth** | Main Ducros CLI client. Entry point into the Ducros network, capable of running as a full node or mining node. Supports JSON-RPC over HTTP, WebSocket and IPC. |
| **clef** | Stand-alone signing tool for secure key management. |
| **devp2p** | Utilities to interact with nodes on the P2P networking layer. |
| **abigen** | Source code generator to convert Ethereum contract ABIs into Go packages. |
| **evm** | Developer utility version of the EVM for testing and debugging. |
| **rlpdump** | Developer utility to convert binary RLP dumps to readable format. |

---

## Hardware Requirements

### **Minimum (Full Node - Validation Only)**

- CPU with 2+ cores
- 4GB RAM
- 50GB free storage space (SSD recommended)
- 5 MBit/sec download Internet service

### **Recommended (Full Node)**

- Fast CPU with 4+ cores
- 8GB+ RAM
- High-performance SSD with at least 100GB of free space
- 10+ MBit/sec download Internet service

### **Mining Node (Full Node + Mining)**

- Fast CPU with 6+ cores (AMD Ryzen recommended for RandomX)
- 16GB+ RAM
- High-performance NVMe SSD with at least 250GB of free space
- 25+ MBit/sec download Internet service

**Note**: RandomX mining is CPU-optimized. AMD Ryzen processors typically offer the best hashrate per dollar. GPUs provide no advantage for RandomX mining.

---

## Running geth

### Initialize the blockchain

First, initialize your node with the genesis block:

```bash
./build/bin/geth init --datadir ./ducros-data genesis-production.json
```

### Start a full node (no mining)

```bash
./build/bin/geth \
  --datadir ./ducros-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx \
  --http.addr 0.0.0.0 \
  --http.port 8545
```

This will start a full node that:
- Syncs with the Ducros Mainnet (ChainId: 33669)
- Exposes JSON-RPC API on `http://localhost:8545`
- Validates transactions and relays blocks

### Start a mining node

```bash
./build/bin/geth \
  --datadir ./ducros-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx,miner \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xYOUR_MINING_ADDRESS
```

Options:
- `--mine`: Enable mining
- `--miner.threads 4`: Use 4 CPU threads for mining (adjust based on your CPU)
- `--miner.etherbase`: Address to receive mining rewards

**Tip**: Start with `--miner.threads 2` and increase based on your CPU cores. Leave 1-2 cores free for system operations.

### Interactive console

Start geth with an interactive JavaScript console:

```bash
./build/bin/geth --datadir ./ducros-data --networkid 33669 console
```

You can then interact with your node using Web3 JavaScript API:

```javascript
> eth.blockNumber
12345

> eth.getBalance("0xYourAddress")
"1000000000000000000"

> miner.start(4)  // Start mining with 4 threads
> miner.stop()    // Stop mining
```

---

## Configuration

As an alternative to passing numerous flags, you can use a configuration file:

```bash
./build/bin/geth --config /path/to/config.toml
```

To generate a config file from your current flags:

```bash
./build/bin/geth --your-favourite-flags dumpconfig > config.toml
```

---

## Docker Quick Start

Run Ducros with Docker:

```bash
docker build -t ducros-node .

docker run -d \
  --name ducros-node \
  -v /path/to/data:/root/.ducros \
  -p 8545:8545 \
  -p 30303:30303 \
  ducros-node \
  --http --http.addr 0.0.0.0 --http.api eth,net,web3,randomx
```

For mining:

```bash
docker run -d \
  --name ducros-miner \
  -v /path/to/data:/root/.ducros \
  -p 8545:8545 \
  -p 30303:30303 \
  ducros-node \
  --http --http.addr 0.0.0.0 \
  --mine --miner.threads 4 \
  --miner.etherbase 0xYOUR_ADDRESS
```

**Important**: Always use `--http.addr 0.0.0.0` if you want to access RPC from outside the container.

---

## Mining Information

### RandomX Algorithm

Ducros uses **RandomX**, the same CPU-optimized PoW algorithm as Monero:
- **ASIC-resistant**: Designed to run efficiently only on general-purpose CPUs
- **Fair distribution**: No advantage for specialized hardware
- **Memory-hard**: Requires 2+ GB RAM per mining thread

### Expected Hashrates

| CPU | Cores | Approx Hashrate |
|-----|-------|-----------------|
| Intel i5-9400 | 6 | ~3-5 KH/s |
| AMD Ryzen 5 3600 | 6 | ~6-9 KH/s |
| AMD Ryzen 7 5800X | 8 | ~10-15 KH/s |
| AMD Ryzen 9 5950X | 16 | ~20-25 KH/s |

### Block Rewards

- **Base Reward**: 2 DCR per block (Constantinople era)
- **Miner Share**: 95% (1.9 DCR)
- **Treasury Share**: 5% (0.1 DCR)
- **Block Time**: ~13 seconds (LWMA adjusted)

### External Mining (Stratum)

You can use external miners like **xmrig** with the stratum-proxy:

```bash
# Start stratum-proxy
cd stratum-proxy
go build
./stratum-proxy \
  --stratum 0.0.0.0:3333 \
  --geth http://localhost:8545 \
  --diff 30000 \
  --algo rx/0

# Configure xmrig to connect to localhost:3333
```

---

## Programmatically Interfacing with Geth

### JSON-RPC API

Geth supports JSON-RPC APIs over multiple transports:

**Enable HTTP RPC:**
```bash
--http                           # Enable HTTP-RPC server
--http.addr 0.0.0.0             # Listening interface (default: localhost)
--http.port 8545                # Listening port
--http.api eth,net,web3,randomx # APIs to expose
--http.corsdomain "*"           # CORS domains (use cautiously)
```

**Enable WebSocket RPC:**
```bash
--ws                      # Enable WS-RPC server
--ws.addr 0.0.0.0        # Listening interface
--ws.port 8546           # Listening port
--ws.api eth,net,web3    # APIs to expose
--ws.origins "*"         # Origins (use cautiously)
```

**IPC is enabled by default** at `~/.ducros/geth.ipc` (Unix) or `\.\pipe\geth.ipc` (Windows).

### Example: Get Block Number

```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

### Available APIs

- `eth`: Ethereum-compatible API
- `net`: Network information
- `web3`: Web3 utilities
- `randomx`: RandomX mining stats and controls
- `miner`: Mining operations (start/stop/setEtherbase)
- `admin`: Node administration
- `debug`: Debugging utilities

⚠️ **Security Warning**: Only expose APIs you need. Never expose `admin` or `debug` APIs publicly.

---

## Treasury System

Ducros includes a built-in treasury for sustainable development:

- **5% of all block rewards** go to the treasury address
- **95% of all block rewards** go to miners
- Treasury address is **hardcoded** in the consensus layer
- Cannot be modified without recompiling the entire network

### Verifying Treasury Distribution

Check the treasury balance:

```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_getBalance",
    "params":["0xTREASURY_ADDRESS", "latest"],
    "id":1
  }'
```

---

## Network Information

| Parameter | Value |
|-----------|-------|
| **Network Name** | Ducros Mainnet |
| **Chain ID** | 33669 |
| **Consensus** | RandomX Proof-of-Work |
| **Block Time** | ~13 seconds (LWMA) |
| **Block Reward** | 2 DCR (95% miner, 5% treasury) |
| **Difficulty Algorithm** | LWMA (Linearly Weighted Moving Average) |
| **Currency Symbol** | DCR |

---

## Contributing

Thank you for considering contributing to go-Ducros! We welcome contributions from anyone.

### Development Guidelines

1. **Fork** the repository
2. Create a **feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. Open a **Pull Request**

### Code Standards

- Code must adhere to **official Go formatting** (`gofmt`)
- Code must be **documented** following Go commentary guidelines
- Pull requests must be based on the `master` branch
- Commit messages should be **prefixed with package names**
  - Example: `consensus, core: implement fee exemption system`

### Testing

Run the test suite:

```bash
make test
```

---

## Security

### Reporting Vulnerabilities

If you discover a security vulnerability, please **DO NOT** open a public issue. Instead:

1. Email the maintainers privately
2. Provide detailed information about the vulnerability
3. Wait for confirmation before public disclosure

### Best Practices

- **Never expose** `admin` or `debug` APIs publicly
- **Use firewalls** to restrict RPC access
- **Keep your node updated** with the latest releases
- **Backup your keystore** regularly
- **Use strong passwords** for account encryption

---

## Documentation

- **Treasury Implementation**: [TREASURY_IMPLEMENTATION.md](./TREASURY_IMPLEMENTATION.md)
- **RandomX Segfault Fix**: [RANDOMX_SEGFAULT_FIX.md](./RANDOMX_SEGFAULT_FIX.md)
- **Difficulty Adjustment**: [DIFFICULTY_ADJUSTMENT.md](./DIFFICULTY_ADJUSTMENT.md)
- **Stratum Proxy Setup**: [START_STRATUM_PROXY.md](./START_STRATUM_PROXY.md)

---

## License

The go-Ducros library (all code outside of the `cmd` directory) is licensed under the [GNU Lesser General Public License v3.0](./COPYING.LESSER).

The go-Ducros binaries (all code inside the `cmd` directory) are licensed under the [GNU General Public License v3.0](./COPYING).

---

## Community & Support

- **GitHub**: [github.com/Aqui-oi/go-Ducros](https://github.com/Aqui-oi/go-Ducros)
- **Issues**: Report bugs and request features on GitHub Issues
- **Chain ID**: 33669

---

## Acknowledgments

- Based on [go-ethereum](https://github.com/ethereum/go-ethereum) by the Ethereum Foundation
- RandomX algorithm by [Tevador](https://github.com/tevador/RandomX) (Monero Research Lab)
- LWMA difficulty algorithm by [Zawy](https://github.com/zawy12)

---

**Built with ❤️ for a decentralized future**
