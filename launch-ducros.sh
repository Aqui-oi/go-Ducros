#!/bin/bash

# 🚀 Script de Lancement Rapide Ducros Network
# Usage: ./launch-ducros.sh [options]

set -e

# Couleurs pour output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration par défaut
DATADIR="${DATADIR:-$HOME/.ducros}"
NETWORKID=9999
HTTP_PORT=8545
WS_PORT=8546
P2P_PORT=30303
MINING_THREADS="${MINING_THREADS:-4}"
VERBOSITY=3

# Détection automatique
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GETH_BIN="$SCRIPT_DIR/build/bin/geth"
GENESIS_FILE="$SCRIPT_DIR/genesis-production.json"

print_banner() {
    echo -e "${BLUE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║        🚀 Ducros Network - RandomX PoW Chain 🚀         ║
║                                                          ║
║               ChainID: 9999                              ║
║        Consensus: RandomX (CPU-friendly)                 ║
║               Block Time: ~13s                           ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

print_step() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[i]${NC} $1"
}

check_prerequisites() {
    print_info "Vérification des prérequis..."

    # Check if RandomX library is installed
    if ldconfig -p 2>/dev/null | grep -q randomx; then
        print_step "RandomX library trouvée"
    else
        print_error "RandomX library NON trouvée !"
        echo ""
        echo "Installez-la avec:"
        echo "  git clone https://github.com/tevador/RandomX.git /tmp/RandomX"
        echo "  cd /tmp/RandomX && mkdir build && cd build"
        echo "  cmake -DARCH=native -DBUILD_SHARED_LIBS=ON .."
        echo "  make -j\$(nproc) && sudo make install && sudo ldconfig"
        exit 1
    fi

    # Check huge pages
    HUGEPAGES=$(grep HugePages_Total /proc/meminfo | awk '{print $2}')
    if [ "$HUGEPAGES" -ge 1280 ]; then
        print_step "Huge pages activées ($HUGEPAGES pages)"
    else
        print_warn "Huge pages NON configurées (performance -15×)"
        echo ""
        echo "Activez-les avec: sudo sysctl -w vm.nr_hugepages=1280"
        echo ""
        read -p "Continuer quand même? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    # Check if geth is compiled
    if [ ! -f "$GETH_BIN" ]; then
        print_error "Geth non compilé !"
        echo ""
        echo "Compilez avec: make geth"
        exit 1
    fi
    print_step "Geth trouvé: $GETH_BIN"

    # Check genesis file
    if [ ! -f "$GENESIS_FILE" ]; then
        print_error "Fichier genesis non trouvé: $GENESIS_FILE"
        exit 1
    fi
    print_step "Genesis trouvé: $GENESIS_FILE"
}

initialize_datadir() {
    if [ -d "$DATADIR/geth/chaindata" ]; then
        print_step "Datadir déjà initialisé: $DATADIR"
        return 0
    fi

    print_info "Initialisation du datadir..."
    mkdir -p "$DATADIR"

    "$GETH_BIN" --datadir "$DATADIR" init "$GENESIS_FILE"

    if [ $? -eq 0 ]; then
        print_step "Genesis initialisé avec succès"
    else
        print_error "Échec de l'initialisation du genesis"
        exit 1
    fi
}

create_account() {
    if [ -n "$(ls -A $DATADIR/keystore 2>/dev/null)" ]; then
        print_step "Comptes existants trouvés"
        FIRST_ACCOUNT=$("$GETH_BIN" --datadir "$DATADIR" account list 2>/dev/null | head -1 | grep -oP '0x[a-fA-F0-9]{40}')
        print_info "Adresse par défaut: $FIRST_ACCOUNT"
        return 0
    fi

    print_info "Aucun compte trouvé. Création d'un nouveau compte..."
    echo ""
    echo "⚠️  SAUVEGARDEZ VOTRE MOT DE PASSE !"
    echo ""

    "$GETH_BIN" --datadir "$DATADIR" account new

    FIRST_ACCOUNT=$("$GETH_BIN" --datadir "$DATADIR" account list 2>/dev/null | head -1 | grep -oP '0x[a-fA-F0-9]{40}')
    print_step "Compte créé: $FIRST_ACCOUNT"
}

show_menu() {
    echo ""
    echo -e "${BLUE}Mode de lancement:${NC}"
    echo ""
    echo "  1) Nœud simple (pas de mining)"
    echo "  2) Nœud + Mining CPU intégré"
    echo "  3) Nœud pour mining externe (xmrig via Stratum)"
    echo "  4) Testnet local (difficulty faible)"
    echo ""
    read -p "Choisissez le mode [1-4]: " MODE_CHOICE
    echo ""
}

launch_node_only() {
    print_info "Lancement du nœud sans mining..."

    exec "$GETH_BIN" \
        --datadir "$DATADIR" \
        --networkid $NETWORKID \
        --port $P2P_PORT \
        --http \
        --http.addr "127.0.0.1" \
        --http.port $HTTP_PORT \
        --http.api "eth,net,web3,txpool,randomx" \
        --http.corsdomain "*" \
        --ws \
        --ws.addr "127.0.0.1" \
        --ws.port $WS_PORT \
        --ws.api "eth,net,web3,txpool,randomx" \
        --verbosity $VERBOSITY \
        --log.rotate \
        --log.maxage 7
}

