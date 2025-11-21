# Guide de Déploiement Mainnet - Ducros Chain

Guide complet pour déployer ta blockchain Ducros en production.

## 📋 Prérequis

### Serveur Recommandé (Mainnet)
```
CPU: 8+ cores (Ryzen 7 / Intel i7+)
RAM: 32GB minimum, 64GB recommandé
Storage: 2TB SSD NVMe (croissance ~100GB/an)
Network: 1 Gbps, bande passante illimitée
OS: Ubuntu 22.04 LTS ou Debian 12
```

### Providers Recommandés
- **OVH** : Serveurs dédiés France (Rise-1, Rise-2)
- **Scaleway** : Instances Dedibox, GP1-M
- **Hetzner** : Serveurs dédiés Allemagne (AX line)

## 🔧 Étape 1 : Préparation du Serveur

### 1.1 Connexion SSH
```bash
ssh root@votre-serveur-ip
```

### 1.2 Mise à jour système
```bash
apt update && apt upgrade -y
apt install -y build-essential git wget curl vim ufw fail2ban htop
```

### 1.3 Créer utilisateur dédié
```bash
# Créer user ducros
adduser ducros
usermod -aG sudo ducros

# Passer sur ce user
su - ducros
```

### 1.4 Configuration Firewall
```bash
# Autoriser SSH, HTTP, HTTPS, ports Geth
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw allow 30303/tcp   # Geth P2P
sudo ufw allow 30303/udp   # Geth P2P discovery
sudo ufw allow 8545/tcp    # RPC (si exposé)
sudo ufw allow 8546/tcp    # WebSocket (si exposé)

sudo ufw enable
sudo ufw status
```

## 📦 Étape 2 : Installation Go & Dépendances

### 2.1 Installer Go 1.21+
```bash
cd ~
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Ajouter au PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Vérifier
go version  # Doit afficher go1.21.6
```

### 2.2 Installer RandomX
```bash
cd ~
git clone https://github.com/tevador/RandomX.git
cd RandomX
mkdir build && cd build
cmake -DARCH=native ..
make -j$(nproc)
sudo make install
sudo ldconfig
```

## 🚀 Étape 3 : Compiler Ducros Chain

### 3.1 Cloner le repository
```bash
cd ~
git clone https://github.com/Aqui-oi/go-Ducros.git
cd go-Ducros

# Checkout version stable (tag release)
git checkout v1.0.0  # Remplace par ta version
```

### 3.2 Compiler Geth
```bash
# Compiler avec optimisations
make geth

# Vérifier compilation
./build/bin/geth version
```

### 3.3 Installer globalement
```bash
sudo cp build/bin/geth /usr/local/bin/geth-ducros
sudo chmod +x /usr/local/bin/geth-ducros

# Vérifier
geth-ducros version
```

## ⚙️ Étape 4 : Configuration Genesis

### 4.1 Créer répertoire data
```bash
mkdir -p ~/ducros-mainnet/data
mkdir -p ~/ducros-mainnet/keystore
cd ~/ducros-mainnet
```

### 4.2 Créer genesis.json
```bash
nano genesis.json
```

Copie ton fichier genesis (exemple) :
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
    "randomx": {
      "period": 0,
      "epoch": 30000
    }
  },
  "difficulty": "0x4000",
  "gasLimit": "0x1c9c380",
  "alloc": {
    "0xYourPremineAddress1": {"balance": "4000000000000000000000000"},
    "0xYourPremineAddress2": {"balance": "2000000000000000000000000"},
    "0xTreasuryAddress": {"balance": "1000000000000000000000000"}
  }
}
```

**⚠️ IMPORTANT**: Configure correctement :
- `chainId` : 33669 (ou ton ID unique)
- `alloc` : Addresses avec premine ICO
- `difficulty` : Difficulté initiale (ajuste selon hashrate attendu)

### 4.3 Initialiser la blockchain
```bash
geth-ducros --datadir ./data init genesis.json
```

Tu dois voir :
```
INFO [XX-XX|XX:XX:XX.XXX] Successfully wrote genesis state
```

## 🌐 Étape 5 : Configuration Bootnode

### 5.1 Générer clé bootnode
```bash
# Installer bootnode tool
cd ~/go-Ducros
make all

# Générer clé
./build/bin/bootnode -genkey boot.key

# Obtenir enode
./build/bin/bootnode -nodekey boot.key -writeaddress
# Note l'adresse (ex: abc123def456...)
```

### 5.2 Démarrer bootnode (terminal séparé ou systemd)
```bash
# En foreground (test)
./build/bin/bootnode -nodekey boot.key -addr :30310

