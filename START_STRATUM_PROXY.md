# Guide de démarrage du Stratum-Proxy Ducros

## 📋 Prérequis

1. **Geth doit être en train de miner**
   ```bash
   # Vérifier que geth tourne avec --mine et --http.api contient 'miner'
   ps aux | grep geth
   ```

2. **L'API RPC de geth doit être accessible**
   ```bash
   # Test rapide
   curl -X POST -H "Content-Type: application/json" \
     --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' \
     http://localhost:8545
   ```

## 🚀 Démarrage rapide

### Méthode 1: Compilation et lancement manuel (RECOMMANDÉ)

```bash
# Sur le VPS (92.222.10.107)
cd ~/go-Ducros/stratum-proxy

# Compiler le stratum-proxy
go build -o stratum-proxy .

# Lancer le proxy
./stratum-proxy \
  --stratum 0.0.0.0:3333 \
  --geth http://localhost:8545 \
  --diff 30000 \
  --algo rx/0 \
  -v
```

**Paramètres importants :**
- `--stratum 0.0.0.0:3333` : Écoute sur toutes les interfaces, port 3333
- `--geth http://localhost:8545` : URL du RPC geth
- `--diff 30000` : Difficulté initiale (correspond à votre nouveau LWMAMinDifficulty)
- `--algo rx/0` : Algorithme RandomX standard
- `-v` : Mode verbose pour voir tous les logs

### Méthode 2: Utiliser le script de déploiement

```bash
cd ~/go-Ducros

# Rendre le script exécutable
chmod +x deploy-stratum-proxy.sh

# Lancer le script interactif
./deploy-stratum-proxy.sh
```

Le script vous posera des questions :
- Geth RPC URL : `http://localhost:8545` (Enter)
- Stratum listen address : `0.0.0.0:3333` (Enter)
- Initial difficulty : `30000` (notre nouvelle difficulté)
- Pool mode : `n` (pour commencer)
- Verbose logging : `y` (recommandé)
- Install as systemd service : `y` (si vous voulez un service permanent)

### Méthode 3: Créer un script de lancement

```bash
cd ~/go-Ducros

# Créer un script de lancement personnalisé
cat > start-stratum.sh << 'EOF'
#!/bin/bash

cd ~/go-Ducros/stratum-proxy

# Compiler si nécessaire
if [ ! -f stratum-proxy ]; then
    echo "Compilation du stratum-proxy..."
    go build -o stratum-proxy .
fi

# Lancer le proxy
./stratum-proxy \
  --stratum 0.0.0.0:3333 \
  --geth http://localhost:8545 \
  --diff 30000 \
  --algo rx/0 \
  --vardiff-target 30.0 \
  --vardiff-window 10 \
  --max-invalid-streak 10 \
  -v
EOF

chmod +x start-stratum.sh

# Lancer
./start-stratum.sh
```

## 📊 Vérifier que le stratum fonctionne

Une fois lancé, vous devriez voir :

```
🚀 Starting Stratum proxy on 0.0.0.0:3333
🔗 Connected to Geth: http://localhost:8545
⛏️  Algorithm: rx/0
💎 Initial difficulty: 30000
⚙️  VarDiff: target 30.0s, window 10 shares
🛡️  Ban system: max 10 invalid shares
✅ Server started successfully
```

Si vous voyez des erreurs comme :
```
⚠️  Failed to get work: getWork failed: RPC error -32000: no mining work available yet
```

Cela signifie que **geth n'est pas en train de miner**. Retournez aux instructions pour démarrer geth correctement.

## 🔥 Ouvrir le port du firewall

Si le firewall bloque le port 3333 :

```bash
# Avec ufw (Ubuntu)
sudo ufw allow 3333/tcp
sudo ufw status

# Avec firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=3333/tcp
sudo firewall-cmd --reload

# Vérifier que le port est ouvert
sudo netstat -tlnp | grep 3333
```

## 🔌 Connecter xmrig

Une fois le stratum en marche, sur votre PC Windows :

```cmd
xmrig.exe -o 92.222.10.107:3333 -u 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2 -p ducros -a rx/0 --verbose
```

Vous devriez voir sur le **stratum-proxy** :

```
🔌 New connection from 77.192.84.136:57365
✅ Miner logged in: 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2 (XMRig/6.24.0)
📤 Sending job to 77.192.84.136
📩 Share received from 77.192.84.136
✅ Share accepted! difficulty=30000
```