launch_node_with_mining() {
    if [ -z "$FIRST_ACCOUNT" ]; then
        print_error "Aucun compte trouvé pour recevoir les récompenses"
        exit 1
    fi

    print_info "Lancement du nœud avec mining CPU..."
    print_info "Mining address: $FIRST_ACCOUNT"
    print_info "Mining threads: $MINING_THREADS"

    exec "$GETH_BIN" \
        --datadir "$DATADIR" \
        --networkid $NETWORKID \
        --port $P2P_PORT \
        --http \
        --http.addr "127.0.0.1" \
        --http.port $HTTP_PORT \
        --http.api "eth,net,web3,txpool,randomx,miner" \
        --http.corsdomain "*" \
        --mine \
        --miner.threads $MINING_THREADS \
        --miner.etherbase "$FIRST_ACCOUNT" \
        --verbosity $VERBOSITY \
        --log.rotate \
        --log.maxage 7
}

launch_node_for_external_mining() {
    if [ -z "$FIRST_ACCOUNT" ]; then
        print_error "Aucun compte trouvé pour recevoir les récompenses"
        exit 1
    fi

    print_warn "⚠️  Mode RPC public - NE PAS EXPOSER À INTERNET"
    print_info "Lancement du nœud pour mining externe..."
    print_info "Mining address: $FIRST_ACCOUNT"
    print_info "RPC disponible sur: http://0.0.0.0:$HTTP_PORT"

    echo ""
    echo "Lancez ensuite le Stratum proxy:"
    echo "  cd stratum-proxy"
    echo "  ./stratum-proxy --geth-rpc http://localhost:$HTTP_PORT"
    echo ""

    exec "$GETH_BIN" \
        --datadir "$DATADIR" \
        --networkid $NETWORKID \
        --port $P2P_PORT \
        --http \
        --http.addr "0.0.0.0" \
        --http.port $HTTP_PORT \
        --http.api "eth,net,web3,txpool,randomx,miner" \
        --http.corsdomain "*" \
        --miner.etherbase "$FIRST_ACCOUNT" \
        --verbosity $VERBOSITY \
        --log.rotate \
        --log.maxage 7
}

launch_testnet_local() {
    if [ -z "$FIRST_ACCOUNT" ]; then
        print_error "Aucun compte trouvé pour recevoir les récompenses"
        exit 1
    fi

    print_info "Lancement testnet local (difficulty faible)..."
    print_warn "Mode TEST - blocks rapides pour développement"

    # Create low difficulty genesis if needed
    if [ ! -f "$DATADIR/genesis-test.json" ]; then
        cat > "$DATADIR/genesis-test.json" << EOF
{
  "config": {
    "chainId": 9999,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "randomx": {
      "lwmaActivationBlock": 0
    }
  },
  "difficulty": "0x10",
  "gasLimit": "0x7a1200",
  "baseFeePerGas": "0x3b9aca00",
  "alloc": {
    "$FIRST_ACCOUNT": {
      "balance": "0x200000000000000000000000"
    }
  }
}
EOF
        rm -rf "$DATADIR/geth"
        "$GETH_BIN" --datadir "$DATADIR" init "$DATADIR/genesis-test.json"
    fi

    exec "$GETH_BIN" \
        --datadir "$DATADIR" \
        --networkid 9999 \
        --port $P2P_PORT \
        --http \
        --http.addr "127.0.0.1" \
        --http.port $HTTP_PORT \
        --http.api "eth,net,web3,txpool,randomx,miner,debug,personal" \
        --http.corsdomain "*" \
        --mine \
        --miner.threads 1 \
        --miner.etherbase "$FIRST_ACCOUNT" \
        --nodiscover \
        --verbosity 4 \
        --allow-insecure-unlock
}

# ============================================
# MAIN
# ============================================

print_banner

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --datadir)
            DATADIR="$2"
            shift 2
            ;;
        --threads)
            MINING_THREADS="$2"
            shift 2
            ;;
        --mode)
            MODE_CHOICE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --datadir PATH    Datadir path (default: ~/.ducros)"
            echo "  --threads N       Mining threads (default: 4)"
            echo "  --mode N          Launch mode 1-4 (default: interactive)"
            echo "  --help            Show this help"
            echo ""
            exit 0
            ;;
        *)
            print_error "Option inconnue: $1"
            exit 1
            ;;
    esac
done

# Run checks
check_prerequisites

# Initialize if needed
initialize_datadir

# Create/load account
create_account

# Show menu if mode not specified
if [ -z "$MODE_CHOICE" ]; then
    show_menu
fi

# Launch based on mode
case $MODE_CHOICE in
    1)
        launch_node_only
        ;;
    2)
        launch_node_with_mining
        ;;
    3)
        launch_node_for_external_mining
        ;;
    4)
        launch_testnet_local
        ;;
    *)
        print_error "Mode invalide: $MODE_CHOICE"
        exit 1
        ;;
esac
