# Ducros Network - Quick Start Production

**ChainID:** 9999 | **Consensus:** RandomX + LWMA-3 | **Status:** ✅ Production Ready

---

## 🚀 Déploiement en 3 Étapes

### Prérequis

- Ubuntu 20.04+ avec Go 1.21+
- Minimum 3 serveurs (1 bootnode + 2 mineurs)
- IP publique statique recommandée

---

### Étape 1: Compilation (sur TOUS les serveurs)

```bash
# Installer RandomX
cd /tmp
git clone https://github.com/tevador/RandomX.git
cd RandomX && mkdir build && cd build
cmake -DARCH=native ..
make -j$(nproc)
sudo make install

# Cloner et compiler go-Ducros
cd ~
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros
git checkout claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi

export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/usr/local/include"
make geth
```

---

### Étape 2: Déployer le Bootnode (1 serveur)

```bash
cd ~/go-Ducros

# Configurer le firewall
sudo ./configure-firewall.sh
# Choisir: 1) Bootnode

# Déployer le bootnode
./deploy-bootnode.sh
```

Le script affiche l'**enode** à la fin:
```
enode://a1b2c3d4...@123.45.67.89:30303
```

**Copier cet enode!** Vous en aurez besoin pour les mineurs.

---

### Étape 3: Déployer les Mineurs (2+ serveurs)

Sur **CHAQUE** serveur mineur:

```bash
cd ~/go-Ducros

# Configurer le firewall
sudo ./configure-firewall.sh
# Choisir: 2) Public Miner  OU  3) Private Miner

# Déployer le mineur
./deploy-miner-node.sh
```

Le script demande:
- **Nom:** `miner1`, `miner2`, etc.
- **Threads:** Nombre de CPU cores (4, 8, 16...)
- **Enode du bootnode:** Coller l'enode copié à l'étape 2

---

## ✅ Vérification

### Sur le Bootnode

```bash
# Vérifier les peers connectés
curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d\n"

# Devrait afficher: 2 ou plus (nombre de mineurs connectés)
```

### Sur les Mineurs

```bash
# Vérifier le mining
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545 | jq

# Result: true ✓

# Vérifier le hashrate
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d H/s\n"

# Devrait afficher: >0 H/s ✓

# Vérifier la synchronisation
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d\n"

# Le block number devrait augmenter toutes les ~13 secondes ✓
```

### Consensus Check

Sur **TOUS** les mineurs, exécuter:

```bash
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://localhost:8545 | jq -r '.result | "\(.number) - \(.hash)"'
```

**Tous les mineurs doivent afficher le même bloc!** ✓

---

## 🎉 Réseau en Production!

Votre réseau Ducros est maintenant **LIVE** en production publique:

- ✅ RandomX consensus actif
- ✅ LWMA difficulty s'ajuste automatiquement
- ✅ Mining décentralisé fonctionnel
- ✅ P2P network établi
- ✅ Prêt pour les utilisateurs

---

## 📊 Monitoring

### Dashboard Simple

```bash
# Créer monitor.sh
cat > monitor.sh <<'EOF'
#!/bin/bash
while true; do
    clear
    echo "========== DUCROS NETWORK =========="
    BLOCK=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")
    HASH=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")
    PEERS=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")

    echo "Block:    $BLOCK"
    echo "Hashrate: $HASH H/s"
    echo "Peers:    $PEERS"
    echo ""
    echo "Refreshing in 10s..."
    sleep 10
done
EOF

chmod +x monitor.sh
./monitor.sh
```

---

## 📚 Documentation Complète

- **DEPLOYMENT-GUIDE.md** - Guide détaillé avec troubleshooting
- **PRODUCTION-READINESS.md** - État de production et architecture
- **BUILD-GUIDE.md** - Compilation et dépendances
- **MINING-API.md** - API RPC mining
- **VERIFYSEAL-LWMA-GUIDE.md** - Détails techniques consensus

---

## 🔧 Gestion

### Arrêter un Node

```bash
# Bootnode
kill $(cat bootnode-data/bootnode.pid)

# Mineur
kill $(cat data-miner1/miner.pid)
```

### Redémarrer

```bash
# Bootnode
./bootnode-data/start-bootnode.sh

# Mineur
./data-miner1/start-miner.sh
```

### Voir les Logs

```bash
# Bootnode
tail -f bootnode-data/bootnode.log

# Mineur
tail -f data-miner1/miner.log
```

---

## 🆘 Support

En cas de problème, consulter **DEPLOYMENT-GUIDE.md** section Troubleshooting.

**Problèmes fréquents:**
- Aucun peer connecté → Vérifier firewall et enode
- Hashrate = 0 → Vérifier unlock du compte
- Blocs ne sync pas → Vérifier genesis identique partout

---

**ChainID:** 9999
**Consensus:** RandomX (CPU-mining)
**Difficulty:** LWMA-3 (target 13s)
**Branch:** `claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi`

✅ **PRODUCTION READY** - Déployez maintenant! 🚀
