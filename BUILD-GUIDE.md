# Build Guide - go-Ducros RandomX

Guide de compilation pour go-Ducros avec le consensus RandomX + LWMA.

---

## 📋 Prérequis

### Système
- **OS:** Linux (Ubuntu 20.04+, Debian 11+, ou équivalent)
- **CPU:** x86_64 avec support AVX2 (recommandé)
- **RAM:** 4GB minimum, 8GB recommandé
- **Disk:** 20GB libre minimum

### Outils de développement
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y build-essential git cmake

# Vérifier les versions
gcc --version    # >= 9.0
g++ --version    # >= 9.0
cmake --version  # >= 3.16
go version       # >= 1.21
```

---

## 🔧 Installation RandomX

**IMPORTANT:** RandomX doit être installé **AVANT** de compiler geth.

### Étape 1: Cloner RandomX

```bash
cd /tmp
git clone https://github.com/tevador/RandomX.git
cd RandomX
```

### Étape 2: Compiler RandomX

```bash
mkdir build && cd build
cmake -DARCH=native ..
make -j$(nproc)
```

**Output attendu:**
```
[100%] Built target randomx
[100%] Built target randomx-tests
[100%] Built target randomx-benchmark
```

### Étape 3: Installer RandomX

```bash
sudo make install
sudo ldconfig
```

**Vérifier l'installation:**
```bash
ls -la /usr/local/lib/librandomx.a
ls -la /usr/local/include/randomx.h
```

Les deux fichiers doivent exister.

---

## 🚀 Compilation go-Ducros

### Étape 1: Cloner le repo

```bash
cd ~
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros
```

### Étape 2: Checkout la branche RandomX

```bash
git checkout claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi
```

### Étape 3: Configurer les flags CGO

**CRITIQUE:** Les flags CGO doivent pointer vers RandomX.

```bash
export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/usr/local/include"
```

**Pour rendre permanent (optionnel):**
```bash
echo 'export CGO_LDFLAGS="-L/usr/local/lib"' >> ~/.bashrc
echo 'export CGO_CFLAGS="-I/usr/local/include"' >> ~/.bashrc
source ~/.bashrc
```

### Étape 4: Compiler geth

```bash
make geth
```

**Durée:** 5-10 minutes selon la machine.

**Output attendu:**
```
>>> /usr/local/go/bin/go build ...
Done building.
Run "./build/bin/geth" to launch geth.
```

### Étape 5: Vérifier le binaire

```bash
./build/bin/geth version
```

**Output attendu:**
```
Geth
Version: 1.16.7-stable
Git Commit: 6f761f29
Architecture: amd64
Go Version: go1.21.x
Operating System: linux
```

---

## 🔍 Troubleshooting

### Erreur: `cannot find -lrandomx`

**Cause:** RandomX n'est pas installé ou le linker ne le trouve pas.

**Solution:**
```bash
# Vérifier que RandomX est installé
ls /usr/local/lib/librandomx.a

# Si absent, réinstaller RandomX (voir section Installation RandomX)

# Si présent, vérifier les flags CGO
echo $CGO_LDFLAGS
# Doit afficher: -L/usr/local/lib

# Relancer la compilation
make clean
export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/usr/local/include"
make geth
```

### Erreur: `undefined reference to randomx_*`

**Cause:** Mauvaise version de RandomX ou bibliothèque corrompue.

**Solution:**
```bash
cd /tmp/RandomX/build
sudo make uninstall
sudo make install
sudo ldconfig
```

### Erreur: Dépendances Go ne se téléchargent pas

**Cause:** Pas de connexion internet ou proxy Go mal configuré.

**Solution 1: Vérifier la connexion**
```bash
ping 8.8.8.8
curl https://proxy.golang.org
```

**Solution 2: Utiliser un proxy Go**
```bash
export GOPROXY=https://proxy.golang.org,direct
make geth
```

**Solution 3: Télécharger les dépendances en avance**
```bash
go mod download
make geth
```

### Erreur: Compilation très lente

**Cause:** Pas assez de CPU ou RAM.

**Solution:**
```bash
# Limiter les jobs parallèles
GOMAXPROCS=2 make geth
```

---

## 🧪 Tests de Compilation

### Test 1: Vérifier que RandomX est bien linké

```bash
ldd ./build/bin/geth | grep randomx
```

**Output attendu:**
```
(devrait montrer librandomx.a ou rien si statiquement linké)
```

### Test 2: Vérifier que les tests passent

```bash
# Tests RandomX consensus
go test -v ./consensus/randomx -run TestLWMABasic