# Ou créer service systemd (voir plus bas)
```

L'enode complet sera :
```
enode://[ADDRESS]@votre-ip:30310
```

### 5.3 Mettre à jour bootnodes_ducros.go
Édite `params/bootnodes_ducros.go` avec tes vrais bootnodes :
```go
var DucrosMainnetBootnodes = []string{
    "enode://[ADDRESS1]@ip1:30310",
    "enode://[ADDRESS2]@ip2:30310",
    "enode://[ADDRESS3]@ip3:30310",
}
```

Recompile si modifié :
```bash
make geth
sudo cp build/bin/geth /usr/local/bin/geth-ducros
```

## 🚀 Étape 6 : Lancer le Node Mainnet

### 6.1 Créer compte coinbase (pour mining)
```bash
geth-ducros --datadir ./data account new

# Note l'adresse créée et le mot de passe !
# Exemple: 0xabcd1234...
```

### 6.2 Script de lancement

Crée `start-mainnet.sh` :
```bash
nano start-mainnet.sh
```

Contenu :
```bash
#!/bin/bash

geth-ducros \
  --datadir ./data \
  --networkid 33669 \
  --port 30303 \
  --http \
  --http.addr "127.0.0.1" \
  --http.port 8545 \
  --http.api "eth,net,web3,personal,admin,miner" \
  --http.corsdomain "*" \
  --ws \
  --ws.addr "127.0.0.1" \
  --ws.port 8546 \
  --ws.api "eth,net,web3" \
  --mine \
  --miner.threads 4 \
  --miner.etherbase "0xVotreAdresseCoinbase" \
  --syncmode "full" \
  --gcmode "archive" \
  --maxpeers 100 \
  --nat "extip:votre-ip-publique" \
  --bootnodes "enode://..." \
  --verbosity 3 \
  --log.file ./ducros.log \
  --metrics \
  --pprof \
  --pprof.addr "127.0.0.1" \
  --pprof.port 6060
```

Rendre exécutable :
```bash
chmod +x start-mainnet.sh
```

### 6.3 Lancer (test foreground)
```bash
./start-mainnet.sh
```

Vérifications :
- ✅ Node démarre sans erreur
- ✅ Connexion aux bootnodes
- ✅ Synchronisation commence
- ✅ Mining démarre (si activé)

## 🔐 Étape 7 : Sécuriser le Déploiement

### 7.1 Ne PAS exposer RPC publiquement
```bash
# RPC DOIT rester sur 127.0.0.1
# Si besoin accès externe, utilise Nginx reverse proxy avec SSL
```

### 7.2 Configurer Fail2Ban
```bash
sudo nano /etc/fail2ban/jail.local
```

Ajoute :
```ini
[sshd]
enabled = true
port = 22
maxretry = 3
bantime = 3600
```

```bash
sudo systemctl restart fail2ban
```

### 7.3 Désactiver mot de passe SSH (clé uniquement)
```bash
sudo nano /etc/ssh/sshd_config
```

Modifier :
```
PasswordAuthentication no
PubkeyAuthentication yes
PermitRootLogin no
```

```bash
sudo systemctl restart sshd
```

### 7.4 Backup clés privées
```bash
# Backup keystore (CRUCIAL!)
tar -czf keystore-backup-$(date +%Y%m%d).tar.gz ./data/keystore

# Télécharge localement
scp ducros@votre-ip:~/ducros-mainnet/keystore-backup-*.tar.gz .

# Stocke dans coffre-fort sécurisé (1Password, Bitwarden, etc.)
```

## 🤖 Étape 8 : Systemd Service (Auto-restart)

### 8.1 Créer service systemd
```bash
sudo nano /etc/systemd/system/ducros-mainnet.service
```

Contenu :
```ini
[Unit]
Description=Ducros Chain Mainnet Node
After=network.target

[Service]
Type=simple
User=ducros
WorkingDirectory=/home/ducros/ducros-mainnet
ExecStart=/usr/local/bin/geth-ducros \
  --datadir /home/ducros/ducros-mainnet/data \
  --networkid 33669 \
  --port 30303 \
  --http \
  --http.addr "127.0.0.1" \
  --http.port 8545 \
  --http.api "eth,net,web3,personal,admin,miner" \
  --ws \
  --ws.addr "127.0.0.1" \
  --ws.port 8546 \
  --mine \
  --miner.threads 4 \
  --miner.etherbase "0xVotreAdresse" \
  --syncmode "full" \
  --maxpeers 100 \
  --nat "extip:votre-ip" \
  --bootnodes "enode://..." \
  --verbosity 3 \
  --metrics

Restart=always
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ducros-mainnet

[Install]
WantedBy=multi-user.target
```

### 8.2 Activer et démarrer
```bash
sudo systemctl daemon-reload
sudo systemctl enable ducros-mainnet
sudo systemctl start ducros-mainnet

# Vérifier status
sudo systemctl status ducros-mainnet

