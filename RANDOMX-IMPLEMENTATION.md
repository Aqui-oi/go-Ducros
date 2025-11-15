# Go-Ducros: Geth Fork with RandomX PoW

## 📋 Vue d'ensemble

Ce fork de go-ethereum (Geth) v1.16.7 remplace le consensus Beacon (PoS) par **RandomX** - un algorithme de Proof-of-Work CPU-friendly et ASIC-resistant, utilisé avec succès par Monero.

## ✨ Caractéristiques

### 🔧 Consensus RandomX
- **CPU-Friendly**: Mining accessible sans GPU/ASIC
- **ASIC-Resistant**: Démocratisation du mining
- **Prouvé en production**: Utilisé par Monero depuis 2019
- **Mémoire-intensive**: ~2GB RAM par VM pour la sécurité

### 💎 Compatibilité Ethereum
- **EVM complet**: Tous les smart contracts Ethereum fonctionnent
- **RPC standards**: Compatible Metamask, Remix, Hardhat, etc.
- **Block rewards**: Identiques à Ethereum (5 ETH → 3 ETH → 2 ETH)
- **Difficulty adjustment**: Algorithme Ethereum standard
- **Uncle rewards**: Support complet

## 🏗️ Architecture

### Structure du code

```
consensus/randomx/
├── randomx.go          # Core logic, C bindings, VM pooling
├── difficulty.go       # Algorithme de difficulté (Frontier → Constantinople)
├── consensus.go        # Implémentation de consensus.Engine
└── consensus_test.go   # Tests unitaires
```

### Fichiers modifiés

- `params/config.go`: Ajout de RandomXConfig
- `eth/ethconfig/config.go`: Logique de sélection du consensus engine

### Fichiers inchangés (désactivés proprement)

- `consensus/beacon/`: Commenté, non supprimé
- `consensus/ethash/`: Conservé pour référence
- `consensus/clique/`: Disponible pour testnets privés

## 📊 Paramètres Économiques

### Block Rewards (identiques à Ethereum)
- **Frontier**: 5 ETH par bloc
- **Byzantium**: 3 ETH par bloc
- **Constantinople**: 2 ETH par bloc (actuel)

### Chain Configuration
- **Chain ID**: 33669
- **Block time**: ~13 secondes (ajustable via difficulty)
- **Gas limit**: 4,700,000 (identique à Ethereum)
- **Minimum difficulty**: 131,072

### Difficulty Bomb
- Support complet de la difficulty bomb Ethereum
- Delays configurables (EIP-649, EIP-1234, EIP-2384, etc.)

## 🚀 Démarrage Rapide

### Prérequis

```bash
# Go 1.23+
go version

# Bibliothèques RandomX (pour le mining réel)
# sudo apt-get install librandomx-dev  # TODO: À compiler ou installer
```

### Compilation

```bash
# Build geth
make geth

# Ou build complet
make all
```

### Initialiser la blockchain

```bash
# Utiliser le genesis RandomX
./build/bin/geth init genesis-randomx.json --datadir ./data-randomx
```

### Lancer un nœud

```bash
# Nœud de développement
./build/bin/geth \
  --datadir ./data-randomx \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,personal,miner \
  --allow-insecure-unlock \
  --nodiscover \
  --maxpeers 0

# Dans un autre terminal: commencer le mining
./build/bin/geth attach ./data-randomx/geth.ipc
> miner.start(1)  # 1 thread
```

## 🔬 Tests

```bash
# Tests du consensus RandomX
cd consensus/randomx
go test -v

# Tests de difficulté
go test -bench=. -benchmem

# Tests d'intégration
cd ../..
go test ./...
```

## 📝 Configuration Genesis

### Exemple minimal

```json
{
  "config": {
    "chainId": 33669,
    "homesteadBlock": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "randomx": {}
  },
  "difficulty": "0x20000",
  "gasLimit": "0x47b760",
  "alloc": {}
}
```

### Options avancées

```json
{
  "config": {
    "chainId": 33669,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "randomx": {},
    "terminalTotalDifficulty": null
  }
}
```

**Important**: `terminalTotalDifficulty` doit être `null` ou omis pour RandomX (pas de passage à PoS).

## 🔐 Sécurité

### Considérations

1. **51% Attack**: Le hashrate initial sera faible, vulnérable
2. **Network Bootstrap**: Démarrer avec des nœuds de confiance
3. **Difficulty Adjustment**: Peut être volatile au début

### Recommandations

- Déployer d'abord un testnet privé
- Monitorer le hashrate réseau
- Ajuster les paramètres de difficulté si nécessaire

## 🛠️ Développement

### TODO: RandomX C Bindings

Actuellement, le code utilise un **fake engine** pour les tests. Pour activer le vrai mining RandomX:

1. Compiler la bibliothèque RandomX:
```bash
git clone https://github.com/tevador/RandomX.git
cd RandomX
mkdir build && cd build
cmake -DARCH=native ..
make
sudo make install
```

2. Modifier `eth/ethconfig/config.go`:
```go
// Remplacer
return ethash.NewFaker(), nil

// Par
return randomx.New(nil), nil
```

3. Recompiler:
```bash
go build -tags randomx ./cmd/geth
```

### Structure de test

Le code suit la structure d'Ethash pour faciliter la maintenance:

- Tests unitaires de difficulté ✅
- Benchmarks de performance ✅
- Tests de validation de headers ✅
- Tests d'intégration ⏳ (TODO)

## 📚 Références

### RandomX
- [RandomX Specs](https://github.com/tevador/RandomX)
- [Monero Implementation](https://github.com/monero-project/monero)
- [RandomX Audit](https://ostif.org/our-audit-of-randomx-is-complete/)

### Ethereum
- [Difficulty Algorithm](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md)
- [Block Rewards](https://eips.ethereum.org/EIPS/eip-1234)
- [Consensus Engine Interface](https://github.com/ethereum/go-ethereum/blob/master/consensus/consensus.go)

## 🤝 Contribution

Ce projet est un fork expérimental de Geth. Contributions bienvenues !

### Guidelines

1. Suivre la structure de code Ethereum
2. Ajouter des tests pour toute nouvelle fonctionnalité
3. Documenter les changements dans les commentaires
4. Ne pas supprimer le code existant, le commenter si nécessaire

## 📜 License

Identique à go-ethereum: LGPL-3.0

## ⚠️ Disclaimer

Ce projet est **expérimental** et **non audité**. Ne pas utiliser en production sans une revue de sécurité complète.

Les block rewards et paramètres économiques sont identiques à Ethereum pour faciliter les tests et la comparaison.

---

**Construit avec ❤️ pour la décentralisation du mining**
