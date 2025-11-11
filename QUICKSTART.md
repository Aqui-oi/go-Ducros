# 🚀 Guide de Démarrage Rapide - RandomX Mining

## ✅ Étape 1 : Pull les derniers changements sur ton VPS

```bash
cd ~/go-Ducros
git pull origin claude/geth-randomx-pow-fork-011CV1zCZx1k45jWEf7eXxMT
```

## ✅ Étape 2 : Vérifier que RandomX est bien installé

```bash
# Vérifier que la bibliothèque est installée
ls /usr/local/lib/librandomx.a
ls /usr/local/include/randomx.h

# Si non trouvé, réinstaller :
# cd ~/RandomX/build && sudo make install
```

## ✅ Étape 3 : Build Geth avec RandomX

```bash
cd ~/go-Ducros

# Activer CGO et compiler
export CGO_ENABLED=1
make geth
```

**Note** : La compilation peut prendre 2-5 minutes.

## ✅ Étape 4 : Initialiser la blockchain

```bash
# Créer le répertoire de données
mkdir -p ./data-randomx

# Initialiser avec le genesis RandomX
./build/bin/geth init genesis-randomx.json --datadir ./data-randomx
```

**Output attendu** :
```
INFO [XX-XX|XX:XX:XX.XXX] Maximum peer count                       ETH=50 total=50
INFO [XX-XX|XX:XX:XX.XXX] Smartcard socket not found, disabling    err="stat /run/pcscd/pcscd.comm: no such file or directory"
INFO [XX-XX|XX:XX:XX.XXX] Set global gas cap                       cap=50,000,000
INFO [XX-XX|XX:XX:XX.XXX] Initializing the KZG library             backend=gokzg
INFO [XX-XX|XX:XX:XX.XXX] Allocated cache and file handles         database=/home/ubuntu/go-Ducros/data-randomx/geth/chaindata cache=16.00MiB handles=16
INFO [XX-XX|XX:XX:XX.XXX] Opened ancient database                  database=/home/ubuntu/go-Ducros/data-randomx/geth/chaindata/ancient/chain readonly=false
INFO [XX-XX|XX:XX:XX.XXX] Writing custom genesis block
INFO [XX-XX|XX:XX:XX.XXX] Persisted trie from memory database      nodes=8 size=1.18KiB time="XXXµs" gcnodes=0 gcsize=0.00B gctime=0s livenodes=0 livesize=0.00B
INFO [XX-XX|XX:XX:XX.XXX] Successfully wrote genesis state         database=chaindata hash=XXXXXX
INFO [XX-XX|XX:XX:XX.XXX] Allocated cache and file handles         database=/home/ubuntu/go-Ducros/data-randomx/geth/lightchaindata cache=16.00MiB handles=16
INFO [XX-XX|XX:XX:XX.XXX] Opened ancient database                  database=/home/ubuntu/go-Ducros/data-randomx/geth/lightchaindata/ancient/chain readonly=false
INFO [XX-XX|XX:XX:XX.XXX] Writing custom genesis block
INFO [XX-XX|XX:XX:XX.XXX] Persisted trie from memory database      nodes=8 size=1.18KiB time="XXXµs" gcnodes=0 gcsize=0.00B gctime=0s livenodes=0 livesize=0.00B
INFO [XX-XX|XX:XX:XX.XXX] Successfully wrote genesis state         database=lightchaindata hash=XXXXXX
```

## ✅ Étape 5 : Créer un compte de mining

```bash
# Créer un nouveau compte (miner address)
./build/bin/geth account new --datadir ./data-randomx

# Entrer un mot de passe (le retenir !)
# Note l'adresse créée : 0x...
```

## ✅ Étape 6 : Lancer le nœud

### Option A : Nœud de développement (solo mining)

```bash
./build/bin/geth \
  --datadir ./data-randomx \
  --networkid 33669 \
  --http \
  --http.addr "0.0.0.0" \
  --http.port 8545 \
  --http.api "eth,net,web3,personal,miner,admin,txpool,debug" \
  --http.corsdomain "*" \
  --allow-insecure-unlock \
  --nodiscover \
  --maxpeers 0 \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xTON_ADRESSE_ICI \
  --verbosity 4 \
  console
```

**Remplacer** `0xTON_ADRESSE_ICI` par l'adresse créée à l'étape 5.

### Option B : Nœud en arrière-plan

```bash
# Lancer en background avec nohup
nohup ./build/bin/geth \
  --datadir ./data-randomx \
  --networkid 33669 \
  --http \
  --http.addr "0.0.0.0" \
  --http.port 8545 \
  --http.api "eth,net,web3,personal,miner,admin,txpool" \
  --http.corsdomain "*" \
  --mine \
  --miner.threads 4 \
  --miner.etherbase 0xTON_ADRESSE_ICI \
  --verbosity 3 \
  > geth.log 2>&1 &

# Suivre les logs
tail -f geth.log
```

## ✅ Étape 7 : Vérifier que le mining fonctionne

### Dans la console Geth (si Option A)

```javascript
// Vérifier le mining
eth.mining
// true

// Voir le hashrate
eth.hashrate
// Exemple: 1234 H/s

// Voir le block actuel
eth.blockNumber
// Devrait augmenter

// Voir ton balance
eth.getBalance(eth.coinbase)
// Augmente à chaque bloc miné

// Arrêter/redémarrer le mining
miner.stop()
miner.start(4)  // 4 threads
```

### Via RPC (si Option B)