# Voir logs
sudo journalctl -u ducros-mainnet -f
```

## 📊 Étape 9 : Monitoring & Maintenance

### 9.1 Vérifier santé du node
```bash
# Se connecter au console
geth-ducros attach http://127.0.0.1:8545

# Dans le console:
> eth.blockNumber        # Numéro bloc actuel
> net.peerCount          # Nombre de peers
> miner.mining           # Mining actif ?
> eth.hashrate           # Hashrate local
> eth.syncing            # Synchro en cours ?
> admin.peers            # Liste peers connectés
```

### 9.2 Monitoring avec Prometheus + Grafana (optionnel)

**Prometheus** :
```bash
# Installer Prometheus
sudo apt install prometheus -y

# Config prometheus.yml
sudo nano /etc/prometheus/prometheus.yml
```

Ajoute target :
```yaml
scrape_configs:
  - job_name: 'ducros-mainnet'
    static_configs:
      - targets: ['localhost:6060']
```

**Grafana** :
```bash
sudo apt install grafana -y
sudo systemctl enable grafana-server
sudo systemctl start grafana-server

# Accès: http://votre-ip:3000 (admin/admin)
```

### 9.3 Backup automatique
```bash
# Créer script backup
nano ~/backup-ducros.sh
```

Contenu :
```bash
#!/bin/bash
BACKUP_DIR="/home/ducros/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup keystore
tar -czf $BACKUP_DIR/keystore-$DATE.tar.gz \
  /home/ducros/ducros-mainnet/data/keystore

# Backup chaindata (si petit)
# tar -czf $BACKUP_DIR/chaindata-$DATE.tar.gz \
#   /home/ducros/ducros-mainnet/data/geth/chaindata

# Garder seulement 7 derniers backups
find $BACKUP_DIR -name "keystore-*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

Cron quotidien :
```bash
chmod +x ~/backup-ducros.sh
crontab -e
```

Ajoute :
```
0 3 * * * /home/ducros/backup-ducros.sh >> /home/ducros/backup.log 2>&1
```

### 9.4 Monitoring disk space
```bash
# Vérifier espace disque
df -h

# Nettoyer anciens logs si besoin
sudo journalctl --vacuum-time=7d
```

## 🌍 Étape 10 : Exposer RPC via Nginx (optionnel)

Si tu veux exposer RPC publiquement avec SSL :

### 10.1 Installer Nginx + Certbot
```bash
sudo apt install nginx certbot python3-certbot-nginx -y
```

### 10.2 Config Nginx
```bash
sudo nano /etc/nginx/sites-available/ducros-rpc
```

Contenu :
```nginx
server {
    listen 80;
    server_name rpc.ducroschain.io;

    location / {
        proxy_pass http://127.0.0.1:8545;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 10.3 Activer et SSL
```bash
sudo ln -s /etc/nginx/sites-available/ducros-rpc /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Obtenir certificat SSL
sudo certbot --nginx -d rpc.ducroschain.io
```

## 📋 Checklist Finale Mainnet

Avant de lancer publiquement :

- [ ] ✅ Genesis configuré avec bon chainId
- [ ] ✅ Premine alloué aux bonnes addresses ICO
- [ ] ✅ Au moins 3 bootnodes déployés sur IPs différentes
- [ ] ✅ Firewall configuré (UFW)
- [ ] ✅ Systemd service fonctionnel
- [ ] ✅ Backup keystore sécurisé OFFLINE
- [ ] ✅ Monitoring configuré (logs, metrics)
- [ ] ✅ Test mining fonctionnel
- [ ] ✅ RPC sécurisé (pas public ou SSL)
- [ ] ✅ Documentation déploiement communauté
- [ ] ✅ Block explorer déployé
- [ ] ✅ Smart contracts Treasury déployés
- [ ] ✅ Audit sécurité externe réalisé

## 🚨 Commandes Utiles

### Redémarrer node
```bash
sudo systemctl restart ducros-mainnet
```

### Voir logs en temps réel
```bash
sudo journalctl -u ducros-mainnet -f
```

### Nettoyer chaindata (DANGEREUX - perte sync)
```bash
sudo systemctl stop ducros-mainnet
rm -rf ~/ducros-mainnet/data/geth/chaindata
geth-ducros --datadir ./data init genesis.json
sudo systemctl start ducros-mainnet
```

### Ajouter peer manuellement
```bash
geth-ducros attach http://127.0.0.1:8545
> admin.addPeer("enode://...")
```

### Importer compte
```bash
geth-ducros --datadir ./data account import /path/to/private.key
```

## 📞 Support

- **GitHub Issues**: https://github.com/Aqui-oi/go-Ducros/issues
- **Discord**: [Créer serveur communauté]
- **Email**: support@ducroschain.io

---

**Créé par**: Alexandre Ducros (Aquí oï SASU)
**Version**: 1.0.0
**Date**: Novembre 2025
