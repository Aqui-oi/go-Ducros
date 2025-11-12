#!/bin/bash
# configure-firewall.sh - Configuration automatique du firewall pour Ducros Network

set -e

# Couleurs
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
echo "  Ducros Network - Firewall Configuration"
echo "========================================="
echo ""

# Vérifier si ufw est installé
if ! command -v ufw &> /dev/null; then
    error "ufw (Uncomplicated Firewall) is not installed"
    echo ""
    echo "Install with:"
    echo "  sudo apt-get update"
    echo "  sudo apt-get install ufw"
    echo ""
    exit 1
fi

# Vérifier les privilèges root
if [ "$EUID" -ne 0 ]; then
    error "This script must be run as root (sudo)"
    echo ""
    echo "Run with: sudo ./configure-firewall.sh"
    echo ""
    exit 1
fi

# Type de node
question "What type of node is this?"
echo "  1) Bootnode (accepts incoming P2P + RPC)"
echo "  2) Miner (public - accepts incoming P2P)"
echo "  3) Miner (private - no incoming connections)"
echo "  4) RPC Node (public RPC + P2P)"
echo ""
echo -n "Choice [1-4]: "
read NODE_TYPE

case $NODE_TYPE in
    1)
        NODE_NAME="Bootnode"
        PORTS_TCP="30303 8545"
        PORTS_UDP="30303"
        ;;
    2)
        NODE_NAME="Public Miner"
        PORTS_TCP="30303"
        PORTS_UDP="30303"
        ;;
    3)
        NODE_NAME="Private Miner"
        PORTS_TCP=""
        PORTS_UDP=""
        ;;
    4)
        NODE_NAME="Public RPC Node"
        PORTS_TCP="30303 8545"
        PORTS_UDP="30303"
        ;;
    *)
        error "Invalid choice"
        exit 1
        ;;
esac

echo ""
info "Configuring firewall for: $NODE_NAME"
echo ""

# Réinitialiser UFW
warn "This will reset UFW to default settings"
question "Continue? (y/n): "
read CONTINUE

if [ "$CONTINUE" != "y" ]; then
    error "Aborted by user"
    exit 1
fi

info "Resetting UFW..."
ufw --force reset

# Définir les règles par défaut
info "Setting default policies..."
ufw default deny incoming
ufw default allow outgoing

# Autoriser SSH (IMPORTANT!)
info "Allowing SSH (port 22)..."
ufw allow 22/tcp

# Autoriser les ports spécifiques au node
if [ ! -z "$PORTS_TCP" ]; then
    for PORT in $PORTS_TCP; do
        info "Allowing TCP port $PORT..."
        ufw allow $PORT/tcp
    done
fi

if [ ! -z "$PORTS_UDP" ]; then
    for PORT in $PORTS_UDP; do
        info "Allowing UDP port $PORT..."
        ufw allow $PORT/udp
    done
fi

# Options de sécurité supplémentaires
question "Enable additional security rules? (recommended) (y/n): "
read SECURITY

if [ "$SECURITY" = "y" ]; then
    info "Enabling additional security..."

    # Limiter les tentatives SSH
    info "Rate limiting SSH connections..."
    ufw limit 22/tcp

    # Bloquer les pings (optionnel)
    question "Block ping (ICMP)? (y/n): "
    read BLOCK_PING
    if [ "$BLOCK_PING" = "y" ]; then
        info "Blocking ICMP echo requests..."
        ufw deny proto icmp
    fi

    # Logging
    question "Enable firewall logging? (y/n): "
    read ENABLE_LOG
    if [ "$ENABLE_LOG" = "y" ]; then
        info "Enabling logging..."
        ufw logging on
    fi
fi

# Activer UFW
info "Enabling UFW..."
ufw --force enable

echo ""
echo "========================================="
echo "  FIREWALL CONFIGURED SUCCESSFULLY"
echo "========================================="
echo ""
echo "Node Type: $NODE_NAME"
echo ""
echo "Allowed Ports:"
if [ ! -z "$PORTS_TCP" ]; then
    echo "  TCP: 22 (SSH), $PORTS_TCP"
else
    echo "  TCP: 22 (SSH)"
fi
if [ ! -z "$PORTS_UDP" ]; then
    echo "  UDP: $PORTS_UDP"
fi
echo ""
echo "Status:"
ufw status verbose
echo ""
info "Firewall configuration complete! ✓"
echo ""
warn "IMPORTANT: Verify SSH still works before closing this session!"
echo "Open a new terminal and test: ssh user@this-server"
echo ""
