#!/bin/bash

# Start mining via RPC after Geth is ready
# Usage: start-mining.sh [threads]

THREADS=${1:-4}
MAX_RETRIES=30
RETRY_DELAY=2

echo "Waiting for Geth to be ready..."

# Wait for Geth to be responsive
for i in $(seq 1 $MAX_RETRIES); do
    if curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' \
        http://localhost:8545 >/dev/null 2>&1; then
        echo "Geth is ready!"
        break
    fi

    if [ $i -eq $MAX_RETRIES ]; then
        echo "ERROR: Geth did not become ready after $MAX_RETRIES attempts"
        exit 1
    fi

    echo "Waiting for Geth... (attempt $i/$MAX_RETRIES)"
    sleep $RETRY_DELAY
done

# Start mining with specified number of threads
echo "Starting mining with $THREADS threads..."

RESULT=$(curl -s -X POST -H "Content-Type: application/json" \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"miner_start\",\"params\":[$THREADS],\"id\":1}" \
    http://localhost:8545)

if echo "$RESULT" | grep -q '"result":null'; then
    echo "Mining started successfully with $THREADS threads!"
    exit 0
else
    echo "Failed to start mining: $RESULT"
    exit 1
fi
