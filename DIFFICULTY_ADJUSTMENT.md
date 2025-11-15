# Ajustement de la difficulté RandomX

## Problème observé

Avec `LWMAMinDifficulty = 1`, les blocs sont minés trop rapidement (plusieurs par seconde), causant :
- Erreurs `invalid timestamp` (timestamps identiques pour blocs consécutifs)
- Difficulté à maintenir l'ordre chronologique des blocs
- Réorganisations de chaîne fréquentes

## Solution appliquée

### Modification de la difficulté minimale

**Fichier :** `consensus/randomx/lwma.go`
**Ligne 18 :** `LWMAMinDifficulty = 100000` (augmenté de 1 → 100000)

### Valeurs de difficulté recommandées

| Contexte | Difficulté minimale | Temps de bloc estimé |
|----------|---------------------|----------------------|
| **Dev local** (1 CPU) | 10,000 - 50,000 | 1-5 secondes |
| **Testnet** (multi-node) | 100,000 - 500,000 | 5-30 secondes |
| **Production** | 1,000,000+ | Selon hashrate réseau |

### Comment ajuster la difficulté

#### Option 1: Modifier le code source (appliqué)

```go
// consensus/randomx/lwma.go ligne 18
LWMAMinDifficulty = 100000  // Ajustez cette valeur
```

#### Option 2: Modifier le genesis.json

```json
{
  "config": {
    "chainId": 33669,
    "randomx": {}
  },
  "difficulty": "0x186A0",  // 100000 en hexadécimal
  ...
}
```

**Conversions hexadécimales utiles :**
- 1,000 = `0x3E8`
- 10,000 = `0x2710`
- 100,000 = `0x186A0`
- 1,000,000 = `0xF4240`
- 10,000,000 = `0x989680`

#### Option 3: Ajuster les paramètres LWMA

Dans `consensus/randomx/lwma.go`, vous pouvez aussi modifier :

```go
LWMATargetBlockTime = 13  // Temps de bloc cible en secondes (défaut: 13s comme Ethereum)
LWMAWindowSize = 60       // Fenêtre d'ajustement (60 blocs)
```

## Comment appliquer les changements

### 1. Recompiler geth

```bash
cd /home/ubuntu/go-Ducros
make clean
make geth
```

### 2. Choisir votre stratégie

#### Stratégie A: Réinitialiser la blockchain (RECOMMANDÉ pour dev)

```bash
# Arrêter geth
pkill -9 geth
fuser -k 30303/tcp 30303/udp 8545/tcp

# Supprimer les données de la chaîne
rm -rf devnet-data/geth/chaindata
rm -rf devnet-data/geth/lightchaindata

# Réinitialiser avec le genesis (difficulté ajustée)
./build/bin/geth init --datadir devnet-data genesis-randomx.json

# Relancer
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http --http.api eth,net,web3,randomx,personal,miner \
  --http.addr 0.0.0.0 --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.etherbase=0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
```

#### Stratégie B: Continuer avec la chaîne existante

La nouvelle difficulté minimale s'appliquera aux prochains blocs, mais l'algorithme LWMA ajustera progressivement :

```bash
# Juste recompiler et relancer
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http --http.api eth,net,web3,randomx,personal,miner \
  --http.addr 0.0.0.0 --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.etherbase=0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
```

La difficulté montera progressivement sur ~60 blocs (fenêtre LWMA).

## Vérifier la difficulté actuelle

```bash
# Via curl
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' \
  http://localhost:8545

# Regarder le champ "difficulty" dans la réponse
```

## Logs attendus après ajustement

Avec difficulté = 100,000, vous devriez voir :

```
INFO Mining new block parent=X difficulty=100000
INFO Starting to seal block number=X difficulty=100000
INFO Allocating RandomX dataset (full mode)
INFO RandomX dataset ready
INFO Starting RandomX mining goroutine
INFO RandomX mine starting block=X difficulty=100000 target=115792...
✅ Found valid nonce! block=X hash=0x...
🎉 Successfully mined block! number=X hash=0x...
```

**Plus d'erreurs `invalid timestamp`** car les blocs sont espacés de plusieurs secondes.

## Comportement de l'algorithme LWMA

L'algorithme LWMA ajuste automatiquement la difficulté pour maintenir un temps de bloc moyen de 13 secondes :

- **Si blocs trop rapides** → difficulté augmente (max 2× par bloc)
- **Si blocs trop lents** → difficulté diminue (max 2× par bloc)
- **Plancher** : Ne descend jamais sous `LWMAMinDifficulty` (maintenant 100,000)

## Résumé

**Changement appliqué :** `LWMAMinDifficulty = 1` → `100000`
**Effet :** Blocs minés en ~5-10 secondes au lieu de <1 seconde
**Prochaine étape :** Recompiler et relancer geth

Vous pouvez ajuster la valeur de 100,000 selon vos besoins :
- Plus bas (10,000) = mining plus rapide
- Plus haut (1,000,000) = mining plus lent
