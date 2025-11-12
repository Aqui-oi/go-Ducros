# Geth Upstream Tracking & Merge Strategy

**Ducros Network Fork Information**

## Current Version

- **Base Version:** Geth v1.16.7
- **Fork Point:** go-ethereum v1.16.7-stable
- **Last Sync:** 2024-11 (initial fork creation)
- **Network:** Mainnet-ready (ChainID: 9999)

## Fork Modifications

### Core Consensus Changes

1. **RandomX Proof-of-Work** (`consensus/randomx/`)
   - Complete replacement of Ethash with RandomX
   - CPU-friendly, ASIC-resistant mining
   - Monero-compatible RandomX algorithm
   - Epoch-based cache rotation (2048 blocks, 64 block lag)

2. **LWMA Difficulty Adjustment** (`consensus/randomx/lwma.go`)
   - Linearly Weighted Moving Average difficulty algorithm
   - Optimized for CPU mining hashrate variance
   - 60-block window, 13-second target
   - Burst detection and damping

3. **Median-Time-Past Validation** (`consensus/randomx/consensus.go`)
   - 11-block median timestamp validation
   - Prevents timestamp manipulation attacks
   - Required for LWMA stability

### Modified Files

```
consensus/randomx/           - NEW: Complete RandomX consensus engine
├── randomx.go               - Core RandomX integration (C bindings)
├── consensus.go             - Consensus interface implementation
├── epoch.go                 - Epoch-based seed rotation
├── lwma.go                  - LWMA difficulty algorithm
├── api.go                   - Mining RPC APIs
├── *_test.go                - Integration and unit tests
stratum-proxy/               - NEW: Stratum mining proxy for xmrig
genesis-production.json      - Ducros Network genesis configuration
version/version.go           - Version: v1.16.7 (maintained for compatibility)
params/config.go             - Added RandomXConfig to ChainConfig
cmd/geth/config.go           - RandomX configuration flags
```

### Unmodified Core Components

- **EVM Execution:** 100% Ethereum-compatible (London fork)
- **State Management:** Unmodified MPT and state processing
- **P2P Networking:** Standard Geth networking stack
- **JSON-RPC:** All standard APIs + RandomX extensions
- **Transaction Pool:** Unmodified txpool implementation
- **Account Management:** Standard Ethereum accounts
- **Smart Contracts:** Solidity 0.8.x compatibility

## Upstream Merge Strategy

### Merge Policy

**Recommendation:** **Selective Merges** with Careful Testing

Ducros maintains a **conservative** merge policy due to consensus-critical changes:

1. **Security Patches:** Merge immediately (if not consensus-affecting)
2. **Performance Improvements:** Merge after testing
3. **Feature Additions:** Evaluate on case-by-case basis
4. **Consensus Changes:** **DO NOT MERGE** without community approval

### Merge Process

#### 1. Identify Upstream Changes

```bash
# Add upstream Geth as remote
git remote add upstream https://github.com/ethereum/go-ethereum.git
git fetch upstream

# View changes since fork point
git log v1.16.7..upstream/master --oneline

# Identify relevant commits
git log --oneline --grep="security\|fix\|perf" v1.16.7..upstream/master
```

#### 2. Categorize Changes

| Category | Examples | Merge Strategy |
|----------|----------|----------------|
| **Security** | CVE fixes, DoS patches | ✅ Merge ASAP |
| **Bug Fixes** | RPC fixes, state fixes | ✅ Merge with tests |
| **Performance** | Database optimizations | ✅ Merge after benchmarks |
| **Features** | New RPCs, tools | ⚠️ Evaluate carefully |
| **Consensus** | PoS, Shanghai, Cancun | ❌ DO NOT MERGE |
| **EVM** | New opcodes (PUSH0) | ⚠️ Requires fork coordination |

#### 3. Cherry-Pick Process

```bash
# Create merge branch
git checkout -b merge/upstream-security-fixes

# Cherry-pick specific commits (NOT ranges to avoid consensus changes)
git cherry-pick <commit-hash-1>
git cherry-pick <commit-hash-2>

# Resolve conflicts (prioritize Ducros consensus logic)
# Test extensively
make test
make integration-test

# Create PR for review
```

#### 4. Conflict Resolution Rules

**Priority Order:**
1. **RandomX consensus logic** - NEVER compromise
2. **LWMA difficulty** - NEVER compromise
3. **Epoch rotation** - NEVER compromise
4. **Upstream improvements** - Adapt to work with RandomX

**Example:**
```go
// WRONG - breaks consensus
func CalcDifficulty() {
    return upstream.NewDifficultyAlgorithm() // PoS or Ethash-based
}

// CORRECT - preserves RandomX
func CalcDifficulty() {
    if ShouldUseLWMA(config, blockNum) {
        return CalcDifficultyLWMA() // Ducros-specific
    }
    return upstream.FallbackAlgorithm()
}
```

### Files to NEVER Merge

These files contain Ducros-specific consensus logic and must not be overwritten:

```
consensus/randomx/**         - Complete consensus engine
genesis-production.json      - Ducros genesis
stratum-proxy/**             - Mining infrastructure
```

