# Fix pour le segfault RandomX

## Problème

Le segmentation fault (`SIGSEGV`) se produit dans `randomx_init_dataset` lors de l'initialisation du dataset RandomX.

### Cause racine

La bibliothèque RandomX (écrite en C) crée plusieurs threads internes (pthreads) pour initialiser le dataset de ~2GB en parallèle. Le problème principal était que la goroutine Go qui appelle `C.randomx_init_dataset` n'était **pas verrouillée à un thread OS** (`runtime.LockOSThread`).

Sans ce verrouillage, le scheduler Go peut déplacer la goroutine entre différents threads OS pendant que la bibliothèque C a des pthreads actifs qui essaient d'accéder à la mémoire. Cela crée un conflit de contexte d'exécution qui résulte en un SIGSEGV lors de l'accès à la mémoire du dataset.

**En résumé :** Quand Go déplace la goroutine d'un thread OS à un autre, les pthreads créés par le code C perdent leur contexte d'exécution et causent une violation de segmentation.

## Solutions appliquées

### 1. Verrouillage du thread OS (FAIT ✓) **[SOLUTION CRITIQUE]**

Ajout de `runtime.LockOSThread()` et `defer runtime.UnlockOSThread()` dans la goroutine qui appelle `C.randomx_init_dataset`. Ceci garantit que:
- La goroutine reste sur le même thread OS pendant toute l'exécution
- Les pthreads créés par RandomX C gardent leur contexte d'exécution valide
- Pas de conflit entre le scheduler Go et les threads natifs C

### 2. Validation des pointeurs (FAIT ✓)

Ajout de validation avant et pendant les appels C pour s'assurer que les pointeurs `dataset` et `cache` restent valides.

### 3. Protection contre les panics (FAIT ✓)

Ajout d'un `defer recover()` dans la goroutine d'initialisation du dataset pour capturer les panics et éviter le crash complet.

### 4. Avertissement GOMAXPROCS (FAIT ✓)

Le code détecte maintenant si `GOMAXPROCS=1` est utilisé et affiche un warning explicite :

```
WARN RandomX dataset initialization with GOMAXPROCS=1 may cause instability
```

### 5. Import de runtime (FAIT ✓)

Ajout de l'import `runtime` pour détecter GOMAXPROCS et verrouiller les threads OS.

## Solutions pour l'utilisateur

### Option 1 : Recompiler avec le fix (RECOMMANDÉ)

Le fix a été appliqué au code. Il suffit de recompiler geth :

```bash
make clean
make geth
```

Puis lancer normalement (fonctionne maintenant avec ou sans GOMAXPROCS) :

```bash
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http --http.api eth,net,web3,randomx,personal,miner \
  --http.addr 0.0.0.0 --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.etherbase=0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
```

Le segfault ne devrait plus se produire car la goroutine est maintenant correctement verrouillée à un thread OS.

### Option 2 : Activer les Huge Pages (BONUS pour performance)

Pour de meilleures performances, activez les huge pages sur votre système :

```bash
# Allouer 1280 huge pages (environ 2.5 GB)
sudo sysctl -w vm.nr_hugepages=1280

# Pour rendre permanent
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf
```

Cela éliminera aussi le warning :
```
WARN RandomX using JIT without huge pages (performance -30%)
```

## Test du fix

Après avoir recompilé avec le fix :

```bash
# 1. Recompiler
make clean
make geth

# 2. Tester (fonctionne maintenant avec ou sans GOMAXPROCS)
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http --http.api eth,net,web3,randomx,personal,miner \
  --http.addr 0.0.0.0 --http.port 8545 \
  --http.corsdomain "*" \
  --mine \
  --miner.etherbase=0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
```

Vous devriez voir :

```
INFO Allocating RandomX dataset (full mode)
INFO Initializing RandomX dataset in background items=34,078,719 seed=0x...
INFO RandomX dataset ready seed=0x... duration=XXs
INFO Starting RandomX mining goroutine
INFO RandomX VM created, starting nonce search...
✅ Found valid nonce! block=X
🎉 Successfully mined block! number=X hash=0x...
```

Au lieu du segfault `SIGSEGV: segmentation violation`.

## Changements dans le code

Les modifications ont été appliquées dans `consensus/randomx/randomx.go` :

1. **Ligne 69** : Ajout de `import "runtime"`
2. **Lignes 502-508** : Détection et warning pour GOMAXPROCS=1
3. **Lignes 522-527** : **[CRITIQUE]** Ajout de `runtime.LockOSThread()` / `UnlockOSThread()`
4. **Lignes 529-535** : Protection panic recovery améliorée
5. **Lignes 537-566** : Validation des pointeurs avant/pendant les appels C
6. **Ligne 570** : Appel C sécurisé avec thread verrouillé

## Détails techniques

### Pourquoi `runtime.LockOSThread()` est critique

Quand un programme Go appelle du code C via CGO :
1. La goroutine Go s'exécute normalement sur différents threads OS (le scheduler Go la déplace)
2. Si le code C crée des pthreads (comme RandomX le fait), ces threads sont liés au thread OS actuel
3. **Problème** : Si Go déplace la goroutine vers un autre thread OS pendant que les pthreads C sont actifs, les pthreads perdent leur contexte et tentent d'accéder à de la mémoire invalide
4. **Solution** : `runtime.LockOSThread()` force la goroutine à rester sur le même thread OS, garantissant que les pthreads C restent valides

### Stack trace du segfault (avant le fix)

```
PC=0x7598b278a01c m=8 sigcode=1 addr=0x7598158e8000
signal arrived during cgo execution

goroutine 8507 [syscall]:
runtime.cgocall(0x17364b0, 0xc001b8ef70)
C.randomx_init_dataset(...)
github.com/ethereum/go-ethereum/consensus/randomx.(*RandomX).buildDataset.func1()
    randomx.go:540
```

Le crash se produisait car la goroutine n'était pas verrouillée, permettant au scheduler Go de la déplacer pendant que `randomx_init_dataset` avait des pthreads actifs.

## Résumé

**Cause racine :** Goroutine Go non verrouillée à un thread OS + pthreads C créés par RandomX = conflit de contexte d'exécution
**Fix critique :** `runtime.LockOSThread()` dans la goroutine qui appelle `C.randomx_init_dataset`
**Fixes additionnels :** Validation des pointeurs, panic recovery, warning GOMAXPROCS
**Résultat :** Le segfault ne devrait plus se produire, quel que soit GOMAXPROCS
