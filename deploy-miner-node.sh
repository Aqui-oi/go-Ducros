#!/bin/bash
# deploy-miner-node.sh - Script pour déployer un node miner Ducros
# Ce node va miner des blocs et participer au consensus RandomX

set -e

echo "========================================="
echo "  Ducros Network - Miner Node Deployment"
echo "========================================="
echo ""

# Configuration par défaut
DATADIR="./miner-data"
NODE_PORT=30303
RPC_PORT=8545
MINER_THREADS=4

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages
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

# Vérifier que geth existe
if [ ! -f "./build/bin/geth" ]; then
    error "Geth binary not found! Please compile first with: make geth"
    exit 1
fi

# Vérifier que genesis existe
if [ ! -f "./genesis-production.json" ]; then
    error "genesis-production.json not found!"
    exit 1
fi

# Configuration interactive
echo "Configuration du nœud mineur:"
echo ""

question "Nom du nœud (ex: miner1, miner2): "
read NODE_NAME
if [ -z "$NODE_NAME" ]; then
    NODE_NAME="miner1"
fi
DATADIR="./data-$NODE_NAME"

question "Port P2P [$NODE_PORT]: "
read INPUT_PORT
if [ ! -z "$INPUT_PORT" ]; then
    NODE_PORT=$INPUT_PORT
fi

question "Port RPC [$RPC_PORT]: "
read INPUT_RPC
if [ ! -z "$INPUT_RPC" ]; then
    RPC_PORT=$INPUT_RPC
fi

question "Nombre de threads pour mining [$MINER_THREADS]: "
read INPUT_THREADS
if [ ! -z "$INPUT_THREADS" ]; then
    MINER_THREADS=$INPUT_THREADS
fi

question "Enode du bootnode (format: enode://...@IP:PORT): "
read BOOTNODE_ENODE

if [ -z "$BOOTNODE_ENODE" ]; then
    warn "Aucun bootnode spécifié. Le nœud fonctionnera en mode isolé."
    BOOTNODE_ARG=""
else
    BOOTNODE_ARG="--bootnodes \"$BOOTNODE_ENODE\""
fi

echo ""
info "Configuration du nœud:"
echo "  - Nom: $NODE_NAME"
echo "  - DataDir: $DATADIR"
echo "  - Port P2P: $NODE_PORT"
echo "  - Port RPC: $RPC_PORT"
echo "  - Threads mining: $MINER_THREADS"
echo "  - Bootnode: ${BOOTNODE_ENODE:-none}"
echo ""

# Étape 1: Initialiser le nœud si nécessaire
if [ ! -d "$DATADIR/geth" ]; then
    info "Initializing node with genesis..."
    ./build/bin/geth init --datadir "$DATADIR" genesis-production.json
    info "Node initialized successfully"
else
    warn "Node already initialized, skipping genesis init"
fi

# Étape 2: Créer un compte mineur
if [ ! -f "$DATADIR/miner-account.txt" ]; then
    info "Creating miner account..."
    question "Enter password for miner account: "
    read -s MINER_PASSWORD
    echo ""

    echo "$MINER_PASSWORD" > "$DATADIR/password.txt"
    chmod 600 "$DATADIR/password.txt"

    ACCOUNT=$(./build/bin/geth --datadir "$DATADIR" account new --password "$DATADIR/password.txt" 2>&1 | grep "Public address" | awk '{print $4}')
    echo "$ACCOUNT" > "$DATADIR/miner-account.txt"
    info "Miner account created: $ACCOUNT"
else
    ACCOUNT=$(cat "$DATADIR/miner-account.txt")
    info "Using existing miner account: $ACCOUNT"

    if [ ! -f "$DATADIR/password.txt" ]; then
        question "Enter password for account $ACCOUNT: "
        read -s MINER_PASSWORD
        echo ""
        echo "$MINER_PASSWORD" > "$DATADIR/password.txt"
        chmod 600 "$DATADIR/password.txt"
    fi
fi

