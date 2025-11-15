# 🏊 Ducros Network - Pool Operator Guide

## Production Mining Pool Setup

**Version:** 1.0
**Network:** Ducros Mainnet (ChainID: 9999)
**Algorithm:** RandomX (CPU-optimized)
**Target Block Time:** 13 seconds

---

## 📋 Table of Contents

1. [Requirements](#requirements)
2. [Architecture](#architecture)
3. [Stratum Proxy Setup](#stratum-proxy-setup)
4. [Pool Backend Integration](#pool-backend-integration)
5. [Miner Configuration](#miner-configuration)
6. [Performance Optimization](#performance-optimization)
7. [Monitoring & Maintenance](#monitoring--maintenance)
8. [Troubleshooting](#troubleshooting)

---

## 🔧 Requirements

### Hardware (Pool Server)

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **CPU** | 4 cores | 8+ cores |
| **RAM** | 8 GB | 16 GB |
| **Disk** | 200 GB SSD | 500 GB NVMe |
| **Network** | 100 Mbps | 1 Gbps |
| **Bandwidth** | Unmetered | Unmetered |

### Software Dependencies

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y \
  build-essential \
  cmake \
  git \
  golang-1.24 \
  redis-server \
  postgresql-14

# Install RandomX library
git clone https://github.com/tevador/RandomX.git
cd RandomX && mkdir build && cd build
cmake -DARCH=native -DBUILD_SHARED_LIBS=ON ..
make -j$(nproc) && sudo make install
sudo ldconfig
```

---

## 🏗️ Architecture

### High-Level Architecture

```
                                   ┌─────────────────┐
                                   │  Ducros Node    │
                                   │  (Full Node)    │
                                   │  Port: 8545     │
                                   └────────┬────────┘
                                            │ JSON-RPC
                                            │
                          ┌─────────────────┴─────────────────┐
                          │                                   │
                 ┌────────▼────────┐              ┌──────────▼──────────┐
                 │ Stratum Proxy 1 │              │  Stratum Proxy 2    │
                 │  Port: 3333     │              │   Port: 3334        │
                 └────────┬────────┘              └──────────┬──────────┘
                          │                                  │
         ┌────────────────┼──────────────────┬──────────────┴─────┐
         │                │                  │                    │
    ┌────▼───┐       ┌────▼───┐        ┌────▼───┐          ┌────▼───┐
    │ Miner 1│       │ Miner 2│        │ Miner 3│          │ Miner N│
    │ xmrig  │       │ xmrig  │        │ xmrig  │          │  ...   │
    └────────┘       └────────┘        └────────┘          └────────┘
         ▲                ▲                 ▲                    ▲
         │                │                 │                    │
         └────────────────┴─────────────────┴────────────────────┘
                              Share submission
                              Difficulty adjustment
```

### Components

1. **Ducros Full Node** - Blockchain validation + block creation
2. **Stratum Proxy** - Translates between xmrig and Geth
3. **Redis** - Share tracking and statistics
4. **PostgreSQL** - Payment processing and history
5. **Pool Backend** - Payment distribution (optional)

---

## 🌐 Stratum Proxy Setup

### Installation

```bash
# Clone Ducros repository
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros/stratum-proxy

# Build Stratum proxy
make build

# Or build manually
go build -o stratum-proxy .
```

### Configuration

Create `config.json`:

```json
{
  "stratum": {
    "listen": "0.0.0.0:3333",
    "difficulty": 10000,
    "varDiff": {
      "enabled": true,
      "minDiff": 1000,
      "maxDiff": 1000000,
      "targetTime": 30,
      "retargetTime": 120,
      "variancePercent": 30
    }
  },
  "geth": {
    "rpc": "http://localhost:8545",
    "timeout": 10
  },
  "pool": {
    "address": "0xYourPoolPayoutAddress",
    "fee": 1.0,
    "coinbase": "Ducros Pool"
  },
  "redis": {
    "host": "localhost:6379",
    "password": "",
    "db": 0,
    "poolSize": 10
  },
  "logging": {
    "level": "info",
    "file": "/var/log/stratum-proxy.log"
  }
}
```

### Systemd Service

Create `/etc/systemd/system/stratum-proxy.service`:

```ini
[Unit]
Description=Ducros Stratum Proxy
After=network.target redis.service

[Service]
Type=simple
User=ducros
WorkingDirectory=/opt/ducros/stratum-proxy
ExecStart=/opt/ducros/stratum-proxy/stratum-proxy -config config.json
Restart=always
RestartSec=10
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log

[Install]
WantedBy=multi-user.target
```

Start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable stratum-proxy
sudo systemctl start stratum-proxy
sudo systemctl status stratum-proxy
```

---

## 🔌 Pool Backend Integration

### Share Validation

```go
// Example: Validate share from miner
func (p *Pool) validateShare(share *Share) error {
    // 1. Check job exists
    job, exists := p.getJob(share.JobID)
    if !exists {
        return errors.New("job not found")
    }

    // 2. Calculate hash
    headerHash := job.HeaderHash
    nonce := share.Nonce
    mixDigest := calculateRandomX(headerHash, nonce, job.SeedHash)

    // 3. Verify difficulty
    hashValue := new(big.Int).SetBytes(mixDigest[:])
    target := difficultyToTarget(share.Difficulty)

    if hashValue.Cmp(target) > 0 {
        return errors.New("share above target")
    }

    // 4. Check if block solution
    networkTarget := difficultyToTarget(job.NetworkDifficulty)
    if hashValue.Cmp(networkTarget) <= 0 {
        // Submit block to Geth!
        return p.submitBlock(headerHash, nonce, mixDigest)
    }

    // Valid share
    return nil
}
```

### Payment Processing

```sql
-- Database schema for payments
CREATE TABLE shares (
    id SERIAL PRIMARY KEY,
    miner_address VARCHAR(42) NOT NULL,
    difficulty BIGINT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    valid BOOLEAN DEFAULT true,
    block_hash VARCHAR(66),
    reward NUMERIC(30, 18)
);

CREATE TABLE payouts (
    id SERIAL PRIMARY KEY,
    miner_address VARCHAR(42) NOT NULL,
    amount NUMERIC(30, 18) NOT NULL,
    tx_hash VARCHAR(66),
    timestamp TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending'
);

CREATE INDEX idx_shares_miner ON shares(miner_address);
CREATE INDEX idx_payouts_miner ON payouts(miner_address);
```

Payment calculation (PPLNS):

```python
def calculate_pplns_payout(block_reward, shares, N=10000):
    """
    Pay-Per-Last-N-Shares
    N = number of shares to consider (e.g., last 10,000 shares)
    """
    # Get last N shares before block was found
    recent_shares = get_last_n_shares(N)

    # Calculate total difficulty
    total_difficulty = sum(share.difficulty for share in recent_shares)

    # Calculate each miner's payout
    payouts = {}
    for share in recent_shares:
        miner = share.miner_address
        miner_share = (share.difficulty / total_difficulty) * block_reward

        if miner in payouts:
            payouts[miner] += miner_share
        else:
            payouts[miner] = miner_share

    # Deduct pool fee (e.g., 1%)
    pool_fee = 0.01
    for miner in payouts:
        payouts[miner] *= (1 - pool_fee)

    return payouts
```

---

## ⛏️ Miner Configuration

### xmrig Configuration

Distribute this config to your miners:

```json
{
  "autosave": true,
  "cpu": {
    "enabled": true,
    "huge-pages": true,
    "hw-aes": true,
    "priority": null,
    "asm": true,
    "max-threads-hint": 100
  },
  "pools": [
    {
      "algo": "rx/0",
      "coin": null,
      "url": "your-pool.com:3333",
      "user": "YOUR_WALLET_ADDRESS",
      "pass": "x",
      "rig-id": "worker1",
      "nicehash": false,
      "keepalive": true,
      "enabled": true,
      "tls": false,
      "daemon": false,
      "self-select": null
    }
  ],
  "randomx": {
    "init": -1,
    "mode": "auto",
    "1gb-pages": false,
    "numa": true
  },
  "api": {
    "enabled": true,
    "host": "127.0.0.1",
    "port": 8888,
    "access-token": null,
    "worker-id": null
  }
}
```

### SRBMiner Configuration

```ini
[General]
algorithm = randomx
pool = your-pool.com:3333
wallet = YOUR_WALLET_ADDRESS
worker = worker1
password = x

[CPU]
threads = 0
affinity_mode = 0
randomx_use_1gb_pages = false
```

---

## 🚀 Performance Optimization

### Stratum Proxy Tuning

```toml
# config.toml
[proxy]
max_connections = 10000
read_buffer_size = 4096
write_buffer_size = 4096
connection_timeout = 300

[difficulty]
# Start difficulty (lower = more shares, higher load)
initial = 10000
# Variance difficulty adjustment
variance_percent = 30
# Target share time (seconds)
target_time = 30
```

### Redis Optimization

```redis
# /etc/redis/redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### PostgreSQL Tuning

```sql
-- postgresql.conf
shared_buffers = 4GB
effective_cache_size = 12GB
maintenance_work_mem = 1GB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
work_mem = 64MB
min_wal_size = 1GB
max_wal_size = 4GB
max_worker_processes = 4
max_parallel_workers_per_gather = 2
max_parallel_workers = 4
```

### Network Optimization

```bash
# /etc/sysctl.conf
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_max_syn_backlog = 8192
```

---

## 📊 Monitoring & Maintenance

### Prometheus Metrics

Expose metrics from Stratum proxy:

```go
// metrics.go
var (
    totalShares = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pool_shares_total",
            Help: "Total number of shares submitted",
        },
        []string{"miner", "valid"},
    )

    currentHashrate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "pool_hashrate_current",
            Help: "Current pool hashrate",
        },
        []string{"miner"},
    )

    blocksFound = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "pool_blocks_found_total",
            Help: "Total blocks found by pool",
        },
    )
)
```

### Grafana Dashboard

Import dashboard from `ops/grafana-pool-dashboard.json`:

- Active miners count
- Total pool hashrate
- Shares per second
- Block find rate
- Payment queue status
- Network difficulty vs pool difficulty

### Health Checks

```bash
#!/bin/bash
# healthcheck.sh

# Check Stratum proxy
if ! nc -z localhost 3333; then
    echo "Stratum proxy down!" | mail -s "ALERT" admin@pool.com
    systemctl restart stratum-proxy
fi

# Check Geth node
if ! curl -sf http://localhost:8545 > /dev/null; then
    echo "Geth node down!" | mail -s "ALERT" admin@pool.com
fi

# Check Redis
if ! redis-cli ping > /dev/null; then
    echo "Redis down!" | mail -s "ALERT" admin@pool.com
fi
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Miners Can't Connect

```bash
# Check Stratum proxy is listening
netstat -tlnp | grep 3333

# Check firewall
sudo ufw status
sudo ufw allow 3333/tcp

# Check logs
tail -f /var/log/stratum-proxy.log
```

#### 2. High Reject Rate

```bash
# Check difficulty is appropriate
# Too high = few shares, too low = server overload
# Adjust in config.json:
{
  "stratum": {
    "difficulty": 10000  // Adjust based on miner hashrate
  }
}
```

#### 3. Blocks Not Submitting

```bash
# Check Geth connection
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545

# Check miner account has funds for gas
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0xYourAddress","latest"],"id":1}' \
  http://localhost:8545
```

#### 4. Payment System Issues

```sql
-- Check pending payouts
SELECT miner_address, SUM(amount) as total_owed
FROM payouts
WHERE status = 'pending'
GROUP BY miner_address;

-- Verify share accounting
SELECT miner_address, COUNT(*) as share_count, SUM(difficulty) as total_diff
FROM shares
WHERE timestamp > NOW() - INTERVAL '24 hours'
  AND valid = true
GROUP BY miner_address
ORDER BY total_diff DESC;
```

---

## 📚 Additional Resources

- **xmrig Documentation:** https://xmrig.com/docs
- **Ducros RPC API:** See `MINING-API.md`
- **Stratum Protocol:** See `stratum-proxy/README.md`
- **Community Support:** https://discord.gg/ducros

---

## 🤝 Community Pools

Register your pool in the community pool list:

```yaml
# Submit PR to: https://github.com/Aqui-oi/ducros-pools
name: "Your Pool Name"
url: "https://your-pool.com"
stratum: "your-pool.com:3333"
fee: 1.0
payout_scheme: "PPLNS"
min_payout: "0.1 DUC"
location: "US-East"
contact: "admin@your-pool.com"
```

---

**Last Updated:** 2025-11-12
**Maintainer:** Ducros Network Team
**License:** LGPL-3.0