### Safe-to-Merge Areas

These areas can typically merge upstream changes safely:

```
rpc/**                       - RPC APIs (unless PoS-specific)
core/state/**                - State management
core/vm/**                   - EVM execution (pre-Shanghai)
p2p/**                       - Networking
crypto/**                    - Cryptography
accounts/**                  - Account management
cmd/utils/**                 - CLI utilities
internal/**                  - Internal utilities
```

### Risky Merge Areas (Requires Testing)

```
core/types/**                - Block/header structure changes
core/blockchain.go           - Chain processing logic
miner/**                     - Mining coordination
consensus/**                 - Consensus interfaces (may break RandomX)
```

## Testing Requirements

### Before Merging Upstream Changes

1. **Unit Tests:**
   ```bash
   make test                                    # All tests
   go test ./consensus/randomx/... -v           # RandomX tests
   go test -tags=integration ./consensus/...    # Integration tests
   ```

2. **Integration Tests:**
   ```bash
   # 3-node network test
   ./scripts/test-network.sh

   # Mining test
   ./geth --randomx.mining --mine
   ```

3. **Consensus Test Vectors:**
   ```bash
   go test -run TestConsensusVector ./consensus/randomx/
   ```

4. **Regression Testing:**
   - Verify epoch rotation at blocks 2048, 4096, 6144
   - Verify LWMA difficulty adjustment
   - Verify burst detection
   - Verify Median-Time-Past validation

## Upstream Version Tracking

### Recommended Update Frequency

- **Security Releases:** Within 1 week
- **Minor Versions:** Every 3-6 months (after thorough testing)
- **Major Versions:** Evaluate carefully (may require fork coordination)

### Version Compatibility Matrix

| Geth Version | Ducros Status | Notes |
|--------------|---------------|-------|
| v1.16.7 | ✅ Current Base | Stable, production-ready |
| v1.17.x | ⚠️ Evaluate | May include Shanghai (PUSH0) |
| v1.18.x | ❌ Not Compatible | Likely includes PoS merge |
| v1.19.x+ | ❌ Not Compatible | Full PoS, no PoW support |

### Breaking Changes to Watch

**Upstream changes that would break Ducros:**

1. **Proof-of-Stake (The Merge)**
   - Status: Will NOT merge
   - Reason: Incompatible with PoW consensus

2. **Shanghai Fork (EIP-3855 - PUSH0)**
   - Status: Can merge with coordination
   - Requires: Community approval + coordinated hardfork

3. **Cancun Fork (EIP-4844 - Blob transactions)**
   - Status: Evaluate benefits
   - May require: Significant adaptation

4. **Consensus interface changes**
   - Status: Review carefully
   - Action: Adapt RandomX to new interfaces

## Long-Term Maintenance

### Divergence Management

As Geth moves further into PoS, Ducros will diverge more. Strategy:

1. **Maintain Common EVM Core**
   - Keep EVM execution layer synchronized
   - Ensures smart contract compatibility

2. **Independent Consensus Layer**
   - RandomX consensus evolves independently
   - Follow Monero's RandomX updates if relevant

3. **Selective Feature Adoption**
   - Adopt non-consensus features (RPC, tooling)
   - Reject PoS-specific features

### Alternative: Rebase on Geth PoW Archive

If Geth removes PoW entirely, consider rebasing on:
- Ethereum Classic (maintains Ethash PoW)
- Last stable PoW Geth version (v1.10.x)
- Fork and maintain PoW codebase independently

## Community Coordination

### Hard Fork Process

If upstream changes require Ducros hard fork:

1. **Proposal Phase** (30 days)
   - Document proposed changes
   - Community discussion
   - Security audit

2. **Testing Phase** (60 days)
   - Testnet deployment
   - Miner testing
   - Exchange coordination

3. **Activation** (coordinated block height)
   - Announce activation block
   - Ensure >95% miner/node upgrade
   - Monitor network consensus

### Communication Channels

- **GitHub Issues:** Technical discussions
- **Discord/Telegram:** Community coordination
- **Documentation:** Update `EVM-COMPATIBILITY.md`

## Security Considerations

### Upstream Security Advisories

Monitor:
- https://github.com/ethereum/go-ethereum/security/advisories
- Geth security mailing list
- Ethereum security Discord

### Ducros-Specific Vulnerabilities

Watch for:
- RandomX cache poisoning
- Epoch rotation attacks
- LWMA difficulty manipulation
- Timestamp manipulation
- Mining API DoS

## Conclusion

**Merge Philosophy:**
*"Be conservative with consensus, liberal with features"*

Ducros maintains Ethereum compatibility at the EVM level while running independent PoW consensus. Upstream merges must preserve this balance.

**Key Principle:**
When in doubt, DO NOT merge consensus-affecting changes without:
1. Thorough review
2. Comprehensive testing
3. Community approval
4. Coordinated hardfork if needed

---

**Last Updated:** 2025-11-12
**Maintainer:** Ducros Core Team
**Review Frequency:** Quarterly