# Étape 3: Obtenir l'IP publique (optionnel pour miners)
info "Detecting public IP..."
PUBLIC_IP=$(curl -s https://api.ipify.org 2>/dev/null || echo "127.0.0.1")
info "IP: $PUBLIC_IP"
echo ""

# Étape 4: Instructions firewall (optionnel pour miners)
question "Voulez-vous exposer ce nœud publiquement? (y/n): "
read EXPOSE_PUBLIC

if [ "$EXPOSE_PUBLIC" = "y" ]; then
    info "Configuration firewall recommandée:"
    echo ""
    echo "  sudo ufw allow $NODE_PORT/tcp"
    echo "  sudo ufw allow $NODE_PORT/udp"
    echo ""
    NAT_ARG="--nat \"extip:$PUBLIC_IP\""
    HTTP_ADDR="0.0.0.0"
else
    NAT_ARG=""
    HTTP_ADDR="127.0.0.1"
fi

# Étape 5: Créer le script de démarrage
info "Creating startup script..."

cat > "$DATADIR/start-miner.sh" <<EOF
#!/bin/bash
# Auto-generated miner startup script for: $NODE_NAME

./build/bin/geth \\
  --datadir "$DATADIR" \\
  --networkid 9999 \\
  --port $NODE_PORT \\
  --http \\
  --http.addr "$HTTP_ADDR" \\
  --http.port $RPC_PORT \\
  --http.api "eth,net,web3,randomx,miner,personal,admin" \\
  --http.corsdomain "*" \\
  --http.vhosts "*" \\
  $NAT_ARG \\
  --maxpeers 50 \\
  --netrestrict "" \\
  $BOOTNODE_ARG \\
  --mine \\
  --miner.threads $MINER_THREADS \\
  --miner.etherbase "$ACCOUNT" \\
  --unlock "$ACCOUNT" \\
  --password "$DATADIR/password.txt" \\
  --allow-insecure-unlock \\
  --verbosity 3 \\
  --syncmode "full" \\
  2>&1 | tee "$DATADIR/miner.log"
EOF

chmod +x "$DATADIR/start-miner.sh"

echo ""
echo "========================================="
echo "  MINER NODE READY TO START"
echo "========================================="
echo ""
echo "Node Information:"
echo "  - Name: $NODE_NAME"
echo "  - ChainID: 9999"
echo "  - P2P Port: $NODE_PORT"
echo "  - RPC Port: $RPC_PORT"
echo "  - Miner Account: $ACCOUNT"
echo "  - Mining Threads: $MINER_THREADS"
echo ""
echo "Start the miner with:"
echo "  $DATADIR/start-miner.sh"
echo ""
echo "Or run in background:"
echo "  nohup $DATADIR/start-miner.sh > /dev/null 2>&1 &"
echo ""
echo "Useful commands:"
echo "  - View logs: tail -f $DATADIR/miner.log"
echo "  - Check mining status:"
echo "    curl -X POST -H 'Content-Type: application/json' \\"
echo "      --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_mining\",\"params\":[],\"id\":1}' \\"
echo "      http://localhost:$RPC_PORT"
echo ""
echo "  - Check hashrate:"
echo "    curl -X POST -H 'Content-Type: application/json' \\"
echo "      --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_hashrate\",\"params\":[],\"id\":1}' \\"
echo "      http://localhost:$RPC_PORT"
echo ""
info "Miner node deployment complete! ✓"
echo ""
question "Start the miner now? (y/n): "
read START_NOW

if [ "$START_NOW" = "y" ]; then
    info "Starting miner node..."
    nohup "$DATADIR/start-miner.sh" > /dev/null 2>&1 &
    MINER_PID=$!
    echo $MINER_PID > "$DATADIR/miner.pid"

    sleep 3

    if ps -p $MINER_PID > /dev/null; then
        info "Miner is running! PID: $MINER_PID"
        echo ""
        info "Checking connection status in 10 seconds..."
        sleep 10

        # Vérifier le nombre de peers
        PEERS=$(curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
            http://localhost:$RPC_PORT 2>/dev/null | grep -o '"result":"[^"]*"' | cut -d'"' -f4)

        if [ ! -z "$PEERS" ]; then
            PEER_COUNT=$((16#${PEERS#0x}))
            info "Connected peers: $PEER_COUNT"

            if [ $PEER_COUNT -gt 0 ]; then
                info "Node is successfully connected to the network! ✓"
            else
                warn "No peers connected yet. This is normal for new nodes."
                warn "If using a bootnode, verify the enode address is correct."
            fi
        fi

        echo ""
        info "Monitor logs with: tail -f $DATADIR/miner.log"
    else
        error "Miner failed to start. Check logs at: $DATADIR/miner.log"
    fi
fi
