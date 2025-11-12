#!/bin/bash
# deploy-stratum-proxy.sh - Deploy Ducros Stratum Proxy for xmrig mining

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

question() {
    echo -e "${BLUE}[?]${NC} $1"
}

echo "========================================="
echo "  Ducros Stratum Proxy Deployment"
echo "========================================="
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    error "Go is not installed!"
    echo "Install Go 1.21+ from: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
info "Go version: $GO_VERSION"

# Check Geth connection
question "Geth RPC URL [http://localhost:8545]: "
read GETH_RPC
if [ -z "$GETH_RPC" ]; then
    GETH_RPC="http://localhost:8545"
fi

info "Testing Geth connection..."
if curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    "$GETH_RPC" > /dev/null 2>&1; then
    info "✅ Geth connection successful"
else
    warn "⚠️  Could not connect to Geth. Make sure it's running and RPC is enabled."
fi

# Configure proxy
question "Stratum listen address [0.0.0.0:3333]: "
read STRATUM_ADDR
if [ -z "$STRATUM_ADDR" ]; then
    STRATUM_ADDR="0.0.0.0:3333"
fi

question "Initial difficulty [10000]: "
read INITIAL_DIFF
if [ -z "$INITIAL_DIFF" ]; then
    INITIAL_DIFF="10000"
fi

question "Pool mode? (y/n) [n]: "
read POOL_MODE

POOL_ADDR=""
POOL_FEE=""
if [ "$POOL_MODE" = "y" ]; then
    question "Pool payout address: "
    read POOL_ADDR

    question "Pool fee % [1.0]: "
    read POOL_FEE
    if [ -z "$POOL_FEE" ]; then
        POOL_FEE="1.0"
    fi
fi

question "Enable verbose logging? (y/n) [n]: "
read VERBOSE
VERBOSE_FLAG=""
if [ "$VERBOSE" = "y" ]; then
    VERBOSE_FLAG="-v"
fi

echo ""
info "Configuration:"
echo "  - Stratum: $STRATUM_ADDR"
echo "  - Geth RPC: $GETH_RPC"
echo "  - Initial Difficulty: $INITIAL_DIFF"
if [ ! -z "$POOL_ADDR" ]; then
    echo "  - Pool Address: $POOL_ADDR"
    echo "  - Pool Fee: $POOL_FEE%"
fi
echo ""

# Build proxy
info "Building Stratum proxy..."
cd stratum-proxy
go build -o ../build/stratum-proxy -ldflags="-s -w" .
cd ..

if [ ! -f "build/stratum-proxy" ]; then
    error "Build failed!"
    exit 1
fi

info "✅ Build successful"

# Create systemd service (optional)
question "Install as systemd service? (y/n) [n]: "
read INSTALL_SERVICE

if [ "$INSTALL_SERVICE" = "y" ]; then
    info "Creating systemd service..."

    WORKING_DIR=$(pwd)

    sudo tee /etc/systemd/system/stratum-proxy.service > /dev/null <<EOF
[Unit]
Description=Ducros Stratum Proxy
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$WORKING_DIR
ExecStart=$WORKING_DIR/build/stratum-proxy \
    --stratum "$STRATUM_ADDR" \
    --geth "$GETH_RPC" \
    --diff $INITIAL_DIFF \
    $([ ! -z "$POOL_ADDR" ] && echo "--pool-addr $POOL_ADDR") \
    $([ ! -z "$POOL_FEE" ] && echo "--pool-fee $POOL_FEE") \
    $VERBOSE_FLAG
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable stratum-proxy

    info "✅ Systemd service installed"
    info "Start with: sudo systemctl start stratum-proxy"
    info "Status: sudo systemctl status stratum-proxy"
    info "Logs: sudo journalctl -u stratum-proxy -f"
else
    # Create run script
    info "Creating run script..."

    cat > run-stratum-proxy.sh <<EOF
#!/bin/bash
./build/stratum-proxy \
    --stratum "$STRATUM_ADDR" \
    --geth "$GETH_RPC" \
    --diff $INITIAL_DIFF \
    $([ ! -z "$POOL_ADDR" ] && echo "--pool-addr $POOL_ADDR") \
    $([ ! -z "$POOL_FEE" ] && echo "--pool-fee $POOL_FEE") \
    $VERBOSE_FLAG
EOF

    chmod +x run-stratum-proxy.sh
    info "✅ Run script created: ./run-stratum-proxy.sh"
fi

# Firewall configuration
question "Configure firewall? (y/n) [y]: "
read CONFIGURE_FW
if [ "$CONFIGURE_FW" != "n" ]; then
    STRATUM_PORT=$(echo $STRATUM_ADDR | cut -d':' -f2)

    info "Opening port $STRATUM_PORT..."
    if command -v ufw &> /dev/null; then
        sudo ufw allow $STRATUM_PORT/tcp
        info "✅ Firewall configured (ufw)"
    elif command -v firewall-cmd &> /dev/null; then
        sudo firewall-cmd --permanent --add-port=$STRATUM_PORT/tcp
        sudo firewall-cmd --reload
        info "✅ Firewall configured (firewalld)"
    else
        warn "⚠️  No supported firewall found. Manually open port $STRATUM_PORT"
    fi
fi

# Create xmrig example config
info "Creating xmrig example config..."
PROXY_IP=$(hostname -I | awk '{print $1}')
cat > xmrig-config-ducros.json <<EOF
{
    "autosave": true,
    "cpu": true,
    "opencl": false,
    "cuda": false,
    "pools": [
        {
            "algo": "rx/0",
            "coin": "monero",
            "url": "$PROXY_IP:3333",
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
    },
    "cpu": {
        "enabled": true,
        "huge-pages": true,
        "max-threads-hint": 100
    },
    "donate-level": 0
}
EOF

info "✅ xmrig config created: xmrig-config-ducros.json"

echo ""
echo "========================================="
echo "  DEPLOYMENT COMPLETE!"
echo "========================================="
echo ""

if [ "$INSTALL_SERVICE" = "y" ]; then
    info "Start the proxy:"
    echo "  sudo systemctl start stratum-proxy"
    echo ""
    info "Check status:"
    echo "  sudo systemctl status stratum-proxy"
    echo ""
    info "View logs:"
    echo "  sudo journalctl -u stratum-proxy -f"
else
    info "Start the proxy:"
    echo "  ./run-stratum-proxy.sh"
fi

echo ""
info "Connect xmrig:"
echo "  1. Edit xmrig-config-ducros.json"
echo "  2. Replace YOUR_DUCROS_WALLET_ADDRESS with your wallet"
echo "  3. Run: xmrig --config=xmrig-config-ducros.json"
echo ""
echo "Or use command line:"
echo "  xmrig -o $PROXY_IP:3333 -u YOUR_DUCROS_ADDRESS -p worker1 --algo rx/0 --coin monero"
echo ""

info "Documentation: stratum-proxy/README.md"
info "Proxy is listening on: $STRATUM_ADDR"
echo ""
info "🎉 Ready to mine Ducros with xmrig!"