# Tests VerifySeal
go test -v ./consensus/randomx -run TestVerifySeal
```

### Test 3: Lancer geth en mode dev

```bash
./build/bin/geth --datadir /tmp/test-data --dev console
```

Dans la console:
```javascript
> eth.blockNumber
0
> miner.start(1)
null
> eth.blockNumber
// Devrait augmenter
```

---

## 📦 Compilation pour Distribution

### Build statique (recommandé pour déploiement)

```bash
CGO_ENABLED=1 \
CGO_LDFLAGS="-L/usr/local/lib -static" \
CGO_CFLAGS="-I/usr/local/include" \
go build -ldflags "-linkmode external -extldflags -static" \
-o ./build/bin/geth-static ./cmd/geth
```

### Build optimisé pour production

```bash
CGO_ENABLED=1 \
CGO_LDFLAGS="-L/usr/local/lib" \
CGO_CFLAGS="-I/usr/local/include -O3 -march=native" \
go build -ldflags "-s -w" \
-o ./build/bin/geth-optimized ./cmd/geth
```

### Build pour différentes architectures

**AMD64:**
```bash
GOARCH=amd64 make geth
```

**ARM64 (cross-compile - nécessite toolchain):**
```bash
# Installer cross-compiler
sudo apt-get install gcc-aarch64-linux-gnu

# Compiler RandomX pour ARM64 d'abord
# Puis compiler geth
CC=aarch64-linux-gnu-gcc \
GOARCH=arm64 \
make geth
```

---

## 🔐 Build Reproductible

Pour garantir la reproductibilité:

```bash
# Fixer la version de Go
export GOVERSION=1.21.5

# Fixer les dépendances
go mod tidy
go mod verify

# Compiler avec flags déterministes
make geth
```

Le Makefile utilise déjà `--buildid=none` et `--strip-all`.

---

## 📊 Benchmarks Post-Compilation

### Benchmark RandomX

```bash
# Depuis le répertoire RandomX
cd /tmp/RandomX/build
./randomx-benchmark

# Output attendu (exemple sur CPU moderne):
# RandomX light mode   | 15000 H/s
# RandomX fast mode    | 25000 H/s
```

### Benchmark VerifySeal

```bash
cd ~/go-Ducros
go test -bench=BenchmarkVerifySeal ./consensus/randomx
```

### Benchmark LWMA

```bash
go test -bench=BenchmarkLWMA ./consensus/randomx
```

---

## 🚢 Déploiement

### Copier le binaire sur le serveur de production

```bash
# Depuis la machine de build
scp ./build/bin/geth user@production-server:/usr/local/bin/geth-ducros

# Sur le serveur de production
sudo chmod +x /usr/local/bin/geth-ducros
/usr/local/bin/geth-ducros version
```

### Vérifier les dépendances sur le serveur

```bash
# RandomX doit être installé sur le serveur de prod aussi
ssh user@production-server
ls /usr/local/lib/librandomx.a

# Si absent, installer RandomX (voir section Installation RandomX)
```

---

## 🔄 Mise à jour

### Mettre à jour go-Ducros

```bash
cd ~/go-Ducros
git pull origin claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi
make clean
make geth
```

### Mettre à jour RandomX (rare)

```bash
cd /tmp/RandomX
git pull
cd build
cmake -DARCH=native ..
make -j$(nproc)
sudo make install
sudo ldconfig

# Recompiler geth
cd ~/go-Ducros
make clean
make geth
```

---

## 📝 Notes Importantes

1. **RandomX JIT est DÉSACTIVÉ** par défaut dans notre implémentation pour éviter les segfaults. C'est intentionnel et documenté dans `consensus/randomx/randomx.go:195`.

2. **LWMA est ACTIVÉ** automatiquement si `randomx: {}` est présent dans genesis.json.

3. **MinimumDifficulty = 1** pour permettre un démarrage rapide. Augmenter pour production si nécessaire.

4. **Les tests unitaires** ne nécessitent PAS RandomX library car ils utilisent le mode fake.

---

## 🆘 Support

Si vous rencontrez des problèmes:

1. Vérifier cette checklist:
   - [ ] RandomX est bien installé (`ls /usr/local/lib/librandomx.a`)
   - [ ] Les flags CGO sont configurés (`echo $CGO_LDFLAGS`)
   - [ ] Go version >= 1.21 (`go version`)
   - [ ] Connexion internet pour dépendances Go

2. Consulter les logs détaillés:
   ```bash
   make geth 2>&1 | tee build.log
   ```

3. Tester RandomX indépendamment:
   ```bash
   cd /tmp/RandomX/build
   ./randomx-tests
   ```

---

**Version:** 1.0.0
**Date:** 2025-11-12
**Branche:** `claude/ducros-randomx-review-011CV3cgBsT5BT8d6UQNiFMi`
**Commit:** `6f761f2`
