# 🪟 Configuration xmrig pour Windows - Ducros Network

Guide pour miner Ducros depuis Windows vers votre serveur **92.222.10.107**

---

## 📥 Étape 1: Télécharger xmrig sur Windows

1. Allez sur https://github.com/xmrig/xmrig/releases/latest
2. Téléchargez **xmrig-X.XX.X-msvc-win64.zip** (version Windows 64-bit)
3. Extrayez le ZIP dans un dossier (ex: `C:\xmrig-ducros\`)

---

## 📋 Étape 2: Copier les fichiers de configuration

Copiez ces 2 fichiers depuis le serveur vers votre dossier xmrig sur Windows :

- `xmrig-windows-config.json` → Placez-le dans `C:\xmrig-ducros\`
- `start-xmrig-ducros.bat` → Placez-le dans `C:\xmrig-ducros\`

### Option A: Télécharger via SCP (depuis Windows avec WinSCP ou via WSL)

```bash
# Depuis WSL ou PowerShell avec scp installé:
scp ubuntu@92.222.10.107:/home/user/go-Ducros/xmrig-windows-config.json C:\xmrig-ducros\
scp ubuntu@92.222.10.107:/home/user/go-Ducros/start-xmrig-ducros.bat C:\xmrig-ducros\
```

### Option B: Copier manuellement le contenu

Créez ces fichiers manuellement dans `C:\xmrig-ducros\` avec le contenu fourni.

---

## 🚀 Étape 3: Lancer le Serveur (sur 92.222.10.107)

### 3.1: Compiler Geth (si pas déjà fait)

```bash
cd /home/user/go-Ducros
make geth
```

### 3.2: Lancer Geth

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
  --mine \
  --miner.threads 1 \
  --miner.etherbase 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2 \
  --verbosity 3
```

**⚠️ IMPORTANT:** Utilisez `--http.addr "0.0.0.0"` pour accepter les connexions externes.

### 3.3: Lancer le Stratum Proxy (dans un autre terminal)

```bash
cd /home/user/go-Ducros/stratum-proxy
./stratum-proxy -geth "http://localhost:8545" -stratum "0.0.0.0:3333" -v
```

**Sortie attendue :**
```
╔═══════════════════════════════════════════════════════════╗
║        Stratum Proxy - RandomX Mining Bridge             ║
╚═══════════════════════════════════════════════════════════╝
⚠️  WARNING: No pool address specified, using miner addresses directly
✅ Connected to Geth at http://localhost:8545
🌐 Stratum server listening on 0.0.0.0:3333
📊 Waiting for miners...
```

---

## ⛏️ Étape 4: Lancer xmrig sur Windows

1. Ouvrez l'Explorateur Windows et allez dans `C:\xmrig-ducros\`
2. **Double-cliquez sur `start-xmrig-ducros.bat`**

**Sortie attendue :**
```
========================================
  Ducros Network - RandomX Mining
========================================

Serveur Stratum: 92.222.10.107:3333
Wallet: 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2

Demarrage de xmrig...

[2025-11-12 14:00:00.000]  * ABOUT        XMRig/6.21.3 msvc/2022
[2025-11-12 14:00:00.000]  * LIBS         libuv/1.48.0 OpenSSL/3.2.0 hwloc/2.9.3
[2025-11-12 14:00:00.000]  * HUGE PAGES   supported
[2025-11-12 14:00:05.000]  net      use pool 92.222.10.107:3333  rx/0
[2025-11-12 14:00:05.000]  net      new job from 92.222.10.107:3333 diff 10000
[2025-11-12 14:00:10.000]  cpu      use profile rx
[2025-11-12 14:00:10.000]  cpu      READY threads 8/8 (8) huge pages 0%
[2025-11-12 14:00:30.000]  miner    speed 10s/60s/15m 1234.5 1234.5 n/a H/s
```

---

## 🔥 Étape 5: Vérifier le Mining

### Sur le Serveur (Stratum Proxy)

Vous devriez voir dans les logs du proxy :
```
📊 New miner connected: windows-pc (worker: ducros-windows-miner)
✅ Share accepted from windows-pc (diff: 10000)
📊 Stats: Miners=1/1 Shares=1/0/1 Blocks=0 Hashrate=1234.50 H/s
```

### Via Console Geth

```bash
./build/bin/geth attach ~/.ducros/geth.ipc

> eth.mining
true

> eth.hashrate
1234567  // Devrait être > 0 maintenant

> eth.blockNumber
5  // Devrait augmenter

> eth.getBlock("latest")
{
  difficulty: 1024,
  miner: "0x25ffa18fb7e35e0a3272020305f4bea0b770a7f2",
  number: 5
}
```

---

## 🔒 Sécurité Important

### Firewall Serveur

Autorisez uniquement les ports nécessaires :

```bash
# P2P Geth
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# Stratum pour mining externe
sudo ufw allow 3333/tcp

# RPC Geth (optionnel, NE PAS exposer publiquement)
# sudo ufw allow from VOTRE_IP_WINDOWS to any port 8545
```

### ⚠️ NE PAS exposer le port 8545 publiquement !

Le port 8545 (RPC Geth) doit rester en localhost sauf si vous utilisez un tunnel SSH ou un VPN.

---

## 🐛 Dépannage

### Erreur: "connect: connection refused"

**Cause:** Le Stratum proxy n'est pas accessible depuis Windows.

**Solutions:**
1. Vérifiez le firewall du serveur: `sudo ufw status`
2. Vérifiez que le proxy écoute bien sur `0.0.0.0:3333`: `netstat -tlnp | grep 3333`
3. Testez depuis Windows: `telnet 92.222.10.107 3333`

### Erreur: "socket error" ou timeout

**Cause:** Firewall bloque le port 3333.

**Solution:**
```bash
sudo ufw allow 3333/tcp
```

### xmrig dit "accepted (0/0)" mais pas de blocks

**Cause:** La difficulté du réseau est trop haute pour votre hashrate.

**Solution:** Attendez, ou baissez la difficulté initiale dans genesis si c'est un testnet.

### Huge pages warning sur Windows

**Solution:** Lancez xmrig en tant qu'**Administrateur** (clic droit → Exécuter en tant qu'administrateur).

