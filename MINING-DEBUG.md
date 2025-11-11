# Mining Debug Guide

## Issue

Mining is reported as active (`eth_mining: true`) but no blocks are being produced even with ultra-low difficulty `0x1`.

## Diagnostic Logging Added

I've added comprehensive logging to help diagnose the issue:

1. **Miner Loop Logging** (`miner/miner.go`):
   - Mining loop start/stop events
   - Current block and difficulty
   - Work generation status
   - Seal operation initiation
   - Block insertion success/failure
   - Timeout warnings

2. **RandomX Consensus Logging** (`consensus/randomx/randomx.go`):
   - Seal function calls
   - RandomX cache initialization
   - VM creation success/failure
   - Mining attempts progress (every 100k attempts)
   - Solution found events
   - Abort/stop signals

## Steps to Diagnose

### 1. Pull Latest Changes and Rebuild

```bash
# Pull the logging changes
git pull origin claude/geth-randomx-pow-fork-011CV1zCZx1k45jWEf7eXxMT

# Stop mining
sudo ./manage-geth-service.sh stop

# Rebuild Geth
make geth

# Start again
sudo ./manage-geth-service.sh start
./manage-geth-service.sh start-mining 4
```

### 2. Watch the Logs in Real-Time

```bash
# Follow all logs
sudo journalctl -u geth-randomx -f

# OR filter for mining-related logs only
sudo journalctl -u geth-randomx -f | grep -E "Mining|RandomX|Seal|mine|block"
```

### 3. What to Look For

The logs should show this sequence if mining is working correctly:

```
INFO Mining loop started
INFO Mining new block parent=0 difficulty=1
INFO RandomX Seal called block=1 difficulty=1
INFO Initializing RandomX cache
INFO Starting RandomX mining goroutine
INFO RandomX mine starting block=1 difficulty=1 target=<large number>
INFO RandomX VM created successfully, starting nonce search...
INFO ✅ Found valid nonce! block=1 nonce=12345 attempts=42
INFO Solution found! block=1
INFO Block sealed successfully! number=1 hash=0x...
INFO 🎉 Successfully mined block! number=1 hash=0x...
```

### 4. Common Issues to Check

#### A. Mining Loop Not Starting
If you don't see `"Mining loop started"` (4 times for 4 threads):
- Mining wasn't actually started via RPC
- Try: `./manage-geth-service.sh start-mining 4`

#### B. Cache Initialization Failing
If you see `"Failed to initialize RandomX cache"`:
- RandomX library issue
- Check that RandomX is properly installed
- Try rebuilding: `make geth`

#### C. VM Creation Failing
If you see `"Failed to create RandomX VM!"`:
- Not enough memory for RandomX
- Try reducing threads: `./manage-geth-service.sh stop-mining && ./manage-geth-service.sh start-mining 1`

#### D. Seal Timeout
If you see `"Sealing timeout after 30 seconds"`:
- RandomX mining is too slow (shouldn't happen with difficulty 0x1)
- OR mining goroutine is stuck/crashed
- Look for errors before the timeout

#### E. Block Insertion Failing
If you see `"Failed to insert block"`:
- Check the error message
- May be a consensus validation issue
- Could be a synchronization problem between threads

### 5. Advanced Diagnostics

#### Check if generateWork is Failing

If logs show mining loop starts but no "RandomX Seal called":
- The `generateWork` function is failing silently
- Check for "Failed to generate work" errors

#### Check Block Difficulty Calculation

If mining starts but never finds a solution:
- Check what difficulty is being reported in logs
- Should be `difficulty=1` for genesis difficulty `0x1`
- If it's higher, the difficulty calculation is wrong

#### Check RandomX Target Calculation

The target should be huge with difficulty 1:
```
target = 2^256 / difficulty = 2^256 / 1 = 2^256
```

This means ANY hash should be valid. If you see mining attempts but no solution with difficulty 1, there's a bug in the hash comparison.

### 6. Quick Test Commands

```bash
# Check if mining status
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545

# Check block number (should increase if mining works)
watch -n 2 'curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}" http://localhost:8545 | jq'

# Check for errors in last 100 log lines
sudo journalctl -u geth-randomx -n 100 | grep -i error

# Count how many mining loops started (should be 4 for 4 threads)
sudo journalctl -u geth-randomx | grep "Mining loop started" | wc -l
```

## Expected Behavior with Difficulty 0x1

With difficulty set to `0x1`:
- **First attempt should almost always succeed** (probability ~100%)
- **First block should mine in < 1 second**
- **Subsequent blocks should mine continuously**

If this doesn't happen, there's a bug in:
1. The RandomX hashing (hashRandomX function)
2. The target calculation (should be maxUint256 / 1)
3. The hash-to-target comparison
4. The block difficulty calculation

## Reporting Back

Please run the diagnostics and share:

1. **The first 200 lines of logs after starting mining:**
   ```bash
   sudo journalctl -u geth-randomx -n 200
   ```

2. **Any ERROR or WARN messages:**
   ```bash
   sudo journalctl -u geth-randomx | grep -E "ERROR|WARN"
   ```

3. **Count of mining loops started:**
   ```bash
   sudo journalctl -u geth-randomx | grep "Mining loop started" | wc -l
   ```

4. **Evidence of seal attempts:**
   ```bash
   sudo journalctl -u geth-randomx | grep "RandomX Seal called" | head -10
   ```

This will help identify exactly where the mining process is failing!
