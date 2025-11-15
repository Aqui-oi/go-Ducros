# xmrig Integration Guide - Ducros Network

**Complete guide to mine Ducros with xmrig**

---

## 🎯 Overview

This guide shows you how to mine **Ducros Network** using **xmrig** (the popular Monero/RandomX miner).

**Architecture:**
```
xmrig (Stratum) ←→ Stratum Proxy ←→ go-Ducros (JSON-RPC)
```

The **Stratum Proxy** translates between xmrig's protocol and Ducros blockchain.

---

## 📋 Prerequisites

### On Proxy Server

✅ **Go 1.21+** - For building the proxy
✅ **go-Ducros node** - Running with mining enabled
✅ **Open port 3333** - For Stratum connections

### On Miner Machine(s)

✅ **xmrig 6.18.0+** - Download from https://xmrig.com/download
✅ **Network access** - To proxy server port 3333

---

## 🚀 Quick Start (3 Steps)

### Step 1: Start go-Ducros Node

```bash
# On your node server
cd ~/go-Ducros

./build/bin/geth \
  --datadir ./data \
  --http \
  --http.addr "0.0.0.0" \
  --http.api "eth,randomx,net,web3" \
  --mine \
  --miner.threads 0 \
  --miner.etherbase 0xYourAddress
```

**Note:** `--miner.threads 0` disables local mining (xmrig will mine instead)

### Step 2: Deploy Stratum Proxy

```bash
# On your proxy server (can be same as node)
cd ~/go-Ducros

chmod +x deploy-stratum-proxy.sh
./deploy-stratum-proxy.sh
```

Answer the prompts:
- Geth RPC: `http://localhost:8545` (or your node IP)
- Stratum address: `0.0.0.0:3333`
- Initial difficulty: `10000`
- Pool mode: `n` (for solo mining)

The script will:
✅ Build the proxy
✅ Configure firewall
✅ Create xmrig config
✅ Optionally install as systemd service

### Step 3: Run xmrig

```bash
# On your miner machine
xmrig \
  -o PROXY_IP:3333 \
  -u YOUR_DUCROS_WALLET_ADDRESS \
  -p worker1 \
  --algo rx/0 \
  --coin monero
```

**That's it!** xmrig will start mining Ducros.

---

## 📊 Verify Everything Works

### 1. Check Node is Running

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545

# Expected: {"result": true}
```

### 2. Check Proxy Logs

```bash
# If installed as service:
sudo journalctl -u stratum-proxy -f

# If running manually:
tail -f stratum-proxy.log
```

**Expected output:**
```
✅ RPC connection verified
📦 New job 1: block 2145, seed 0xabcd...
🔌 New connection from 192.168.1.100:54321
✅ Miner 192.168.1.100:54321 logged in: 0x1234... (xmrig/6.21.0)
✅ Valid share from 192.168.1.100:54321 (diff: 10000)
```

### 3. Check xmrig Output

```
[2025-11-12 12:34:56.789]  net      use pool PROXY_IP:3333
[2025-11-12 12:34:56.890]  net      new job from PROXY_IP:3333 diff 10000 algo rx/0
[2025-11-12 12:34:57.000]  cpu      use profile rx
[2025-11-12 12:35:10.123]  cpu      accepted (1/0) diff 10000 (2450 H/s)
```

**"accepted"** means shares are being submitted successfully! ✅

---

## 🔧 Configuration

### Stratum Proxy Options

| Option | Description | Example |
|--------|-------------|---------|
| `--stratum` | Listen address | `0.0.0.0:3333` |
| `--geth` | Geth RPC URL | `http://localhost:8545` |
| `--diff` | Initial difficulty | `10000` |
| `--pool-addr` | Pool payout address | `0xYourPoolAddress` |
| `--pool-fee` | Pool fee % | `1.0` |
| `-v` | Verbose logging | (flag) |

### xmrig Options

