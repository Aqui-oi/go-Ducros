#!/bin/bash

# Script to reset blockchain with low difficulty genesis

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "Resetting Blockchain with Low Difficulty"
echo "========================================="
echo ""

# Stop mining
echo "1. Stopping Geth service..."
sudo ./manage-geth-service.sh stop || true
sleep 2

# Remove old blockchain data
echo ""
echo "2. Removing old blockchain data..."
rm -rf "${SCRIPT_DIR}/data-randomx"
echo "   ✓ Old data removed"

# Verify genesis has difficulty 0x1
echo ""
echo "3. Checking genesis difficulty..."
DIFFICULTY=$(grep -oP '"difficulty":\s*"\K[^"]+' "${SCRIPT_DIR}/genesis-randomx.json")
if [ "$DIFFICULTY" != "0x1" ]; then
    echo "   ERROR: Genesis difficulty is $DIFFICULTY, not 0x1!"
    echo "   Please check genesis-randomx.json"
    exit 1
fi
echo "   ✓ Genesis difficulty is 0x1"

# Rebuild Geth with VM fix
echo ""
echo "4. Rebuilding Geth with RandomX VM fix..."
make geth
echo "   ✓ Geth rebuilt"

# Initialize with new genesis
echo ""
echo "5. Initializing blockchain with low-difficulty genesis..."
"${SCRIPT_DIR}/build/bin/geth" init --datadir "${SCRIPT_DIR}/data-randomx" "${SCRIPT_DIR}/genesis-randomx.json"
echo "   ✓ Blockchain initialized"

# Start service
echo ""
echo "6. Starting Geth service..."
sudo ./manage-geth-service.sh start
sleep 3

# Start mining
echo ""
echo "7. Starting mining with 4 threads..."
./manage-geth-service.sh start-mining 4
sleep 2

echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "With difficulty 0x1, blocks should mine INSTANTLY!"
echo ""
echo "Watch the logs:"
echo "  sudo journalctl -u geth-randomx -f"
echo ""
echo "Check mining status:"
echo "  ./manage-geth-service.sh mining-info"
echo ""
