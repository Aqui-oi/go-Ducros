#!/bin/bash
# Script de démarrage rapide du Stratum-Proxy Ducros

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "========================================="
echo "  Ducros Stratum-Proxy Quick Start"
echo "========================================="
echo ""

# Vérifier que geth tourne
echo -e "${YELLOW}Vérification de geth...${NC}"
if ! pgrep -x geth > /dev/null; then
    echo -e "${RED}❌ Geth ne tourne pas !${NC}"
    echo ""
    echo "Lancez geth d'abord avec :"
    echo "  ./build/bin/geth --datadir devnet-data --networkid 33669 \\"
    echo "    --http --http.api eth,net,web3,randomx,miner \\"
    echo "    --http.addr 0.0.0.0 --http.port 8545 \\"
    echo "    --mine --miner.threads 6"
    echo ""
    exit 1
fi
echo -e "${GREEN}✅ Geth est en cours d'exécution${NC}"

# Vérifier l'API RPC
echo -e "${YELLOW}Vérification de l'API RPC...${NC}"
if curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
    http://localhost:8545 | grep -q '"result":true'; then
    echo -e "${GREEN}✅ Geth est en train de miner${NC}"
else
    echo -e "${RED}❌ Geth ne mine pas !${NC}"
    echo "Vérifiez que geth a le flag --mine"
    exit 1
fi

# Vérifier eth_getWork
echo -e "${YELLOW}Vérification du travail disponible...${NC}"
if curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
    http://localhost:8545 | grep -q '"result":\["0x'; then
    echo -e "${GREEN}✅ Travail de mining disponible${NC}"
else
    echo -e "${YELLOW}⚠️  Pas encore de travail disponible (attendez 30s que le dataset s'initialise)${NC}"
fi

echo ""

# Aller dans le dossier stratum-proxy
cd "$(dirname "$0")/stratum-proxy"

# Compiler si nécessaire
if [ ! -f stratum-proxy ]; then
    echo -e "${YELLOW}Compilation du stratum-proxy...${NC}"
    go build -o stratum-proxy .
    echo -e "${GREEN}✅ Compilation réussie${NC}"
fi

# Configuration par défaut
STRATUM_ADDR="${STRATUM_ADDR:-0.0.0.0:3333}"
GETH_RPC="${GETH_RPC:-http://localhost:8545}"
INITIAL_DIFF="${INITIAL_DIFF:-100000}"
ALGO="${ALGO:-rx/0}"

echo ""
echo "Configuration :"
echo "  - Écoute : $STRATUM_ADDR"
echo "  - Geth : $GETH_RPC"
echo "  - Difficulté : $INITIAL_DIFF"
echo "  - Algorithme : $ALGO"
echo ""

# Vérifier le firewall
if command -v ufw &> /dev/null; then
    if ! sudo ufw status | grep -q "3333.*ALLOW"; then
        echo -e "${YELLOW}⚠️  Le port 3333 n'est peut-être pas ouvert${NC}"
        echo "Ouvrez-le avec : sudo ufw allow 3333/tcp"
        echo ""
    fi
fi

echo -e "${GREEN}🚀 Démarrage du Stratum-Proxy...${NC}"
echo ""

# Lancer le proxy
./stratum-proxy \
  --stratum "$STRATUM_ADDR" \
  --geth "$GETH_RPC" \
  --diff "$INITIAL_DIFF" \
  --algo "$ALGO" \
  --vardiff-target 30.0 \
  --vardiff-window 10 \
  --max-invalid-streak 10 \
  -v
