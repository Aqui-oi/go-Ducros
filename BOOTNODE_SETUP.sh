#!/bin/bash
# Script pour générer et afficher l'enode URL d'un bootnode
# Usage: ./BOOTNODE_SETUP.sh

set -e

echo "========================================="
echo "  DUCROS BOOTNODE GENERATOR"
echo "========================================="
echo ""

# Vérifier que bootnode existe
if [ ! -f "./build/bin/bootnode" ]; then
    echo "❌ Bootnode binary not found!"
    echo "Run 'make all' first to compile it."
    exit 1
fi

# Générer la clé si elle n'existe pas
if [ ! -f "boot.key" ]; then
    echo "🔑 Generating new bootnode key..."
    ./build/bin/bootnode -genkey boot.key
    echo "✅ Key generated: boot.key"
    echo ""
else
    echo "⚠️  boot.key already exists, using existing key"
    echo ""
fi

# Obtenir l'adresse publique
echo "📋 Extracting bootnode public address..."
ENODE_ADDR=$(./build/bin/bootnode -nodekey boot.key -writeaddress)
echo "✅ Bootnode address: $ENODE_ADDR"
echo ""

# Obtenir l'IP publique
echo "🌐 Detecting public IP address..."
IP_PUBLIC=$(curl -s ifconfig.me)
if [ -z "$IP_PUBLIC" ]; then
    IP_PUBLIC=$(curl -s icanhazip.com)
fi
echo "✅ Public IP: $IP_PUBLIC"
echo ""

# Construire l'enode URL
ENODE_URL="enode://${ENODE_ADDR}@${IP_PUBLIC}:30310"

echo "========================================="
echo "  YOUR BOOTNODE ENODE URL"
echo "========================================="
echo ""
echo "$ENODE_URL"
echo ""
echo "========================================="
echo ""
echo "📝 Next steps:"
echo "1. Copy the enode URL above"
echo "2. Add it to params/bootnodes_ducros.go"
echo "3. Recompile: make geth"
echo "4. Start bootnode: ./build/bin/bootnode -nodekey boot.key -addr :30310"
echo ""
echo "⚠️  IMPORTANT: Keep boot.key file secure!"
echo "    It should NOT be committed to git."
echo ""