**Command Line:**
```bash
xmrig \
  -o PROXY_IP:3333 \              # Proxy address
  -u YOUR_WALLET \                # Your Ducros address
  -p worker1 \                    # Worker name
  --algo rx/0 \                   # RandomX algorithm
  --coin monero \                 # Use Monero protocol
  --threads=$(nproc) \            # Use all CPU cores
  --randomx-mode light \          # Light mode (< 2GB RAM)
  --cpu-max-threads-hint=100 \    # Use 100% of threads
  --donate-level=0                # No donations
```

**Config File** (`config.json`):
```json
{
  "pools": [{
    "algo": "rx/0",
    "coin": "monero",
    "url": "PROXY_IP:3333",
    "user": "YOUR_DUCROS_WALLET",
    "pass": "worker1"
  }],
  "randomx": {
    "mode": "light",
    "1gb-pages": false
  },
  "cpu": {
    "enabled": true,
    "huge-pages": true,
    "max-threads-hint": 100
  }
}
```

Run with: `xmrig --config=config.json`

---

## 💡 Performance Optimization

### 1. Enable Huge Pages (Linux)

**Temporary:**
```bash
sudo sysctl -w vm.nr_hugepages=1280
```

**Permanent:**
```bash
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### 2. MSR Mod (Advanced)

```bash
# Install msr-tools
sudo apt-get install msr-tools

# Run xmrig with --randomx-wrmsr=1
xmrig --randomx-wrmsr=1 ...
```

**Warning:** May require root. Use with caution.

### 3. CPU Priority

```bash
# Run xmrig with higher priority
sudo nice -n -20 xmrig ...
```

### 4. Multiple Workers

Run multiple xmrig instances:
```bash
# Worker 1
xmrig -o PROXY:3333 -u WALLET -p worker1 &

# Worker 2
xmrig -o PROXY:3333 -u WALLET -p worker2 &
```

---

## 🏊 Pool vs Solo Mining

### Solo Mining (Recommended for Start)

```bash
# Proxy config
./stratum-proxy --stratum "0.0.0.0:3333" --geth "http://localhost:8545"

# Each miner uses their own address
xmrig -o PROXY:3333 -u 0xMiner1Address ...
xmrig -o PROXY:3333 -u 0xMiner2Address ...
```

**Pros:**
- No pool fees
- Direct rewards to wallet
- Full control

**Cons:**
- Variable rewards (luck-based)
- Need to run own node

### Pool Mining

```bash
# Proxy config
./stratum-proxy \
  --stratum "0.0.0.0:3333" \
  --geth "http://localhost:8545" \
  --pool-addr "0xPoolAddress" \
  --pool-fee 1.0

# All miners use pool address
xmrig -o POOL_PROXY:3333 -u 0xPoolAddress -p miner1id ...
```

**Pros:**
- Consistent rewards
- Lower variance
- Professional operation

**Cons:**
- Pool fees (1-2%)
- Trust in pool operator
- Need separate payout system

---

## 🔍 Monitoring

### Proxy Statistics

The proxy logs stats every 30 seconds:
```
📊 Stats: Miners=3/5 Shares=145/12/157 Blocks=1 Hashrate=12450.50 H/s Uptime=2h15m
```

- **Miners:** 3 active out of 5 total
- **Shares:** 145 valid / 12 invalid / 157 total
- **Blocks:** 1 block found
- **Hashrate:** 12.45 KH/s network total

### xmrig Hashrate

```
[2025-11-12 12:45:00.000]  cpu      speed 10s/60s/15m 2450.0 2448.5 2449.2 H/s max 2500.0 H/s
```

- **10s:** 2450 H/s (instant)
- **60s:** 2448.5 H/s (1 minute avg)
- **15m:** 2449.2 H/s (15 minute avg)

### Check Blockchain

```bash
# On the node
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq

# Check last block miner
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://localhost:8545 | jq '.result.miner'
```

---

## 🐛 Troubleshooting

### Issue: xmrig Can't Connect

**Error:** `socket error`

**Fix:**
1. Check proxy is running: `ps aux | grep stratum-proxy`
2. Check port is open: `sudo netstat -tlnp | grep 3333`
3. Test connectivity: `telnet PROXY_IP 3333`
4. Check firewall: `sudo ufw status`

### Issue: All Shares Invalid

**Error:** `❌ Invalid share` in proxy logs

**Fix:**
1. **Wrong epoch seed:**
   - Proxy and Geth must be synced
   - Check epoch: see RANDOMX-EPOCH-SCHEDULE.md
2. **Wrong algorithm:**
   - Verify `--algo rx/0` in xmrig
   - Verify proxy is using `rx/0`
3. **Stale jobs:**
   - Reduce work update interval in proxy
   - Check network latency

### Issue: Low Hashrate

**Fix:**
1. **Enable huge pages:** (see Performance section)
2. **Check CPU usage:** `htop` - should be near 100%
3. **Disable other programs:** Close browsers, etc.
4. **Use all threads:** `--threads=$(nproc)`
5. **Check temperatures:** Ensure CPU not throttling

### Issue: No Work from Proxy

**Error:** `No work available yet`

**Fix:**
1. Check Geth is mining: `eth_mining` should return `true`
2. Check RPC is accessible from proxy server
3. Verify `--http.api` includes `randomx` or `eth`
4. Check proxy logs for RPC errors

---

## 📈 Expected Performance

### Hashrate by CPU

| CPU | Cores | Hashrate | Difficulty |
|-----|-------|----------|------------|
| i5-8400 | 6 | ~2 KH/s | 10,000 |
| Ryzen 5 3600 | 6 | ~5 KH/s | 20,000 |
| Ryzen 7 5800X | 8 | ~9 KH/s | 40,000 |
| Ryzen 9 5950X | 16 | ~18 KH/s | 80,000 |
| Threadripper 3970X | 32 | ~35 KH/s | 150,000 |

### Share Submission Rate

**Target:** ~1 share per 30 seconds per miner

- **Too many shares** (< 10s): Proxy auto-increases difficulty
- **Too few shares** (> 60s): Proxy auto-decreases difficulty

---

## 🔐 Security Best Practices

### For Proxy Operators

1. **Restrict RPC access:**
   ```bash
   # In geth: only allow localhost
   --http.addr "127.0.0.1"
   ```

2. **Firewall rules:**
   ```bash
   # Only allow miners
   sudo ufw allow from MINER_IP to any port 3333
   ```

3. **Monitor for abuse:**
   - Watch for unusually high invalid shares
   - Implement IP banning for bad actors
   - Log all connections

### For Miners

1. **Use your own proxy** - Don't trust unknown pools
2. **Verify rewards** - Check blockchain for your blocks
3. **Monitor uptime** - Ensure proxy is reliable

---

## 📚 Additional Resources

- **Stratum Proxy README:** stratum-proxy/README.md
- **Epoch Schedule:** RANDOMX-EPOCH-SCHEDULE.md
- **Mining API:** MINING-API.md
- **Deployment Guide:** DEPLOYMENT-GUIDE.md
- **xmrig Documentation:** https://xmrig.com/docs
- **RandomX Spec:** https://github.com/tevador/RandomX

---

## 🎉 Success Checklist

- [ ] go-Ducros node running with `--mine`
- [ ] RPC accessible on port 8545
- [ ] Stratum proxy running on port 3333
- [ ] Firewall allows port 3333
- [ ] xmrig connected and showing "accepted" shares
- [ ] Proxy logs show valid shares
- [ ] Blocks increasing on blockchain

**If all checked:** You're mining Ducros with xmrig! 🚀

---

**Status:** ✅ Production Ready
**Compatibility:** xmrig 6.18.0+, SRBMiner, XMRig-NVIDIA/AMD
**Network:** Ducros RandomX (ChainID 9999)
**Branch:** `claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi`
