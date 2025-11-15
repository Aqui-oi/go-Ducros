# 🚀 Guide de Lancement Ducros Network

Guide complet pour démarrer un nœud Ducros Network avec mining RandomX.

---

## Prérequis

### 1. Installer RandomX Library

```bash
# Installer les dépendances
sudo apt-get update
sudo apt-get install -y git build-essential cmake

# Cloner et compiler RandomX
git clone https://github.com/tevador/RandomX.git /tmp/RandomX
cd /tmp/RandomX
mkdir build && cd build
cmake -DARCH=native -DBUILD_SHARED_LIBS=ON ..
make -j$(nproc)
sudo make install
sudo ldconfig

# Vérifier l'installation
ldconfig -p | grep randomx
```

### 2. Activer Huge Pages (CRITIQUE pour performance)

```bash
# Activer huge pages (requis pour JIT + performance +1500%)
sudo sysctl -w vm.nr_hugepages=1280

# Rendre permanent
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf

# Vérifier
cat /proc/meminfo | grep HugePages
```

**⚠️ IMPORTANT:** Sans huge pages, le mining sera 15× plus lent !

---

## 🔨 Étape 1: Compiler Geth

```bash
cd /home/user/go-Ducros

# Compiler geth avec RandomX
make geth

# Vérifier la compilation
./build/bin/geth version
```

**Sortie attendue:**
```
Geth
Version: 1.16.7-stable
Architecture: amd64
Go Version: go1.21.x
Operating System: linux
```

---

## 🌱 Étape 2: Initialiser la Blockchain

```bash
# Créer le répertoire de données
mkdir -p ~/.ducros

# Initialiser avec le genesis
./build/bin/geth --datadir ~/.ducros init genesis-production.json
```

**Sortie attendue:**
```
INFO [11-12|12:00:00.000] Successfully wrote genesis state
INFO [11-12|12:00:00.000] Allocated cache and file handles
```

---

## 🚀 Étape 3: Lancer le Nœud

### Option A: Nœud Simple (sans mining)

```bash
./build/bin/geth \
  --datadir ~/.ducros \
  --networkid 9999 \
  --port 30303 \
  --http \
  --http.addr "127.0.0.1" \
  --http.port 8545 \
  --http.api "eth,net,web3,txpool,randomx" \
  --http.corsdomain "*" \
  --ws \
  --ws.addr "127.0.0.1" \
  --ws.port 8546 \
  --ws.api "eth,net,web3,txpool,randomx" \
  --verbosity 3
```

### Option B: Nœud avec Mining CPU Intégré

```bash
./build/bin/geth \
  --datadir ~/.ducros \
  --networkid 9999 \
  --port 30303 \
  --http \
  --http.addr "127.0.0.1" \
  --http.port 8545 \
  --http.api "eth,net,web3,txpool,randomx,miner" \
  --http.corsdomain "*" \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xVOTRE_ADRESSE_ICI \
  --verbosity 3
```

### Option C: Nœud pour Mining Externe (xmrig)

```bash
./build/bin/geth \
  --datadir ~/.ducros \
  --networkid 9999 \
  --port 30303 \
  --http \
  --http.addr "0.0.0.0" \
  --http.port 8545 \
  --http.api "eth,net,web3,txpool,randomx,miner" \
  --http.corsdomain "*" \
  --miner.etherbase 0xVOTRE_ADRESSE_ICI \
  --verbosity 3
```

**Note:** `http.addr "0.0.0.0"` expose le RPC pour le Stratum proxy.

---

## ⛏️ Étape 4: Démarrer le Mining

### Méthode 1: Mining CPU Intégré

Si vous avez lancé avec `--mine`, c'est déjà actif. Sinon:

```bash
# Dans un autre terminal
./start-mining.sh 4  # 4 = nombre de threads
```

Ou via `geth attach`:

```bash
./build/bin/geth attach ~/.ducros/geth.ipc

# Dans la console
> miner.start(4)  // Démarre avec 4 threads
> eth.mining      // Vérifier si mining actif
true
> eth.hashrate    // Voir le hashrate
1234567
```

### Méthode 2: Mining avec xmrig (Recommandé)

**Avantage:** Performance optimale, multi-GPU possible.

