# Debug : "No mining work available"

## Symptômes

Le stratum-proxy reçoit l'erreur :
```
RPC error -32000: no mining work available yet
```

XMRig affiche :
```
error: "No work available", code: -1
```

## Causes possibles

1. **Geth n'est pas en train de miner** (flag `--mine` manquant)
2. **L'API miner n'est pas exposée** (manquant dans `--http.api`)
3. **Le worker n'a pas encore généré de travail** (geth vient de démarrer)
4. **Geth n'est pas lancé du tout**

## Diagnostic pas à pas

### Étape 1: Vérifier si geth tourne

```bash
# Sur le VPS (92.222.10.107)
ps aux | grep geth

# Devrait afficher quelque chose comme:
# ubuntu   12345  ... ./build/bin/geth --datadir devnet-data ...
```

**Si geth ne tourne pas**, passez à l'étape 5 pour le lancer.

### Étape 2: Vérifier les logs de geth

```bash
# Si geth tourne en background, vérifier les logs
journalctl -u geth -f   # Si c'est un service systemd

# OU si lancé dans un terminal/tmux
# Regarder la sortie du terminal où geth tourne
```

Cherchez ces lignes dans les logs :
```
✅ BON SIGNE:
INFO Mining will start after node initialization
INFO Starting mining operation threads=X
INFO Mining loop started
INFO Mining new block parent=X difficulty=Y

❌ MAUVAIS SIGNE:
- Aucune mention de "mining"
- ERROR ou WARN liés au mining
```

### Étape 3: Tester l'API RPC miner

```bash
# Test 1: Vérifier si l'API répond
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://92.222.10.107:8545

# Réponse attendue:
# {"jsonrpc":"2.0","id":1,"result":true}
#                              ^^^^^ doit être true

# Test 2: Obtenir du travail de mining
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
  http://92.222.10.107:8545

# Si ça marche, vous verrez:
# {"jsonrpc":"2.0","id":1,"result":["0x...","0x...","0x..."]}

# Si ça ne marche pas:
# {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"no mining work available yet"}}
```

### Étape 4: Vérifier la configuration de geth

```bash
# Vérifier les arguments de lancement
ps aux | grep geth | grep -o '\-\-[^ ]*'

# Vérifier que ces flags sont présents:
# --mine                          ✅ CRITIQUE
# --http                          ✅ Nécessaire
# --http.api ... miner ...        ✅ CRITIQUE pour getWork
```

### Étape 5: Lancer geth correctement

Si geth ne tourne pas ou n'est pas configuré correctement :

```bash
# Arrêter geth s'il tourne mal
pkill -9 geth
fuser -k 30303/tcp 30303/udp 8545/tcp

# Lancer avec la bonne configuration
cd /home/ubuntu/go-Ducros

./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx,miner \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.threads 6 \
  --miner.etherbase 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
```

**Flags critiques pour le stratum :**
- `--mine` : Active le mining
- `--http.api eth,net,web3,randomx,miner` : Expose l'API `miner` (contient `eth_getWork`)
- `--miner.etherbase 0x...` : Adresse qui reçoit les récompenses

### Étape 6: Vérifier que le mining démarre

Après avoir lancé geth, attendez ~30 secondes et cherchez ces logs :

```
INFO Allocating RandomX dataset (full mode)
INFO Initializing RandomX dataset in background items=34,078,719
INFO RandomX dataset ready duration=XXs
INFO Mining will start after node initialization
INFO Starting mining operation threads=6
INFO Mining loop started
INFO Mining new block parent=X difficulty=Y
```

**Si vous voyez "Mining loop started"**, le mining est actif.

### Étape 7: Re-tester avec xmrig

Une fois que geth mine correctement, relancez xmrig :

```cmd
xmrig.exe -o 92.222.10.107:3333 -u 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2 -p ducros -a rx/0
```

Vous devriez voir sur le stratum-proxy :
```
✅ Miner logged in: 0x25f...
📤 Sending job to 77.192.84.136
✅ Share accepted!
```

## Cas particuliers

### Geth mine localement mais pas via stratum

Si vous voyez dans les logs de geth :
```
INFO Mining new block parent=X difficulty=Y
✅ Found valid nonce! block=X
```

Mais le stratum dit toujours "No work available", alors :

**Solution :** L'API `miner` n'est probablement pas exposée. Ajoutez `miner` à `--http.api` :

```bash
--http.api eth,net,web3,randomx,personal,miner
#                                        ^^^^^ Ajouter ceci
```

### Geth dit "Mining will start AFTER node initialization"

Si les logs montrent cette ligne mais que le mining ne démarre jamais :

**Cause :** Le nœud n'a pas fini de se synchroniser ou d'initialiser.

**Solution :** Attendez que vous voyiez :
```
INFO Mining operation started
INFO Mining loop started
```

### RandomX dataset prend trop de temps

Si le dataset met >5 minutes à s'initialiser :

**Cause :** Pas de huge pages activées.

**Solution :**
```bash
sudo sysctl -w vm.nr_hugepages=1280
```

Puis relancez geth.

## Checklist de vérification rapide

- [ ] Geth tourne (`ps aux | grep geth`)
- [ ] Flag `--mine` est présent
- [ ] Flag `--http.api` contient `miner`
- [ ] Logs montrent "Mining loop started"
- [ ] `curl ... eth_mining` retourne `true`
- [ ] `curl ... eth_getWork` retourne un tableau de 3 hashes
- [ ] Stratum-proxy ne montre plus "No work available"
- [ ] XMRig reçoit des jobs et mine

## Solution rapide (TL;DR)

```bash
# Sur le VPS
pkill -9 geth
cd /home/ubuntu/go-Ducros

# Avec le nouveau fix compilé
make clean && make geth

# Lancer avec tous les bons flags
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx,miner \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.threads 6 \
  --miner.etherbase 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2

# Attendre ~30 secondes que le dataset s'initialise
# Chercher "Mining loop started" dans les logs
# Puis relancer xmrig
```

## Commandes de diagnostic utiles

```bash
# Est-ce que geth mine ?
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://92.222.10.107:8545 | jq

# Obtenir du travail de mining
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
  http://92.222.10.107:8545 | jq

# Dernier bloc miné
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  http://92.222.10.107:8545 | jq

# Hashrate actuel
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://92.222.10.107:8545 | jq
```