```bash
# Depuis un autre terminal
./build/bin/geth attach http://localhost:8545

# Ou via curl
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}'

curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

## 🎯 Logs à Surveiller

### ✅ Logs Normaux (Tout va bien)

```
INFO [XX-XX|XX:XX:XX.XXX] Using RandomX PoW consensus engine
INFO [XX-XX|XX:XX:XX.XXX] Starting mining operation
INFO [XX-XX|XX:XX:XX.XXX] Commit new mining work                   number=1 txs=0 uncles=0 elapsed=XXXµs
INFO [XX-XX|XX:XX:XX.XXX] Successfully sealed new block            number=1 hash=0x... elapsed=XXs
INFO [XX-XX|XX:XX:XX.XXX] 🔨 mined potential block                  number=1 hash=0x...
```

### ⚠️ Logs Problèmes

```
ERROR: Failed to load RandomX library
→ Solution: Vérifier que librandomx.a est installée

ERROR: undefined reference to randomx_*
→ Solution: Recompiler avec CGO_ENABLED=1

WARN: Mining too far in the future
→ Normal au démarrage, s'arrête après quelques blocs
```

## 📊 Performance Attendue

| CPU | Threads | Hashrate Estimé |
|-----|---------|-----------------|
| 2 cores | 2 | ~200-500 H/s |
| 4 cores | 4 | ~500-1000 H/s |
| 8 cores | 8 | ~1000-2000 H/s |

**Note** : RandomX est CPU-intensive et utilise ~2GB RAM par thread.

## 🔧 Paramètres Importants

### Ajuster le nombre de threads

```javascript
// Dans console Geth
miner.stop()
miner.start(8)  // Utiliser 8 threads

// Via ligne de commande
--miner.threads 8
```

### Ajuster la difficulté initiale

Si les blocs sont trop lents/rapides, modifier `genesis-randomx.json` :

```json
{
  "difficulty": "0x20000"  // Plus haut = plus difficile
}
```

Puis réinitialiser :
```bash
rm -rf ./data-randomx/geth
./build/bin/geth init genesis-randomx.json --datadir ./data-randomx
```

## 🧪 Tests Rapides

### Test 1 : Vérifier que RandomX est actif

```bash
./build/bin/geth --datadir ./data-randomx --exec "admin.nodeInfo.protocols.eth.consensus" attach
# Devrait retourner "randomx"
```

### Test 2 : Miner 1 bloc

```bash
./build/bin/geth --datadir ./data-randomx \
  --networkid 33669 \
  --mine \
  --miner.threads 1 \
  --nodiscover \
  --maxpeers 0 \
  console

# Dans console
miner.start(1)
# Attendre quelques secondes
eth.blockNumber
# Devrait être > 0
```

### Test 3 : Vérifier les rewards

```javascript
// Balance du coinbase
web3.fromWei(eth.getBalance(eth.coinbase), "ether")
// Exemple: "10" (si 2 blocs minés à 5 ETH chacun)

// Voir les détails d'un bloc
eth.getBlock(1)
// "miner": "0xTON_ADRESSE"
// "reward": devrait être 5 ETH (5000000000000000000 wei)
```

## 🛑 Arrêter proprement

```bash
# Dans console
exit

# Ou tuer le processus
pkill geth

# Ou via signal
kill -SIGTERM $(pgrep geth)
```

## 📝 Fichiers Importants

| Fichier | Description |
|---------|-------------|
| `./data-randomx/geth/chaindata/` | Base de données blockchain |
| `./data-randomx/keystore/` | Clés privées des comptes |
| `./data-randomx/geth.ipc` | Socket IPC pour geth attach |
| `geth.log` | Logs (si lancé en background) |

## 🔐 Sécurité

⚠️ **IMPORTANT** :
- Ne jamais exposer le RPC HTTP avec `--allow-insecure-unlock` en production
- Sauvegarder le `keystore/` régulièrement
- Utiliser un mot de passe fort
- Ne pas partager ton adresse de mining publiquement avant le mainnet

## ❓ Dépannage

### Problème : `librandomx.a: cannot find -lrandomx`

**Solution** :
```bash
# Vérifier que la lib est installée
sudo ldconfig -p | grep randomx

# Si vide, réinstaller
cd ~/RandomX/build
sudo make install
sudo ldconfig
```

### Problème : `consensus/randomx: build constraints exclude all Go files`

**Solution** :
```bash
# Activer CGO
export CGO_ENABLED=1
make geth
```

### Problème : Blocs ne sont pas minés

**Vérifications** :
1. `eth.mining` retourne `true` ?
2. `eth.hashrate` > 0 ?
3. Difficulté trop haute ? Voir section "Ajuster la difficulté"
4. Logs d'erreur ? Augmenter verbosity : `--verbosity 5`

### Problème : Out of memory

**Solution** :
```bash
# Réduire le nombre de threads
miner.stop()
miner.start(2)  # Au lieu de 8

# Ou augmenter la RAM swap
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

## 🎉 Prochaines Étapes

1. ✅ Faire tourner un nœud solo
2. ⬜ Connecter plusieurs nœuds (P2P)
3. ⬜ Déployer un explorer (blockscout)
4. ⬜ Créer un faucet
5. ⬜ Tester les smart contracts
6. ⬜ Déployer un testnet public

## 📚 Ressources

- **Documentation RandomX** : https://github.com/tevador/RandomX
- **Geth Docs** : https://geth.ethereum.org/docs
- **Votre README** : `RANDOMX-IMPLEMENTATION.md`

---

**Bonne chance avec ton mining RandomX ! 🚀💎**