#### 4.1: Lancer le Stratum Proxy

```bash
cd /home/user/go-Ducros/stratum-proxy

# Compiler le proxy
go build -o stratum-proxy .

# Lancer le proxy
./stratum-proxy \
  --geth-rpc "http://localhost:8545" \
  --stratum-addr "0.0.0.0:3333" \
  --verbose
```

**Sortie attendue:**
```
🚀 Stratum proxy starting...
✅ RPC connection verified
🌐 Stratum server listening on 0.0.0.0:3333
📊 Difficulty adjustment enabled
```

#### 4.2: Configurer xmrig

Créer `xmrig-config.json`:

```json
{
    "autosave": true,
    "cpu": true,
    "opencl": false,
    "cuda": false,
    "pools": [
        {
            "algo": "rx/0",
            "coin": "monero",
            "url": "localhost:3333",
            "user": "0xVOTRE_ADRESSE_DUCROS",
            "pass": "worker1",
            "keepalive": true,
            "tls": false
        }
    ],
    "randomx": {
        "init": -1,
        "mode": "auto",
        "1gb-pages": true,
        "numa": true
    },
    "cpu": {
        "enabled": true,
        "huge-pages": true,
        "max-threads-hint": 100
    },
    "log-file": "xmrig.log",
    "print-time": 60
}
```

#### 4.3: Lancer xmrig

```bash
# Télécharger xmrig
wget https://github.com/xmrig/xmrig/releases/download/v6.21.0/xmrig-6.21.0-linux-x64.tar.gz
tar -xzf xmrig-6.21.0-linux-x64.tar.gz
cd xmrig-6.21.0

# Lancer
./xmrig -c xmrig-config.json
```

**Sortie attendue:**
```
[2025-11-12 12:00:00.000]  * ABOUT        XMRig/6.21.0 gcc/11.4.0
[2025-11-12 12:00:00.000]  * LIBS         libuv/1.44.2 OpenSSL/3.0.2 hwloc/2.7.1
[2025-11-12 12:00:00.000]  * HUGE PAGES   supported
[2025-11-12 12:00:00.000]  * 1GB PAGES    available
[2025-11-12 12:00:05.000]  net      use pool localhost:3333  rx/0
[2025-11-12 12:00:05.000]  net      new job from localhost:3333 diff 1000
[2025-11-12 12:00:10.000]  cpu      use profile rx
[2025-11-12 12:00:10.000]  cpu      READY threads 8/8 (8) huge pages 100%
[2025-11-12 12:00:30.000]  miner    speed 10s/60s/15m 1234.5 1234.5 n/a H/s
```

---

## 📊 Étape 5: Vérifier le Fonctionnement

### Vérifier la Synchronisation

```bash
./build/bin/geth attach ~/.ducros/geth.ipc

> eth.syncing
false  // false = synchronisé

> eth.blockNumber
123  // Numéro du dernier bloc

> admin.peers.length
5  // Nombre de pairs connectés
```

### Vérifier le Mining

```bash
> eth.mining
true

> eth.hashrate
1234567  // Hashrate en H/s

> miner.getHashrate()
1234567

> eth.getBlock("latest")
{
  difficulty: 2048,
  hash: "0x...",
  miner: "0xVOTRE_ADRESSE",
  number: 123,
  timestamp: 1731412345
}
```

### Vérifier le Solde

```bash
> eth.getBalance("0xVOTRE_ADRESSE")
"5000000000000000000"  // 5 ETH en wei

> web3.fromWei(eth.getBalance("0xVOTRE_ADRESSE"), "ether")
"5"
```

---

## 🐛 Dépannage

### Problème: "randomx: failed to allocate cache"

**Solution:** Activer huge pages
```bash
sudo sysctl -w vm.nr_hugepages=1280
```

### Problème: "RandomX using interpreted mode"

**Cause:** Huge pages non disponibles
**Impact:** Performance -15×
**Solution:** Voir section "Activer Huge Pages"

### Problème: Mining hashrate = 0

```bash
# Vérifier les logs
tail -f ~/.ducros/geth.log | grep -i "randomx\|mining"

# Vérifier la difficulté
> eth.getBlock("latest").difficulty
```