Et sur **xmrig** :

```
[2025-11-15 ...] net      new job from 92.222.10.107:3333 diff 30000
[2025-11-15 ...] cpu      accepted (1/0) diff 30000
[2025-11-15 ...] miner    speed 10s/60s/15m 5000.0 5000.0 n/a H/s max 5500.0 H/s
```

## 🛠️ Dépannage

### Problème: "Failed to get work"

**Solution :** Geth ne mine pas. Lancez geth avec :
```bash
./build/bin/geth \
  --datadir devnet-data \
  --networkid 33669 \
  --http \
  --http.api eth,net,web3,randomx,miner \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --mine \
  --miner.threads 6
```

### Problème: "Connection refused" sur xmrig

**Solutions :**
1. Vérifier que le stratum écoute : `netstat -tlnp | grep 3333`
2. Vérifier le firewall : `sudo ufw status`
3. Vérifier l'IP du VPS : `curl ifconfig.me`

### Problème: "Invalid share" répété

**Solution :** La difficulté est peut-être trop élevée. Réduisez `--diff` :
```bash
./stratum-proxy --diff 10000  # Au lieu de 100000
```

## 📈 Logs utiles

### Voir les logs en temps réel

Si lancé en systemd :
```bash
sudo journalctl -u stratum-proxy -f
```

Si lancé en script :
```bash
# Les logs s'affichent directement dans le terminal
```

### Statistiques du stratum

Le stratum affiche des stats toutes les 30 secondes :
```
📊 Stats: Miners=1/0 Shares=45/0/0 Blocks=3 Hashrate=5000.00 H/s Uptime=5m30s
```

**Explication :**
- `Miners=1/0` : 1 mineur connecté, 0 banni
- `Shares=45/0/0` : 45 valides, 0 invalides, 0 rejetées
- `Blocks=3` : 3 blocs trouvés
- `Hashrate=5000.00 H/s` : Hashrate total du pool
- `Uptime=5m30s` : Temps de fonctionnement

## 🔄 Arrêter/Redémarrer le stratum

### Si lancé manuellement
```bash
# Ctrl+C dans le terminal
# OU
pkill -9 stratum-proxy
```

### Si lancé en systemd
```bash
sudo systemctl stop stratum-proxy
sudo systemctl restart stratum-proxy
sudo systemctl status stratum-proxy
```

## 📝 Configuration avancée

### Mode pool (optionnel)

Si vous voulez faire un pool public :
```bash
./stratum-proxy \
  --stratum 0.0.0.0:3333 \
  --geth http://localhost:8545 \
  --diff 30000 \
  --pool-addr 0xVOTRE_ADRESSE_POOL \
  --pool-fee 1.0 \
  --max-connections 1000 \
  --share-rate-limit 100.0
```

### VarDiff (ajustement automatique de difficulté)

```bash
./stratum-proxy \
  --diff 30000 \
  --vardiff-target 30.0 \   # Cible : 1 share toutes les 30 secondes
  --vardiff-window 10       # Fenêtre de 10 shares pour ajuster
```

Le proxy ajustera automatiquement la difficulté de chaque mineur pour qu'il trouve ~1 share/30s.

## ✅ Checklist finale

- [ ] Geth est en train de miner (`Mining loop started` dans les logs)
- [ ] L'API miner est exposée (`--http.api` contient `miner`)
- [ ] `eth_getWork` retourne du travail (test avec curl)
- [ ] Stratum-proxy est compilé
- [ ] Stratum-proxy est lancé et écoute sur port 3333
- [ ] Port 3333 est ouvert dans le firewall
- [ ] XMRig se connecte avec succès
- [ ] XMRig reçoit des jobs
- [ ] Les shares sont acceptés
- [ ] Les blocs sont trouvés et soumis

## 🎯 Commande complète tout-en-un

Sur le VPS, une seule commande pour tout démarrer :

```bash
cd ~/go-Ducros/stratum-proxy && \
go build -o stratum-proxy . && \
./stratum-proxy --stratum 0.0.0.0:3333 --geth http://localhost:8545 --diff 30000 --algo rx/0 -v
```

Voilà ! Le stratum devrait maintenant distribuer le travail de geth à vos mineurs xmrig. 🚀
