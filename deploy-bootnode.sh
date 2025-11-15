#!/bin/bash
# deploy-bootnode.sh - Script pour déployer un bootnode Ducros
# Ce bootnode permet aux autres nodes de se découvrir et se connecter

set -e

echo "========================================="
echo "  Ducros Network - Bootnode Deployment"
echo "========================================="
echo ""

# Configuration
DATADIR="./bootnode-data"
BOOTNODE_PORT=30303
RPC_PORT=8545

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Étape 1: Initialiser le bootnode si nécessaire
if [ ! -d "$DATADIR/geth" ]; then
    info "Initializing bootnode with genesis..."
    ./build/bin/geth init --datadir "$DATADIR" genesis-production.json
    info "Bootnode initialized successfully"
else
    warn "Bootnode already initialized, skipping genesis init"
fi

# Étape 2: Créer un compte pour le bootnode (si nécessaire)
if [ ! -f "$DATADIR/bootnode-account.txt" ]; then
    info "Creating bootnode account..."
    ACCOUNT=$(./build/bin/geth --datadir "$DATADIR" account new --password /dev/stdin <<< "bootnode-password" 2>&1 | grep "Public address" | awk '{print $4}')
    echo "$ACCOUNT" > "$DATADIR/bootnode-account.txt"
    info "Bootnode account created: $ACCOUNT"
else
    ACCOUNT=$(cat "$DATADIR/bootnode-account.txt")
    info "Using existing bootnode account: $ACCOUNT"
fi

# Étape 3: Afficher les informations de connexion
info "Bootnode configuration:"
echo "  - ChainID: 9999"
echo "  - P2P Port: $BOOTNODE_PORT"
echo "  - RPC Port: $RPC_PORT"
echo "  - Account: $ACCOUNT"
echo ""

# Étape 4: Obtenir l'IP publique
info "Detecting public IP..."
PUBLIC_IP=$(curl -s https://api.ipify.org 2>/dev/null || echo "YOUR_PUBLIC_IP")
if [ "$PUBLIC_IP" = "YOUR_PUBLIC_IP" ]; then
    warn "Could not detect public IP automatically"
    echo -n "Please enter your public IP address: "
    read PUBLIC_IP
fi
info "Public IP: $PUBLIC_IP"
echo ""

# Étape 5: Instructions firewall
info "Firewall configuration required:"
echo ""
echo "  sudo ufw allow $BOOTNODE_PORT/tcp   # P2P connections"
echo "  sudo ufw allow $BOOTNODE_PORT/udp   # P2P discovery"
echo "  sudo ufw allow $RPC_PORT/tcp        # RPC (optional, for management)"
echo ""
echo -n "Have you configured the firewall? (y/n): "
read FIREWALL_OK
if [ "$FIREWALL_OK" != "y" ]; then
    warn "Please configure firewall and restart this script"
    exit 1
fi

# Étape 6: Lancer le bootnode
info "Starting bootnode..."
echo ""
echo "========================================="
echo "  BOOTNODE IS STARTING"
echo "========================================="
echo ""

# Créer le script de démarrage
cat > "$DATADIR/start-bootnode.sh" <<EOF
#!/bin/bash
# Auto-generated bootnode startup script

./build/bin/geth \\
  --datadir "$DATADIR" \\
  --networkid 9999 \\
  --port $BOOTNODE_PORT \\
  --http \\
  --http.addr "0.0.0.0" \\
  --http.port $RPC_PORT \\
  --http.api "eth,net,web3,admin,personal" \\
  --http.corsdomain "*" \\
  --http.vhosts "*" \\
  --nat "extip:$PUBLIC_IP" \\
  --maxpeers 100 \\
  --netrestrict "" \\
  --nodekey "$DATADIR/bootnode.key" \\
  --verbosity 3 \\
  --syncmode "full" \\
  --gcmode "archive" \\
  2>&1 | tee "$DATADIR/bootnode.log"
EOF

chmod +x "$DATADIR/start-bootnode.sh"

# Générer la nodekey si elle n'existe pas
if [ ! -f "$DATADIR/bootnode.key" ]; then
    info "Generating bootnode key..."
    openssl rand -hex 32 > "$DATADIR/bootnode.key"
fi

# Lancer le bootnode en arrière-plan
info "Launching bootnode daemon..."
nohup "$DATADIR/start-bootnode.sh" > /dev/null 2>&1 &
BOOTNODE_PID=$!

sleep 5

# Vérifier que le bootnode est bien lancé
if ps -p $BOOTNODE_PID > /dev/null; then
    info "Bootnode is running! PID: $BOOTNODE_PID"
    echo $BOOTNODE_PID > "$DATADIR/bootnode.pid"
else
    error "Bootnode failed to start. Check logs at: $DATADIR/bootnode.log"
    exit 1
fi

# Attendre que le RPC soit disponible
info "Waiting for RPC to be ready..."
for i in {1..30}; do
    if curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' \
        http://localhost:$RPC_PORT > /dev/null 2>&1; then
        info "RPC is ready!"
        break
    fi
    sleep 1
done

# Obtenir l'enode
info "Retrieving bootnode enode..."
ENODE=$(./build/bin/geth --datadir "$DATADIR" attach --exec 'admin.nodeInfo.enode' 2>/dev/null | tr -d '"')

if [ -z "$ENODE" ]; then
    warn "Could not retrieve enode automatically. Getting it manually..."
    sleep 3
    ENODE=$(./build/bin/geth --datadir "$DATADIR" attach --exec 'admin.nodeInfo.enode' 2>/dev/null | tr -d '"')
fi

# Remplacer l'IP locale par l'IP publique dans l'enode
ENODE=$(echo $ENODE | sed "s/127.0.0.1/$PUBLIC_IP/g" | sed "s/\[::\]/$PUBLIC_IP/g")

echo ""
echo "========================================="
echo "  BOOTNODE SUCCESSFULLY DEPLOYED!"
echo "========================================="
echo ""
echo "Bootnode Information:"
echo "  - Public IP: $PUBLIC_IP"
echo "  - P2P Port: $BOOTNODE_PORT"
echo "  - RPC Port: $RPC_PORT"
echo "  - ChainID: 9999"
echo ""
echo "ENODE (share this with other nodes):"
echo ""
echo "  $ENODE"
echo ""
echo "Save this enode to a file:"
echo "  echo '$ENODE' > bootnode-enode.txt"
echo ""
echo "Management commands:"
echo "  - View logs: tail -f $DATADIR/bootnode.log"
echo "  - Stop bootnode: kill \$(cat $DATADIR/bootnode.pid)"
echo "  - Restart: $DATADIR/start-bootnode.sh"
echo ""
echo "Other nodes can connect with:"
echo "  --bootnodes '$ENODE'"
echo ""

# Sauvegarder l'enode
echo "$ENODE" > "$DATADIR/bootnode-enode.txt"
info "Enode saved to: $DATADIR/bootnode-enode.txt"
echo ""
info "Bootnode deployment complete! ✓"
