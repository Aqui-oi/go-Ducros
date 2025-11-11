# Lowering Genesis Difficulty for Fast Testing

## What Changed

The genesis difficulty has been lowered from `0x20000` (131,072) to `0x1` (1) in `genesis-randomx.json`. This will make mining almost instant for testing purposes.

## Steps to Apply the Change

### 1. Stop Mining Service (if running)
```bash
sudo ./manage-geth-service.sh stop
# OR if systemd isn't available:
# pkill -9 geth
```

### 2. Backup Old Data (Optional)
```bash
# If you want to keep the old blockchain data:
mv data-randomx data-randomx.backup-$(date +%Y%m%d-%H%M%S)
```

### 3. Remove Old Blockchain Data
```bash
# Remove the old blockchain data to start fresh:
rm -rf data-randomx/
```

### 4. Rebuild Geth (if needed)
```bash
# Make sure you have the latest code:
make geth
```

### 5. Initialize with New Genesis
```bash
./build/bin/geth init --datadir data-randomx genesis-randomx.json
```

Expected output:
```
INFO [MM-DD|HH:MM:SS.mmm] Successfully wrote genesis state
```

### 6. Start Mining
```bash
# If using systemd service:
sudo ./manage-geth-service.sh start

# OR manually:
./build/bin/geth \
  --datadir data-randomx \
  --networkid 33669 \
  --http --http.addr "0.0.0.0" --http.port 8545 \
  --http.api "eth,net,web3,miner,admin,debug" \
  --miner.etherbase "0x797c03e9994d77e7d3f2a84ad33857dec585c3a7" \
  --verbosity 3 &

# Wait a few seconds for Geth to start, then:
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"miner_start","params":[4],"id":1}' \
  http://localhost:8545
```

### 7. Verify Mining
```bash
# Check block number (should increase rapidly):
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545

# Check if mining:
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545
```

## Expected Results

With difficulty set to `0x1`:
- **First block**: Should mine within seconds
- **Subsequent blocks**: Should mine continuously every few seconds
- **Block time**: Extremely fast (1-5 seconds per block)

## Important Notes

### ⚠️ This is for LOCAL TESTING ONLY

**DO NOT use difficulty `0x1` for production/mainnet!**

This ultra-low difficulty is ONLY for:
- ✅ Verifying that mining works
- ✅ Testing transaction processing
- ✅ Testing smart contracts
- ✅ Development and debugging

### For Production Mainnet

Before launching a public mainnet, you need to:

1. **Set Appropriate Difficulty**
   - Consider your target block time (e.g., 15 seconds)
   - Set difficulty to achieve that block time with expected hashrate
   - Typical values: `0x4000` to `0x100000` depending on network size

2. **Test Thoroughly**
   - Run private testnet for weeks/months
   - Test network synchronization
   - Test under various load conditions
   - Perform security audits

3. **Set Up Infrastructure**
   - Deploy bootnodes
   - Set up block explorers
   - Create documentation
   - Build community tools

4. **Security Considerations**
   - Audit all consensus code
   - Test 51% attack resistance
   - Test network partition scenarios
   - Review economic incentives

## Troubleshooting

### Mining but blocks aren't appearing
```bash
# Check logs:
sudo journalctl -u geth-randomx -f

# OR:
tail -f logs/geth.log
```

### "Genesis mismatch" error
- You forgot to remove old data
- Run: `rm -rf data-randomx/` and reinitialize

### Can't build Geth
```bash
# Make sure you have build dependencies:
make geth

# If CGo errors, install RandomX library:
# (should already be set up from previous build)
```

## Next Steps After First Block

Once you successfully mine several blocks with difficulty `0x1`:

1. **Verify functionality**:
   - Create transactions
   - Deploy contracts
   - Test RPC APIs

2. **Adjust difficulty** (for more realistic testing):
   - Stop mining
   - Edit `genesis-randomx.json`
   - Change `"difficulty": "0x1"` to something higher (e.g., `"0x4000"`)
   - Remove data and reinitialize
   - Restart

3. **Plan your network**:
   - Decide on target block time
   - Calculate appropriate difficulty
   - Design token economics
   - Plan distribution strategy

## Quick Command Summary

```bash
# Stop, reset, and start with new genesis:
sudo ./manage-geth-service.sh stop
rm -rf data-randomx/
./build/bin/geth init --datadir data-randomx genesis-randomx.json
sudo ./manage-geth-service.sh start
./manage-geth-service.sh start-mining 4

# Monitor:
./manage-geth-service.sh mining-info
sudo ./manage-geth-service.sh logs -f
```