---

## 📊 Performance Attendue (RandomX)

| CPU                  | Hashrate    |
|---------------------|-------------|
| Ryzen 9 5950X       | ~15,000 H/s |
| Intel i9-12900K     | ~18,000 H/s |
| Ryzen 7 5800X       | ~10,000 H/s |
| Intel i7-10700K     | ~8,000 H/s  |
| Ryzen 5 5600X       | ~6,000 H/s  |

**Note:** Activez les huge pages sur Windows pour +10-15% de performance (nécessite droits admin).

---

## 🎯 Architecture Réseau

```
┌─────────────────┐         ┌──────────────────┐
│   Windows PC    │         │ Serveur Linux    │
│  92.222.X.X     │         │ 92.222.10.107    │
│                 │         │                  │
│  ┌───────────┐  │  Port   │  ┌────────────┐  │
│  │  xmrig    │──┼──3333──→│  │  Stratum   │  │
│  │  (miner)  │  │         │  │  Proxy     │  │
│  └───────────┘  │         │  └─────┬──────┘  │
│                 │         │        │ RPC     │
└─────────────────┘         │        ↓ 8545    │
                            │  ┌────────────┐  │
                            │  │   Geth     │  │
                            │  │ (blockchain)│ │
                            │  └────────────┘  │
                            │                  │
                            └──────────────────┘
```

---

**Wallet Ducros:** `0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2`
**Serveur Stratum:** `92.222.10.107:3333`
**ChainID:** 9999
**Algo:** rx/0 (RandomX)

Bon mining ! 🚀
