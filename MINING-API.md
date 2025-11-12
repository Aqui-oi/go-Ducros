# Mining API - go-Ducros RandomX

API RPC pour le mining RandomX, compatible avec les mineurs externes (comme ethminer pour Ethereum).

---

## 📋 Vue d'ensemble

go-Ducros expose 3 endpoints RPC principaux pour le mining, exactement comme Ethereum avec Ethash :

| Endpoint | Description |
|----------|-------------|
| `eth_getWork` | Récupère un job de mining |
| `eth_submitWork` | Soumet une solution PoW |
| `eth_submitHashrate` | Rapporte le hashrate du mineur |

Ces endpoints sont disponibles dans les namespaces **`eth`** et **`randomx`**.

---

## 🔌 Endpoints RPC

### 1. `eth_getWork` / `randomx_getWork`

Récupère le travail de mining actuel.

**Paramètres:** Aucun

**Retour:** `[4]string`

```json
[
  "0x1234...",  // [0] Header hash (32 bytes) - hash du bloc sans nonce
  "0xabcd...",  // [1] Seed hash (32 bytes) - ParentHash pour RandomX cache
  "0x0000...",  // [2] Target (32 bytes) - boundary condition (2^256/difficulty)
  "0x42"        // [3] Block number (hex)
]
```

**Exemple curl:**

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
  http://localhost:8545
```

**Réponse:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
    "0x0000000112e0be826d694b2e62d01511f12a6061fbaec8bc02357593e70e52ba",
    "0x10"
  ]
}
```

---

### 2. `eth_submitWork` / `randomx_submitWork`

Soumet une solution PoW trouvée.

**Paramètres:**

- `nonce` (8 bytes hex) - Le nonce trouvé
- `headerHash` (32 bytes hex) - Hash du header (de getWork[0])
- `mixDigest` (32 bytes hex) - Hash RandomX calculé

**Retour:** `boolean` - `true` si accepté, `false` sinon

**Exemple curl:**

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc":"2.0",
    "method":"eth_submitWork",
    "params":[
      "0x0000000000000042",
      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "0x9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba"
    ],
    "id":1
  }' \
  http://localhost:8545
```

**Réponse:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": true
}
```

---

### 3. `eth_submitHashrate` / `randomx_submitHashrate`

Rapporte le hashrate du mineur.

**Paramètres:**

- `hashrate` (hex) - Hashrate en H/s
- `id` (32 bytes hex) - ID unique du mineur

**Retour:** `boolean` - `true` si accepté

**Exemple curl:**

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc":"2.0",
    "method":"eth_submitHashrate",
    "params":[
      "0x500",
      "0x59daa26581d0acd1fce254fb7e85952f4c09d0915afd33d3886cd914bc7d283c"
    ],
    "id":1
  }' \
  http://localhost:8545
```

**Réponse:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": true
}
```

---

### 4. `eth_hashrate` / `randomx_hashrate` (lecture)

Récupère le hashrate total du réseau (local + remote miners).

**Paramètres:** Aucun

**Retour:** `string` - Hashrate en H/s (hex)

**Exemple curl:**

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545
```

**Réponse:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x1f40"  // 8000 H/s
}
```

---

## 🔧 Configuration geth

### Activer l'API mining

```bash
./geth \
  --datadir ./data \
  --http \
  --http.addr "0.0.0.0" \
  --http.port 8545 \
  --http.api "eth,net,web3,randomx" \
  --mine \
  --miner.etherbase 0xYourAddress
```

**Flags importants:**

- `--http` : Active le serveur HTTP RPC
- `--http.api "eth,randomx"` : Expose les namespaces eth et randomx
- `--mine` : Active le mining
- `--miner.etherbase` : Adresse qui reçoit les rewards

---

## 💻 Intégration Mineur

### Format RandomX

L'input RandomX est construit comme suit :

```
Input (40 bytes) = SealHash (32 bytes) + Nonce (8 bytes, little-endian)
```

**Processus:**

1. Récupérer le work avec `eth_getWork`
2. Extraire : `headerHash`, `seedHash`, `target`
3. Initialiser RandomX cache avec `seedHash` (ParentHash)
4. Boucler sur les nonces :
   ```
   input = headerHash + nonce (LE)
   hash = RandomX(input)
   if hash <= target: soumettre avec eth_submitWork
   ```

### Pseudo-code mineur