### Problème: Pas de pairs

```bash
# Ajouter des bootnodes manuellement
> admin.addPeer("enode://BOOTNODE_ID@IP:30303")

# Ou redémarrer avec:
--bootnodes "enode://..."
```

### Problème: xmrig "Invalid share"

**Cause:** Bug encodage nonce (corrigé dans dernier commit)
**Solution:** Rebuild stratum-proxy:
```bash
cd stratum-proxy
git pull
go build -o stratum-proxy .
```

---

## 📈 Optimisations Performance

### CPU Mining Optimal

```bash
# Désactiver CPU frequency scaling
sudo cpupower frequency-set -g performance

# Augmenter la priorité du processus
sudo nice -n -20 ./build/bin/geth --mine ...

# Utiliser tous les cœurs sauf 1
--miner.threads $(nproc --ignore=1)
```

### Réseau Optimal

```bash
# Augmenter les limites de connexion
--maxpeers 100 \
--maxpendpeers 50
```

### Disque Optimal

```bash
# Utiliser SSD si possible
# Augmenter le cache
--cache 2048  # 2GB cache
```

---

## 🔒 Sécurité Production

### Ne PAS Exposer le RPC Publiquement

```bash
# ❌ DANGEREUX
--http.addr "0.0.0.0" --http.api "eth,net,web3,miner,admin"

# ✅ SÉCURISÉ
--http.addr "127.0.0.1" --http.api "eth,net,web3"
```

### Utiliser Firewall

```bash
# Autoriser seulement P2P
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# Bloquer RPC par défaut
sudo ufw deny 8545/tcp
```

### Backup Clés Privées

```bash
# Backup du keystore
cp -r ~/.ducros/keystore ~/backup/keystore-$(date +%Y%m%d)
```

---

## 📋 Résumé des Commandes Rapides

### Démarrage Rapide (tout-en-un)

```bash
# 1. Compiler
cd /home/user/go-Ducros && make geth

# 2. Init genesis
./build/bin/geth --datadir ~/.ducros init genesis-production.json

# 3. Lancer nœud + mining
./build/bin/geth \
  --datadir ~/.ducros \
  --networkid 9999 \
  --http --http.port 8545 \
  --http.api "eth,net,web3,randomx,miner" \
  --mine --miner.threads 4 \
  --miner.etherbase 0xVOTRE_ADRESSE \
  --verbosity 3
```

### Monitoring

```bash
# Voir les logs
tail -f ~/.ducros/geth.log

# Attach console
./build/bin/geth attach ~/.ducros/geth.ipc

# Vérifier mining
> eth.mining && eth.hashrate

# Vérifier blocks
> eth.blockNumber && eth.getBlock("latest").miner
```

---

## 🎯 Checklist Pré-Production

- [ ] RandomX library installée (`ldconfig -p | grep randomx`)
- [ ] Huge pages activées (`cat /proc/meminfo | grep HugePages`)
- [ ] Geth compilé (`./build/bin/geth version`)
- [ ] Genesis initialisé (`ls ~/.ducros/geth/chaindata/`)
- [ ] Adresse mining configurée (`--miner.etherbase`)
- [ ] Firewall configuré (port 30303 ouvert)
- [ ] Bootnodes configurés (`--bootnodes`)
- [ ] Backup keystore fait
- [ ] Monitoring activé (Prometheus/Grafana optionnel)

---

## 🆘 Support

**Logs:**
- Geth: `~/.ducros/geth.log`
- xmrig: `./xmrig.log`
- Stratum: `./stratum-proxy.log`

**Documentation:**
- [EVM-COMPATIBILITY.md](./EVM-COMPATIBILITY.md) - Compatibilité EVM
- [GETH-UPSTREAM-STRATEGY.md](./GETH-UPSTREAM-STRATEGY.md) - Stratégie upstream
- [POOL-OPERATOR-GUIDE.md](./POOL-OPERATOR-GUIDE.md) - Guide pool operators

**Performance Attendue:**
- Ryzen 9 5950X: ~15,000 H/s
- Intel i9-12900K: ~18,000 H/s
- Ryzen 7 5800X: ~10,000 H/s

---

**Bonne chance avec ton lancement Ducros Network! 🚀**
