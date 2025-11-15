#!/bin/bash
# Script de diagnostic pour "No mining work available"

set -e

GETH_RPC="http://92.222.10.107:8545"
MINER_ADDRESS="0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2"

echo "================================================"
echo "🔍 Diagnostic du mining Geth + Stratum"
echo "================================================"
echo ""

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Vérifier si geth tourne
echo "1️⃣  Vérification du processus geth..."
if pgrep -x geth > /dev/null; then
    echo -e "${GREEN}✅ Geth est en cours d'exécution${NC}"
    GETH_PID=$(pgrep -x geth)
    echo "   PID: $GETH_PID"

    # Afficher les arguments
    echo "   Arguments:"
    ps -p $GETH_PID -o args= | tr ' ' '\n' | grep -- '--' | sed 's/^/      /'

    # Vérifier --mine
    if ps -p $GETH_PID -o args= | grep -q -- '--mine'; then
        echo -e "   ${GREEN}✅ Flag --mine présent${NC}"
    else
        echo -e "   ${RED}❌ Flag --mine MANQUANT${NC}"
    fi

    # Vérifier --http.api avec miner
    if ps -p $GETH_PID -o args= | grep -q -- '--http.api.*miner'; then
        echo -e "   ${GREEN}✅ API 'miner' exposée${NC}"
    else
        echo -e "   ${RED}❌ API 'miner' MANQUANTE dans --http.api${NC}"
    fi
else
    echo -e "${RED}❌ Geth ne tourne pas${NC}"
    echo "   Lancez geth avec: ./build/bin/geth --mine --http.api eth,net,web3,randomx,miner ..."
    exit 1
fi

echo ""

# 2. Tester l'API RPC
echo "2️⃣  Test de l'API RPC..."
if ! command -v curl &> /dev/null; then
    echo -e "${YELLOW}⚠️  curl n'est pas installé, test RPC ignoré${NC}"
else
    # Test eth_mining
    MINING_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
      $GETH_RPC 2>/dev/null || echo '{"error":"connection failed"}')

    if echo "$MINING_RESPONSE" | grep -q '"result":true'; then
        echo -e "${GREEN}✅ eth_mining retourne true (mining actif)${NC}"
    elif echo "$MINING_RESPONSE" | grep -q '"result":false'; then
        echo -e "${RED}❌ eth_mining retourne false (mining inactif)${NC}"
    else
        echo -e "${RED}❌ Impossible de se connecter à l'API RPC${NC}"
        echo "   Réponse: $MINING_RESPONSE"
    fi

    # Test eth_getWork
    WORK_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
      $GETH_RPC 2>/dev/null || echo '{"error":"connection failed"}')

    if echo "$WORK_RESPONSE" | grep -q '"result":\["0x'; then
        echo -e "${GREEN}✅ eth_getWork retourne du travail de mining${NC}"
        echo "   Work disponible pour les mineurs"
    elif echo "$WORK_RESPONSE" | grep -q 'no mining work available'; then
        echo -e "${RED}❌ eth_getWork: no mining work available yet${NC}"
        echo -e "   ${YELLOW}Attendez ~30 secondes que le dataset RandomX s'initialise${NC}"
    else
        echo -e "${RED}❌ eth_getWork échoue${NC}"
        echo "   Réponse: $WORK_RESPONSE"
    fi

    # Dernier bloc
    BLOCK_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
      $GETH_RPC 2>/dev/null || echo '{"error":"connection failed"}')

    if echo "$BLOCK_RESPONSE" | grep -q '"number"'; then
        BLOCK_NUM=$(echo "$BLOCK_RESPONSE" | grep -o '"number":"0x[0-9a-f]*"' | head -1 | cut -d'"' -f4)
        BLOCK_DECIMAL=$((16#${BLOCK_NUM#0x}))
        echo -e "${GREEN}✅ Dernier bloc: #$BLOCK_DECIMAL${NC}"
    fi
fi

echo ""

# 3. Vérifier les logs geth
echo "3️⃣  Vérification des logs récents..."
if [ -f "/var/log/geth.log" ]; then
    echo "   Dernières lignes importantes:"
    tail -100 /var/log/geth.log | grep -E "(Mining|RandomX|dataset|ERROR|WARN)" | tail -5
elif command -v journalctl &> /dev/null; then
    echo "   Dernières lignes importantes (journalctl):"
    journalctl -u geth -n 100 --no-pager 2>/dev/null | grep -E "(Mining|RandomX|dataset|ERROR|WARN)" | tail -5 || echo "   Pas de logs systemd"
else
    echo -e "   ${YELLOW}⚠️  Logs non trouvés (vérifiez la sortie du terminal où geth tourne)${NC}"
fi

echo ""

# 4. Résumé et recommandations
echo "================================================"
echo "📊 RÉSUMÉ"
echo "================================================"

ISSUES=0

if ! pgrep -x geth > /dev/null; then
    echo -e "${RED}❌ Geth ne tourne pas${NC}"
    ISSUES=$((ISSUES + 1))
fi

if pgrep -x geth > /dev/null && ! ps -p $(pgrep -x geth) -o args= | grep -q -- '--mine'; then
    echo -e "${RED}❌ Flag --mine manquant${NC}"
    ISSUES=$((ISSUES + 1))
fi

if pgrep -x geth > /dev/null && ! ps -p $(pgrep -x geth) -o args= | grep -q -- '--http.api.*miner'; then
    echo -e "${RED}❌ API 'miner' non exposée${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ "$ISSUES" -eq 0 ]; then
    echo -e "${GREEN}✅ Configuration semble correcte${NC}"
    echo ""
    echo "Si le stratum affiche toujours 'No work available':"
    echo "  1. Attendez 30-60 secondes que le dataset RandomX s'initialise"
    echo "  2. Vérifiez les logs geth pour 'Mining loop started'"
    echo "  3. Relancez le stratum-proxy"
else
    echo -e "${RED}❌ $ISSUES problème(s) détecté(s)${NC}"
    echo ""
    echo "SOLUTION RECOMMANDÉE:"
    echo ""
    echo "pkill -9 geth"
    echo "cd /home/ubuntu/go-Ducros"
    echo "./build/bin/geth \\"
    echo "  --datadir devnet-data \\"
    echo "  --networkid 33669 \\"
    echo "  --http \\"
    echo "  --http.api eth,net,web3,randomx,miner \\"
    echo "  --http.addr 0.0.0.0 \\"
    echo "  --http.port 8545 \\"
    echo "  --http.corsdomain \"*\" \\"
    echo "  --mine \\"
    echo "  --miner.threads 6 \\"
    echo "  --miner.etherbase $MINER_ADDRESS"
fi

echo ""
echo "================================================"
