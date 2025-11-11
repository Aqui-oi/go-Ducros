#!/bin/bash

# Geth RandomX Service Management Script
# This script helps install, start, stop, and monitor the Geth RandomX service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="geth-randomx"
SERVICE_FILE="${SCRIPT_DIR}/geth-randomx.service"
SYSTEM_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"
LOG_DIR="${SCRIPT_DIR}/logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored message
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "This command requires root privileges. Please run with sudo."
        exit 1
    fi
}

# Install the service
install_service() {
    print_info "Installing Geth RandomX service..."

    # Create log directory
    mkdir -p "${LOG_DIR}"
    chown ubuntu:ubuntu "${LOG_DIR}"

    # Copy service file
    cp "${SERVICE_FILE}" "${SYSTEM_SERVICE_PATH}"
    print_info "Service file copied to ${SYSTEM_SERVICE_PATH}"

    # Reload systemd
    systemctl daemon-reload
    print_info "Systemd configuration reloaded"

    # Enable service
    systemctl enable "${SERVICE_NAME}"
    print_info "Service enabled (will start on boot)"

    print_info "Installation complete!"
    print_info "Use 'sudo $(basename $0) start' to start the service"
}

# Uninstall the service
uninstall_service() {
    print_warning "Uninstalling Geth RandomX service..."

    # Stop service if running
    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        systemctl stop "${SERVICE_NAME}"
        print_info "Service stopped"
    fi

    # Disable service
    systemctl disable "${SERVICE_NAME}" 2>/dev/null || true

    # Remove service file
    rm -f "${SYSTEM_SERVICE_PATH}"

    # Reload systemd
    systemctl daemon-reload

    print_info "Service uninstalled successfully"
}

# Start the service
start_service() {
    print_info "Starting Geth RandomX service..."
    systemctl start "${SERVICE_NAME}"
    sleep 2
    systemctl status "${SERVICE_NAME}" --no-pager
}

# Stop the service
stop_service() {
    print_info "Stopping Geth RandomX service..."
    systemctl stop "${SERVICE_NAME}"
    print_info "Service stopped"
}

# Restart the service
restart_service() {
    print_info "Restarting Geth RandomX service..."
    systemctl restart "${SERVICE_NAME}"
    sleep 2
    systemctl status "${SERVICE_NAME}" --no-pager
}

# Show service status
status_service() {
    systemctl status "${SERVICE_NAME}" --no-pager
}

# Show service logs
logs_service() {
    if [ "$1" == "follow" ] || [ "$1" == "-f" ]; then
        journalctl -u "${SERVICE_NAME}" -f
    else
        journalctl -u "${SERVICE_NAME}" -n 100 --no-pager
    fi
}

# Show mining info
mining_info() {
    print_info "Fetching mining information..."

    # Check if service is running
    if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
        print_error "Service is not running"
        exit 1
    fi

    # Query via HTTP RPC
    echo ""
    echo "=== Mining Status ==="
    curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
        http://localhost:8545 | jq -r '.result' && echo "Mining: Active" || echo "Mining: Inactive"

    echo ""
    echo "=== Hashrate ==="
    HASHRATE=$(curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
        http://localhost:8545 | jq -r '.result')
    HASHRATE_DEC=$((16#${HASHRATE#0x}))
    echo "Hashrate: ${HASHRATE_DEC} H/s"

    echo ""
    echo "=== Block Number ==="
    BLOCK=$(curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        http://localhost:8545 | jq -r '.result')
    BLOCK_DEC=$((16#${BLOCK#0x}))
    echo "Current Block: ${BLOCK_DEC}"

    echo ""
    echo "=== Coinbase Balance ==="
    curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x0000000000000000000000000000000000000000","latest"],"id":1}' \
        http://localhost:8545 | jq -r '.result'
}

# Set mining address
set_coinbase() {
    if [ -z "$1" ]; then
        print_error "Please provide an Ethereum address"
        echo "Usage: $0 set-coinbase <ethereum_address>"
        exit 1
    fi

    print_info "Setting coinbase address to: $1"

    # Update service file
    sed -i "s/--miner.etherbase \".*\"/--miner.etherbase \"$1\"/" "${SYSTEM_SERVICE_PATH}"

    # Reload and restart
    systemctl daemon-reload
    systemctl restart "${SERVICE_NAME}"

    print_info "Coinbase address updated. Service restarted."
}

# Show help
show_help() {
    cat << EOF
Geth RandomX Service Management Script

Usage: sudo $0 [command]

Commands:
    install         Install and enable the systemd service
    uninstall       Stop and remove the systemd service
    start           Start the service
    stop            Stop the service
    restart         Restart the service
    status          Show service status
    logs            Show recent logs
    logs -f         Follow logs in real-time
    mining-info     Show mining information (hashrate, blocks, etc.)
    set-coinbase    Set mining reward address

Examples:
    sudo $0 install
    sudo $0 start
    sudo $0 logs -f
    sudo $0 mining-info
    sudo $0 set-coinbase 0xYourEthereumAddress

Note: Most commands require sudo/root privileges.
EOF
}

# Main script logic
case "${1:-}" in
    install)
        check_root
        install_service
        ;;
    uninstall)
        check_root
        uninstall_service
        ;;
    start)
        check_root
        start_service
        ;;
    stop)
        check_root
        stop_service
        ;;
    restart)
        check_root
        restart_service
        ;;
    status)
        status_service
        ;;
    logs)
        logs_service "$2"
        ;;
    mining-info)
        mining_info
        ;;
    set-coinbase)
        check_root
        set_coinbase "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: ${1:-}"
        echo ""
        show_help
        exit 1
        ;;
esac
