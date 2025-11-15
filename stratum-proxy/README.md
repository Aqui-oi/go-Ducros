# Ducros Stratum Proxy

**RandomX Mining Bridge for xmrig Compatibility**

---

## 🎯 Overview

The Ducros Stratum Proxy bridges **xmrig** (and other Monero/RandomX miners) to **Ducros Network** by translating between:

- **Stratum protocol** (what miners speak) ↔ **JSON-RPC** (what Geth speaks)
- **Monero job format** ↔ **Ethereum work format**
- **Epoch-based RandomX seeds** ↔ **Ducros blocks**

---

## ✨ Features

✅ **xmrig Compatible** - Works with standard xmrig miners
✅ **Epoch-Aware** - Uses Ducros 2048-block epoch system
✅ **Zero External Dependencies** - Pure Go stdlib
✅ **Multi-Miner** - Supports multiple concurrent miners
✅ **Difficulty Adjustment** - Auto-adjusts per-miner difficulty
✅ **Statistics** - Real-time hashrate and share tracking
✅ **Pool Mode** - Optional pool fee and address

---

## 🚀 Quick Start

### 1. Build

```bash
cd stratum-proxy
go build -o stratum-proxy
```

### 2. Run

```bash
./stratum-proxy \
  --stratum "0.0.0.0:3333" \
  --geth "http://localhost:8545" \
  --diff 10000
```

### 3. Connect xmrig

```bash
xmrig \
  -o localhost:3333 \
  -u YOUR_DUCROS_ADDRESS \
  -p x \
  --algo rx/0 \
  --coin monero
```

---

## 📋 Command-Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--stratum` | `0.0.0.0:3333` | Stratum server listen address |
| `--geth` | `http://localhost:8545` | Geth JSON-RPC endpoint |
| `--diff` | `10000` | Initial difficulty for miners |
| `--pool-addr` | `` | Pool payout address (optional) |
| `--pool-fee` | `1.0` | Pool fee percentage (1.0 = 1%) |
| `--algo` | `rx/0` | RandomX algorithm variant |
| `-v` | `false` | Verbose logging |

---

## 🔧 Configuration Examples

### Solo Mining (No Pool)

```bash
./stratum-proxy \
  --stratum "0.0.0.0:3333" \
  --geth "http://localhost:8545"
```

Miners use their own address directly. No pool fees.

### Pool Mode

```bash
./stratum-proxy \
  --stratum "0.0.0.0:3333" \
  --geth "http://localhost:8545" \
  --pool-addr "0xYourPoolAddress" \
  --pool-fee 2.0
```

All mining rewards go to pool address. 2% fee deducted.

### High-Difficulty (Farm)

```bash
./stratum-proxy \
  --stratum "0.0.0.0:3333" \
  --geth "http://localhost:8545" \
  --diff 100000
```

Higher initial difficulty for mining farms with many rigs.

---

## 🖥️ xmrig Configuration

### Basic xmrig Command

```bash
xmrig \
  -o YOUR_PROXY_IP:3333 \
  -u YOUR_DUCROS_WALLET_ADDRESS \
  -p workerName \
  --algo rx/0 \
  --coin monero \
  --randomx-mode light
```

### xmrig Config File

Create `config.json`:

```json
{
  "autosave": true,
  "cpu": true,
  "opencl": false,
  "cuda": false,
  "pools": [
    {
      "algo": "rx/0",
      "coin": "monero",
      "url": "YOUR_PROXY_IP:3333",
      "user": "YOUR_DUCROS_WALLET_ADDRESS",
      "pass": "worker1",
      "keepalive": true,
      "nicehash": false
    }
  ],
  "randomx": {
    "init": -1,
    "mode": "light",
    "1gb-pages": false,
    "numa": true
  }
}
```

Run with: `xmrig --config=config.json`

---

## 📊 Monitoring

### Proxy Logs

The proxy prints statistics every 30 seconds:

```
📊 Stats: Miners=3/5 Shares=145/12/157 Blocks=1 Hashrate=12450.50 H/s Uptime=2h15m30s
           ↑    ↑      ↑    ↑    ↑       ↑       ↑                         ↑
         Active Total Valid Inv Total  Blocks   Network                  Uptime
```

### Per-Miner Stats

```
✅ Miner 192.168.1.100:12345 logged in: 0x1234...abcd (xmrig/6.21.0)
✅ Valid share from 192.168.1.100:12345 (diff: 10000)
📤 Share from 192.168.1.100:12345: job=42 nonce=12345678 result=0xabcd...
```

---

## 🔗 Protocol Flow

### 1. Miner Connects

```
xmrig → Proxy: {"method": "login", "params": {"login": "0x123...", "agent": "xmrig/6.21.0"}}
Proxy → xmrig: {"result": {"id": "miner1", "job": {...}, "status": "OK"}}
```

### 2. Proxy Gets Work