```python
while True:
    # 1. Récupérer le travail
    work = rpc_call("eth_getWork")
    header_hash = work[0]
    seed_hash = work[1]
    target = work[2]

    # 2. Init RandomX cache
    rx_cache = randomx_init_cache(seed_hash)
    rx_vm = randomx_create_vm(rx_cache)

    # 3. Mining loop
    nonce = random_uint64()
    while True:
        # Construire input : headerHash (32) + nonce (8 LE)
        input = header_hash + nonce.to_bytes(8, 'little')

        # Calculer hash RandomX
        hash = randomx_calculate_hash(rx_vm, input)

        # Vérifier si solution valide
        if hash <= target:
            mix_digest = hash
            result = rpc_call("eth_submitWork", [nonce, header_hash, mix_digest])
            if result:
                print("Block found!")
            break

        nonce += 1

        # Check nouveau work tous les 1000 nonces
        if nonce % 1000 == 0:
            new_work = rpc_call("eth_getWork")
            if new_work[0] != header_hash:
                break  # Nouveau bloc, restart
```

---

## 🧪 Tests

### Test 1: Vérifier que le mining RPC fonctionne

```bash
# Lancer geth avec mining activé
./geth --datadir ./data --http --http.api eth,randomx --mine --miner.threads 1

# Dans un autre terminal, récupérer le work
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getWork","params":[],"id":1}' \
  http://localhost:8545

# Devrait retourner [headerHash, seedHash, target, blockNumber]
```

### Test 2: Soumettre une solution invalide

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc":"2.0",
    "method":"eth_submitWork",
    "params":[
      "0x0000000000000001",
      "0x0000000000000000000000000000000000000000000000000000000000000000",
      "0x0000000000000000000000000000000000000000000000000000000000000000"
    ],
    "id":1
  }' \
  http://localhost:8545

# Devrait retourner false
```

### Test 3: Rapporter hashrate

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc":"2.0",
    "method":"eth_submitHashrate",
    "params":[
      "0x1000",
      "0x59daa26581d0acd1fce254fb7e85952f4c09d0915afd33d3886cd914bc7d283c"
    ],
    "id":1
  }' \
  http://localhost:8545

# Devrait retourner true
```

---

## 🔍 Troubleshooting

### Erreur: `method not found`

**Cause:** L'API mining n'est pas exposée.

**Solution:**

```bash
./geth --http.api "eth,randomx,net,web3"
```

### Erreur: `no mining work available yet`

**Cause:** Le mining n'est pas démarré ou aucun bloc en attente.

**Solution:**

```bash
# Vérifier que le mining est actif
./geth --mine --miner.threads 4
```

### Erreur: `invalid or stale proof-of-work solution`

**Cause:** La solution est incorrecte ou le work a changé.

**Solutions:**

1. Vérifier que le RandomX cache est initialisé avec le bon `seedHash`
2. Vérifier que l'input est bien : `headerHash + nonce (LE)`
3. Le work peut avoir changé (nouveau bloc) - récupérer un nouveau work

---

## 📊 Monitoring

### Métriques à surveiller

```bash
# Hashrate total
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result'

# Mining actif ?
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result'

# Block number
curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result'
```

---

## 🔗 Compatibilité

### Mineurs compatibles (avec adaptation)

| Mineur | Statut | Notes |
|--------|--------|-------|
| **ethminer** | ⚠️ Adapter | Change Ethash → RandomX |
| **XMRig** | ✅ Via proxy | Nécessite proxy Stratum → RPC |
| **SRBMiner** | ✅ Via proxy | Nécessite proxy Stratum → RPC |
| **Mineur custom** | ✅ Direct | Utilise directement les RPC |

### Proxy Stratum recommandé

Pour utiliser XMRig/SRBMiner, un proxy Stratum → RPC est nécessaire :

```
XMRig (Stratum) → Proxy → go-Ducros (RPC)
```

**Projet proxy à venir:** `ducros-stratum-proxy`

---

## 📝 Notes Importantes

1. **RandomX JIT est désactivé** par défaut dans go-Ducros (ligne randomx.go:195)
   - Raison : Éviter segfaults sur certains systèmes
   - Mode interprété uniquement (plus lent mais stable)

2. **Work timeout:** Le work est valide jusqu'au prochain bloc

3. **Hashrate reporting:** Les mineurs doivent rapporter leur hashrate régulièrement (recommandé : toutes les 5-10 secondes)

4. **Nonce format:** Little-endian obligatoire (comme Monero RandomX)

---

## 🆘 Support

Pour les problèmes d'intégration :

1. Vérifier les logs geth : `--verbosity 4`
2. Tester avec curl d'abord
3. Vérifier que RandomX library est installée
4. Consulter `VERIFYSEAL-LWMA-GUIDE.md` pour les détails RandomX

---

**Version:** 1.0.0
**Date:** 2025-11-12
**Branche:** `claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi`
