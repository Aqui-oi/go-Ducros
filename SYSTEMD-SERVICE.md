# Geth RandomX Systemd Service

This guide explains how to run your Geth RandomX node as a systemd service that starts automatically and runs in the background.

## Quick Start

```bash
# 1. Pull the latest changes
git pull origin claude/geth-randomx-pow-fork-011CV1zCZx1k45jWEf7eXxMT

# 2. Make the management script executable
chmod +x manage-geth-service.sh

# 3. Set your mining address (IMPORTANT: replace with your actual address!)
sudo nano geth-randomx.service
# Edit the line: --miner.etherbase "0xYourEthereumAddressHere"

# 4. Install and start the service
sudo ./manage-geth-service.sh install
sudo ./manage-geth-service.sh start

# 5. Check if it's running
sudo ./manage-geth-service.sh status

# 6. Watch the logs (mining will start automatically after 5 seconds)
sudo ./manage-geth-service.sh logs -f
```

## Service Management Commands

### Installation
```bash
# Install the service (enables auto-start on boot)
sudo ./manage-geth-service.sh install
```

### Start/Stop/Restart
```bash
# Start the service
sudo ./manage-geth-service.sh start

# Stop the service
sudo ./manage-geth-service.sh stop

# Restart the service
sudo ./manage-geth-service.sh restart
```

### Monitoring
```bash
# Check service status
sudo ./manage-geth-service.sh status

# View recent logs
sudo ./manage-geth-service.sh logs

# Follow logs in real-time (Ctrl+C to exit)
sudo ./manage-geth-service.sh logs -f

# Show mining information (hashrate, blocks, balance)
./manage-geth-service.sh mining-info
```

### Mining Control
```bash
# Start mining with 4 threads (default)
./manage-geth-service.sh start-mining

# Start mining with 8 threads
./manage-geth-service.sh start-mining 8

# Stop mining (node keeps running)
./manage-geth-service.sh stop-mining
```

### Configuration
```bash
# Change mining reward address (requires service restart)
sudo ./manage-geth-service.sh set-coinbase 0xYourNewAddress

# Change mining threads (requires service restart)
sudo ./manage-geth-service.sh set-threads 8
```

## Service Configuration

The service file is located at: `geth-randomx.service`

### Key Settings:

```ini
# Mining settings
--miner.etherbase "0x..." # Address to receive mining rewards
# Mining starts automatically via start-mining.sh script after 5 seconds
# Default: 4 threads (configurable in ExecStartPost line)

# Network settings
--networkid 33669         # Custom network ID

# HTTP RPC API
--http                    # Enable HTTP-RPC server
--http.addr "0.0.0.0"     # Listen on all interfaces
--http.port 8545          # HTTP-RPC port
--http.api "eth,net,web3,miner,admin,debug"

# WebSocket API
--ws                      # Enable WS-RPC server
--ws.addr "0.0.0.0"       # Listen on all interfaces
--ws.port 8546            # WS-RPC port
```

**Note on Mining**: Mining starts automatically 5 seconds after Geth starts, using the `start-mining.sh` script. This is configured in the service file's `ExecStartPost` directive.

### Adjusting Mining Threads

#### Option 1: Using the management script (recommended)
```bash
sudo ./manage-geth-service.sh set-threads 8
```

#### Option 2: Manual configuration
1. Edit the service file:
```bash
sudo nano /etc/systemd/system/geth-randomx.service
```

2. Modify the `ExecStartPost` line:
```ini
ExecStartPost=/home/ubuntu/go-Ducros/start-mining.sh 8  # Use 8 threads instead of 4
```

3. Reload and restart:
```bash
sudo systemctl daemon-reload
sudo ./manage-geth-service.sh restart
```

#### Option 3: Control mining without restarting
```bash
# Stop current mining
./manage-geth-service.sh stop-mining

# Start with different thread count
./manage-geth-service.sh start-mining 8
```

## Monitoring Mining Activity

### Real-time Log Monitoring
```bash
# Watch logs for mining activity
sudo journalctl -u geth-randomx -f | grep -E "mined|block|hash"
```

### Check Mining Statistics
```bash
# Get current mining info
./manage-geth-service.sh mining-info
```

### Using RPC API Directly
```bash
# Check if mining
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545

# Get hashrate
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545

# Get current block number
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545
```

## Troubleshooting

### Service won't start
```bash
# Check for errors
sudo systemctl status geth-randomx
sudo journalctl -u geth-randomx -n 50
```

### High CPU usage
- This is normal for mining! RandomX uses CPU for proof-of-work
- Reduce `--miner.threads` to lower CPU usage
- Or stop mining: edit service file and remove `--mine` flag

### Service crashes or restarts
- Check logs: `sudo journalctl -u geth-randomx -n 100`
- The service will automatically restart after 10 seconds
- Check system resources (RAM, disk space)

### Change mining address after installation
```bash
sudo ./manage-geth-service.sh set-coinbase 0xYourNewAddress
```

## Uninstallation

To completely remove the service:
```bash
# Stop and uninstall
sudo ./manage-geth-service.sh uninstall

# Optional: Remove data directory
rm -rf data-randomx/
```

## Firewall Configuration

If you want other nodes to connect to you:
```bash
# Allow P2P connections
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# Allow HTTP RPC (only if needed externally - security risk!)
# sudo ufw allow 8545/tcp

# Allow WebSocket (only if needed externally - security risk!)
# sudo ufw allow 8546/tcp
```

**Security Warning**: Only expose RPC ports (8545, 8546) if you absolutely need external access and understand the security implications.

## Performance Tips

1. **Optimize for your CPU**: RandomX performs best with:
   - At least 2GB RAM per mining thread
   - Modern CPU with AES-NI support
   - 2MB L3 cache per thread

2. **Thread count**: Use `--miner.threads N` where N is:
   - Recommended: Number of physical cores - 1
   - Maximum: Number of logical cores

3. **Monitor system load**:
```bash
# Check CPU usage
htop

# Check memory usage
free -h
```

## Logs Location

- **Systemd logs**: `sudo journalctl -u geth-randomx`
- **Geth log file**: `~/go-Ducros/logs/geth.log`

## Auto-start on Boot

The service is configured to start automatically on system boot. To disable:
```bash
sudo systemctl disable geth-randomx
```

To re-enable:
```bash
sudo systemctl enable geth-randomx
```

## Support

For issues or questions:
- Check logs first: `sudo ./manage-geth-service.sh logs`
- Check mining status: `./manage-geth-service.sh mining-info`
- GitHub: https://github.com/Aqui-oi/go-Ducros
