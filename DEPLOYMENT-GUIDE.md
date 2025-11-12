# Guide de Déploiement Production - Ducros Network (RandomX)

**Version:** 1.0
**Date:** 2025-11-12
**ChainID:** 9999

---

## 📋 Table des Matières

1. [Vue d'ensemble](#vue-densemble)
2. [Prérequis](#prérequis)
3. [Compilation](#compilation)
4. [Architecture Réseau](#architecture-réseau)
5. [Déploiement Bootnode](#déploiement-bootnode)
6. [Déploiement Nodes Mineurs](#déploiement-nodes-mineurs)
7. [Configuration Firewall](#configuration-firewall)
8. [Vérification](#vérification)
9. [Monitoring](#monitoring)
10. [Troubleshooting](#troubleshooting)

---

## 🎯 Vue d'ensemble

Ce guide explique comment déployer un réseau **Ducros public en production** avec:

- **Consensus:** RandomX (CPU-mining, ASIC-resistant)
- **Difficulté:** LWMA-3 (optimisé pour CPU)
- **Mining API:** Compatible Ethereum (eth_getWork, eth_submitWork)
- **P2P:** Réseau décentralisé avec bootnodes
- **ChainID:** 9999 (unique Ducros Network)

### Architecture Recommandée

Pour un réseau public robuste:

```
┌──────────────────────────────────────────────────────┐
│                  DUCROS NETWORK                       │
│                                                       │
│  ┌─────────────┐      ┌─────────────┐                │
│  │  Bootnode 1 │◄────►│  Bootnode 2 │  (Découverte)  │
│  └──────┬──────┘      └──────┬──────┘                │
│         │                    │                        │
│    ┌────┴────────────────────┴────┐                  │
│    │                               │                  │
│  ┌─▼────────┐  ┌──────────┐  ┌───▼─────┐            │
│  │ Miner 1  │  │ Miner 2  │  │ Miner 3 │  (Consensus)│
│  │ 4 threads│  │ 8 threads│  │ 4 threads│            │
│  └──────────┘  └──────────┘  └──────────┘            │
│                                                       │
│  ┌──────────┐  ┌──────────┐                          │
│  │ Node RPC │  │ Explorer │       (Services)         │
│  └──────────┘  └──────────┘                          │
└──────────────────────────────────────────────────────┘
```

**Minimum recommandé:**
- 1 Bootnode (2 pour redondance)
- 3-5 Nodes mineurs (pour décentralisation)
- 1 Node RPC public (optionnel)

---

## 🔧 Prérequis

### Matériel

**Bootnode:**
- CPU: 2 cores
- RAM: 4 GB
- Stockage: 50 GB SSD
- Bande passante: 100 Mbps

**Node Mineur:**
- CPU: 4-16 cores (plus = meilleur hashrate)
- RAM: 8 GB minimum, 16 GB recommandé
- Stockage: 100 GB SSD
- Bande passante: 100 Mbps

### Logiciel

- Ubuntu 20.04+ ou Debian 11+
- Go 1.21+
- GCC/G++ compiler
- CMake 3.10+
- Git
- curl, jq (utilitaires)

### Réseau

- IP publique statique (recommandé)
- Ports ouverts:
  - **30303/tcp** - P2P connections
  - **30303/udp** - P2P discovery
  - **8545/tcp** - RPC (optionnel)

---

## 🏗️ Compilation

### Étape 1: Installer RandomX Library

```bash
# Sur CHAQUE serveur (bootnode + mineurs)
cd /tmp
git clone https://github.com/tevador/RandomX.git
cd RandomX
mkdir build && cd build
cmake -DARCH=native ..
make -j$(nproc)
sudo make install

# Vérifier l'installation
ls -lh /usr/local/lib/librandomx.a
# Output attendu: -rw-r--r-- 1 root root 1.5M ... /usr/local/lib/librandomx.a
```

### Étape 2: Cloner go-Ducros

```bash
cd ~
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros

# Checkout la branche production
git checkout claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi
```

### Étape 3: Compiler Geth

```bash
export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/usr/local/include"
make geth

# Vérifier la compilation
./build/bin/geth version
# Output attendu: Geth Version: 1.x.x-stable
```

### Étape 4: Vérifier RandomX

```bash
# Vérifier que RandomX est bien lié
ldd ./build/bin/geth | grep randomx
# Output attendu: librandomx.so => /usr/local/lib/librandomx.so
```

---

## 🌐 Architecture Réseau

### Genesis Configuration

Le fichier **genesis-production.json** définit les paramètres du réseau:

```json
{
  "config": {
    "chainId": 9999,           // Unique Ducros Network
    "randomx": {
      "lwmaActivationBlock": 0  // LWMA actif dès le bloc 0
    }
  },
  "difficulty": "1",            // Difficulté initiale basse
  "gasLimit": "8000000",        // 8M gas par bloc
  "alloc": {}                   // Pas de prémine
}
```

**IMPORTANT:** Tous les nodes doivent utiliser le **même fichier genesis**.

### Network ID

- **ChainID:** 9999
- **NetworkID:** 9999

Ces IDs sont uniques à Ducros Network et empêchent les connexions avec d'autres réseaux Ethereum.

---

## 🚀 Déploiement Bootnode

Le bootnode permet aux autres nodes de se découvrir via le protocole DevP2P.

### Étape 1: Préparer le Serveur

```bash
# SSH sur le serveur bootnode
ssh user@bootnode-server

# Aller dans le répertoire go-Ducros
cd ~/go-Ducros

# Rendre le script exécutable
chmod +x deploy-bootnode.sh
```

### Étape 2: Lancer le Déploiement

```bash
./deploy-bootnode.sh
```

Le script va:
1. ✅ Initialiser le bootnode avec genesis
2. ✅ Créer un compte bootnode
3. ✅ Détecter l'IP publique
4. ✅ Configurer le firewall
5. ✅ Lancer le bootnode daemon
6. ✅ Générer l'enode URL

### Étape 3: Récupérer l'Enode

L'enode sera affiché à la fin du script:

```
ENODE (share this with other nodes):

  enode://a1b2c3d4...@123.45.67.89:30303
```

**Sauvegarder cet enode!** Vous en aurez besoin pour connecter les autres nodes.

```bash
# L'enode est aussi sauvegardé dans:
cat bootnode-data/bootnode-enode.txt
```

### Étape 4: Vérifier le Bootnode

```bash
# Vérifier que le processus tourne
ps aux | grep geth

# Vérifier les logs
tail -f bootnode-data/bootnode.log

# Vérifier le RPC
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' \
  http://localhost:8545

# Output attendu: {"jsonrpc":"2.0","id":1,"result":"9999"}
```

### Gestion du Bootnode

```bash
# Voir les logs
tail -f bootnode-data/bootnode.log

# Arrêter le bootnode
kill $(cat bootnode-data/bootnode.pid)

# Redémarrer
./bootnode-data/start-bootnode.sh
```

---

## ⛏️ Déploiement Nodes Mineurs

Les nodes mineurs participent au consensus en calculant des preuves de travail RandomX.

### Étape 1: Préparer le Serveur

```bash
# SSH sur chaque serveur mineur
ssh user@miner-server

# Aller dans le répertoire go-Ducros
cd ~/go-Ducros

# Rendre le script exécutable
chmod +x deploy-miner-node.sh
```

### Étape 2: Lancer le Déploiement

```bash
./deploy-miner-node.sh
```

Le script va demander:

1. **Nom du nœud:** `miner1`, `miner2`, etc.
2. **Port P2P:** 30303 (par défaut)
3. **Port RPC:** 8545 (par défaut)
4. **Threads mining:** 4, 8, 16 (selon votre CPU)
5. **Enode du bootnode:** Collez l'enode récupéré précédemment
6. **Exposer publiquement:** y/n (pour accepter des connexions entrantes)
7. **Password:** Pour sécuriser le compte mineur

### Exemple de Session Interactive

```
Configuration du nœud mineur:

[?] Nom du nœud (ex: miner1, miner2): miner1
[?] Port P2P [30303]: 30303
[?] Port RPC [8545]: 8545
[?] Nombre de threads pour mining [4]: 8
[?] Enode du bootnode: enode://a1b2c3d4...@123.45.67.89:30303
[?] Enter password for miner account: ********

[INFO] Miner account created: 0x1234567890abcdef1234567890abcdef12345678
```

### Étape 3: Démarrer le Mineur

Le script propose de démarrer automatiquement. Sinon:

```bash
# Démarrer manuellement
./data-miner1/start-miner.sh

# Ou en arrière-plan
nohup ./data-miner1/start-miner.sh > /dev/null 2>&1 &
```

### Étape 4: Vérifier le Mining

```bash
# Vérifier que le mining est actif
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545

# Output attendu: {"jsonrpc":"2.0","id":1,"result":true}

# Vérifier le hashrate
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545

# Output attendu: {"jsonrpc":"2.0","id":1,"result":"0x1f40"} (exemple: 8000 H/s)

# Vérifier les peers connectés
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545

# Output attendu: {"jsonrpc":"2.0","id":1,"result":"0x3"} (3 peers)

# Vérifier le dernier bloc
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545

# Output attendu: {"jsonrpc":"2.0","id":1,"result":"0x2a"} (bloc 42 par exemple)
```

### Gestion du Mineur

```bash
# Voir les logs
tail -f data-miner1/miner.log

# Arrêter le mineur
kill $(cat data-miner1/miner.pid)

# Redémarrer
./data-miner1/start-miner.sh
```

---

## 🔒 Configuration Firewall

### Sur le Bootnode

```bash
# Autoriser les connexions P2P
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# Autoriser RPC (si accès externe nécessaire)
sudo ufw allow 8545/tcp

# Activer le firewall
sudo ufw enable

# Vérifier
sudo ufw status
```

### Sur les Mineurs

```bash
# Si le mineur est exposé publiquement
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# RPC uniquement local (plus sécurisé)
# Pas besoin d'ouvrir le port 8545

sudo ufw enable
```

### Vérification des Ports

```bash
# Vérifier les ports en écoute
sudo netstat -tulpn | grep geth

# Output attendu:
# tcp   0.0.0.0:30303   LISTEN   12345/geth
# udp   0.0.0.0:30303            12345/geth
# tcp   127.0.0.1:8545  LISTEN   12345/geth
```

---

## ✅ Vérification

### Checklist Post-Déploiement

#### Bootnode ✓

- [ ] Processus geth en cours d'exécution
- [ ] Port 30303 ouvert et accessible
- [ ] RPC répond sur le port 8545
- [ ] Enode généré et sauvegardé
- [ ] Logs sans erreurs

```bash
# Vérifier tout d'un coup
ps aux | grep geth && \
curl -s http://localhost:8545 && \
cat bootnode-data/bootnode-enode.txt && \
tail -5 bootnode-data/bootnode.log
```

#### Mineurs ✓

- [ ] Processus geth en cours d'exécution
- [ ] Mining actif (eth_mining = true)
- [ ] Hashrate > 0
- [ ] Connecté au bootnode (peers > 0)
- [ ] Blocs synchronisés
- [ ] Logs sans erreurs

```bash
# Vérifier tout d'un coup
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545 | jq && \
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545 | jq && \
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545 | jq
```

### Vérification du Consensus

```bash
# Sur CHAQUE mineur, vérifier que le dernier bloc est le même
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://localhost:8545 | jq '.result.number, .result.hash'

# Output attendu (même sur tous les mineurs):
# "0x2a"
# "0xabcd1234..."
```

Si tous les mineurs affichent le **même bloc number et hash**, le consensus fonctionne! ✓

---

## 📊 Monitoring

### Métriques Clés

#### 1. Block Height

```bash
# Hauteur du bloc actuel
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d\n"
```

#### 2. Hashrate Réseau

```bash
# Hashrate total du réseau (depuis chaque mineur)
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d H/s\n"
```

#### 3. Difficulté

```bash
# Difficulté actuelle
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://localhost:8545 | jq -r '.result.difficulty' | xargs printf "%d\n"
```

#### 4. Peers Connectés

```bash
# Nombre de peers
curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result' | xargs printf "%d\n"
```

#### 5. Block Time (LWMA target: 13s)

```bash
# Temps entre les 2 derniers blocs
CURRENT=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' http://localhost:8545 | jq -r '.result.timestamp' | xargs printf "%d\n")

PREV=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x$((CURRENT_BLOCK-1))",false],"id":1}' http://localhost:8545 | jq -r '.result.timestamp' | xargs printf "%d\n")

echo "Block time: $((CURRENT - PREV)) seconds"
# Target: ~13 seconds
```

### Script de Monitoring Automatique

Créer `monitor.sh`:

```bash
#!/bin/bash
while true; do
    clear
    echo "========================================="
    echo "  DUCROS NETWORK - MONITORING"
    echo "========================================="
    echo ""

    BLOCK=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")
    HASHRATE=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")
    PEERS=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' http://localhost:8545 | jq -r '.result' | xargs printf "%d\n")
    MINING=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' http://localhost:8545 | jq -r '.result')

    echo "Block Height:  $BLOCK"
    echo "Hashrate:      $HASHRATE H/s"
    echo "Peers:         $PEERS"
    echo "Mining:        $MINING"
    echo ""
    echo "Refreshing in 10 seconds..."

    sleep 10
done
```

```bash
chmod +x monitor.sh
./monitor.sh
```

---

## 🔧 Troubleshooting

### Problème: Mineur ne se connecte pas au bootnode

**Symptômes:**
```bash
curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545
# Result: "0x0" (0 peers)
```

**Solutions:**

1. **Vérifier l'enode du bootnode**
   ```bash
   # Sur le bootnode
   ./build/bin/geth --datadir bootnode-data attach --exec 'admin.nodeInfo.enode'
   ```

2. **Vérifier le firewall**
   ```bash
   # Sur le bootnode
   sudo ufw status
   # Port 30303/tcp et 30303/udp doivent être ALLOW
   ```

3. **Vérifier la connectivité réseau**
   ```bash
   # Depuis le mineur
   nc -zv BOOTNODE_IP 30303
   # Devrait afficher: Connection to BOOTNODE_IP 30303 port [tcp/*] succeeded!
   ```

4. **Relancer avec le bon enode**
   ```bash
   # Modifier le script de démarrage du mineur
   nano data-miner1/start-miner.sh
   # Corriger la ligne --bootnodes "enode://..."
   # Relancer
   ```

### Problème: Hashrate à 0

**Symptômes:**
```bash
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545
# Result: "0x0"
```

**Solutions:**

1. **Vérifier que le mining est activé**
   ```bash
   curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
     http://localhost:8545
   # Result devrait être: true
   ```

2. **Vérifier le unlock du compte**
   ```bash
   # Dans les logs
   tail -100 data-miner1/miner.log | grep -i "unlock\|password"
   # Vérifier qu'il n'y a pas d'erreur de déverrouillage
   ```

3. **Vérifier RandomX**
   ```bash
   # Vérifier que la lib est installée
   ldconfig -p | grep randomx
   # Devrait afficher: librandomx.so
   ```

4. **Augmenter la verbosité**
   ```bash
   # Modifier start-miner.sh
   # Changer --verbosity 3 à --verbosity 4
   # Relancer et vérifier les logs
   ```

### Problème: Synchronisation bloquée

**Symptômes:**
```bash
# Le block number ne change pas
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545
# Toujours le même résultat après plusieurs minutes
```

**Solutions:**

1. **Vérifier les peers**
   ```bash
   # Besoin d'au moins 1 peer pour sync
   curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
     http://localhost:8545
   ```

2. **Vérifier le chainID**
   ```bash
   curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
     http://localhost:8545
   # Devrait être: "0x270f" (9999 en décimal)
   ```

3. **Vérifier que tous les nodes ont le même genesis**
   ```bash
   # Sur chaque node
   curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x0",false],"id":1}' \
     http://localhost:8545 | jq '.result.hash'
   # Le hash du bloc genesis doit être identique partout
   ```

4. **Reset et re-init (dernier recours)**
   ```bash
   # ATTENTION: Efface toute la blockchain locale!
   rm -rf data-miner1/geth
   ./build/bin/geth init --datadir data-miner1 genesis-production.json
   ./data-miner1/start-miner.sh
   ```

### Problème: "Too many open files"

**Solutions:**

```bash
# Augmenter les limites
sudo nano /etc/security/limits.conf

# Ajouter:
* soft nofile 65536
* hard nofile 65536

# Redémarrer la session
# Vérifier
ulimit -n
# Devrait afficher: 65536
```

### Problème: Difficulté augmente trop vite

**Solutions:**

LWMA a des limites d'ajustement (max 2× par bloc). Si la difficulté augmente anormalement:

1. **Vérifier le block time moyen**
   ```bash
   # Devrait être ~13 secondes
   # Si <13s → difficulté augmente (normal)
   # Si >13s → difficulté diminue (normal)
   ```

2. **Vérifier les timestamps des blocs**
   ```bash
   # Les timestamps ne doivent pas être manipulés
   curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
     http://localhost:8545 | jq '.result.timestamp'
   ```

---

## 📝 Résumé des Commandes

### Démarrage Rapide

```bash
# 1. Compiler (une seule fois)
make geth

# 2. Déployer bootnode (sur 1 serveur)
./deploy-bootnode.sh

# 3. Récupérer l'enode
cat bootnode-data/bootnode-enode.txt

# 4. Déployer mineurs (sur chaque serveur mineur)
./deploy-miner-node.sh
# Coller l'enode quand demandé

# 5. Vérifier le réseau
curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545 | jq
```

### Gestion Quotidienne

```bash
# Vérifier le statut
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545 | jq

# Voir les logs
tail -f data-miner1/miner.log

# Arrêter
kill $(cat data-miner1/miner.pid)

# Redémarrer
./data-miner1/start-miner.sh
```

---

## 🎉 Conclusion

Votre réseau Ducros est maintenant en **PRODUCTION PUBLIQUE**! 🚀

- ✅ RandomX consensus actif
- ✅ LWMA difficulty ajuste automatiquement
- ✅ Mining décentralisé
- ✅ P2P network établi
- ✅ RPC API disponible

### Prochaines Étapes (Optionnel)

1. **Block Explorer** - Pour visualiser les blocs/transactions
2. **Wallet Interface** - Pour les utilisateurs finaux
3. **Mining Pool** - Pour agréger les petits mineurs
4. **Monitoring Dashboard** - Grafana + Prometheus

### Support

- **Docs:** VERIFYSEAL-LWMA-GUIDE.md, MINING-API.md
- **Build:** BUILD-GUIDE.md
- **Production:** PRODUCTION-READINESS.md

---

**Branche:** `claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi`
**ChainID:** 9999
**Consensus:** RandomX + LWMA-3
**Status:** ✅ PRODUCTION READY