```
Proxy → Geth: POST /  {"method": "randomx_getWork", "params": []}
Geth → Proxy: {"result": ["0xheader", "0xseed", "0xtarget", "0xblocknum"]}
```

### 3. Proxy Converts to Stratum Job

```
Ethereum Work:
  headerHash: 0x1234...abcd (32 bytes)
  seedHash:   0xef56...7890 (32 bytes) ← Epoch seed!
  target:     0x0000...ffff
  blockNum:   0x820 (2080)

Stratum Job:
  {
    "job_id": "42",
    "blob": "1234...abcd0000000000000000",  ← header + nonce space
    "target": "ffff0000...",
    "algo": "rx/0",
    "height": 2080,
    "seed_hash": "0xef56...7890"  ← Miner uses this for RandomX cache
  }
```

### 4. Miner Submits Share

```
xmrig → Proxy: {"method": "submit", "params": {"job_id": "42", "nonce": "12345678", "result": "0xhash"}}
Proxy → Geth: POST /  {"method": "randomx_submitWork", "params": ["0x12345678", "0x1234...abcd", "0xhash"]}
Geth → Proxy: {"result": true}
Proxy → xmrig: {"result": {"status": "OK"}}
```

---

## 🔒 Security Considerations

### Firewall

```bash
# Allow Stratum port from miners
sudo ufw allow 3333/tcp

# Restrict Geth RPC to localhost (or proxy server)
# In geth: --http.addr "127.0.0.1" --http.api "eth,randomx"
```

### DDoS Protection

The proxy implements:
- Connection timeouts (5 minutes idle)
- Per-miner rate limiting
- Invalid share tracking
- Automatic bad miner disconnection

### Pool Operator Security

If running a public pool:
- Use `--pool-addr` to control rewards
- Set reasonable `--pool-fee`
- Monitor for unusually high invalid shares
- Implement payout system separately

---

## 🐛 Troubleshooting

### xmrig Can't Connect

**Symptom:** `socket error`

**Solutions:**
1. Check proxy is running: `netstat -tlnp | grep 3333`
2. Check firewall: `sudo ufw status`
3. Test connection: `telnet PROXY_IP 3333`

### High Invalid Shares

**Symptom:** Many `❌ Invalid share` in logs

**Solutions:**
1. **Wrong seed:** Ensure Geth and proxy are on same epoch
2. **Stale jobs:** Reduce work update interval
3. **Network latency:** Increase submission timeout
4. **Wrong algo:** Verify `--algo rx/0` in both proxy and xmrig

### No Work Available

**Symptom:** `No work available yet`

**Solutions:**
1. Check Geth is mining: `curl -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' http://localhost:8545`
2. Check RPC is exposed: `--http.api "eth,randomx"`
3. Verify proxy can reach Geth: `curl http://localhost:8545`

### Miner Disconnects Frequently

**Symptom:** Miners connect then disconnect after few seconds

**Solutions:**
1. Check difficulty isn't too high for miner hashrate
2. Increase keepalive interval in xmrig
3. Check network stability
4. Verify RandomX mode (light vs full)

---

## 📈 Performance Tuning

### Difficulty Adjustment

The proxy auto-adjusts difficulty to target ~1 share per 30 seconds per miner.

**Manual adjustment:**
```bash
# Low hashrate miners (< 1 KH/s)
--diff 5000

# Medium hashrate (1-10 KH/s)
--diff 10000

# High hashrate (> 10 KH/s)
--diff 50000
```

### xmrig Tuning

```bash
# Use all CPU cores
xmrig --threads=$(nproc)

# Enable huge pages (Linux)
sudo sysctl -w vm.nr_hugepages=1280

# Set CPU priority
nice -n -20 xmrig ...

# Optimize for your CPU
xmrig --cpu-max-threads-hint=100 --cpu-priority=5
```

---

## 🔬 Development

### Build from Source

```bash
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros/stratum-proxy
go build -o stratum-proxy
```

### Run Tests

```bash
go test -v ./...
```

### Enable Verbose Logging

```bash
./stratum-proxy -v
```

---

## 📚 References

- **Stratum Protocol:** https://en.bitcoin.it/wiki/Stratum_mining_protocol
- **xmrig Documentation:** https://xmrig.com/docs
- **Ducros RandomX:** ../RANDOMX-EPOCH-SCHEDULE.md
- **Monero Stratum:** https://github.com/xmrig/xmrig-proxy

---

## 🆘 Support

- **Issues:** https://github.com/Aqui-oi/go-Ducros/issues
- **Documentation:** ../DEPLOYMENT-GUIDE.md

---

## 📝 License

MIT License - See ../LICENSE

---

**Status:** ✅ Production Ready
**Compatibility:** xmrig 6.18.0+, SRBMiner-MULTI, XMRig-NVIDIA
**Network:** Ducros RandomX (ChainID 9999)
