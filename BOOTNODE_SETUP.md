# 🌐 Ducros Bootnode Setup Guide

## Qu'est-ce qu'un Bootnode ?

Les **bootnodes** sont les **premiers points d'entrée** de votre réseau P2P. Sans eux, les nouveaux nœuds ne peuvent pas découvrir le réseau Ducros.

### Rôle des Bootnodes

```
Nouveau Nœud → Bootnode → Liste des Peers → Connexion au Réseau
```

**Important** : Les bootnodes doivent être **stables** et **toujours en ligne** !

---

## 🎯 Nombre de Bootnodes Recommandé

| Réseau | Minimum | Recommandé | Optimal |
|--------|---------|------------|---------|
| **Testnet** | 1 | 2-3 | 3-5 |
| **Mainnet** | 3 | 5 | 7-10 |

**Pourquoi plusieurs ?**
- Redondance (si un tombe, les autres fonctionnent)
- Distribution géographique
- Charge distribuée

---

## 🔧 Étape 1 : Générer les Clés

### Sur votre serveur local

```bash
# Créer un dossier pour les clés
mkdir -p ~/ducros-bootnodes
cd ~/ducros-bootnodes

# Générer 5 clés de bootnodes
for i in {1..5}; do
  ../build/bin/bootnode -genkey=bootnode${i}.key
  echo "✅ Bootnode $i key generated"
done
```

### Obtenir les Enodes

```bash
# Pour chaque bootnode
for i in {1..5}; do
  echo "Bootnode $i:"
  ../build/bin/bootnode -nodekey=bootnode${i}.key -writeaddress
  echo ""
done
```

**Exemple de sortie :**
```
Bootnode 1:
enode://a1b2c3d4e5f6...@127.0.0.1:0

Bootnode 2:
enode://f6e5d4c3b2a1...@127.0.0.1:0
```

**Notez les clés publiques** (la partie hexadécimale après `enode://`)

---

## 🌍 Étape 2 : Déployer les Bootnodes

### Option A : VPS Cloud (Recommandé)

Louez 3-5 VPS dans différentes régions :

| Bootnode | Région | Provider | IP Example |
|----------|--------|----------|------------|
| Bootnode 1 | Europe | Hetzner | `95.217.xxx.xxx` |
| Bootnode 2 | US East | DigitalOcean | `164.92.xxx.xxx` |
| Bootnode 3 | Asia | Vultr | `45.76.xxx.xxx` |
| Bootnode 4 | Europe | OVH | `51.210.xxx.xxx` |
| Bootnode 5 | US West | Linode | `172.105.xxx.xxx` |

**Specs minimales par bootnode :**
- CPU: 1 core
- RAM: 512 MB
- Storage: 10 GB
- Réseau: 1 Gbps
- **Coût**: ~$3-5/mois par bootnode

### Option B : Serveurs Personnels

Si vous avez des serveurs dédiés, utilisez-les pour les bootnodes.

---

## 📦 Étape 3 : Installation sur les Serveurs

### Sur chaque serveur bootnode

```bash
# 1. Installer les dépendances
sudo apt update
sudo apt install -y wget

# 2. Télécharger geth
wget https://github.com/Aqui-oi/go-Ducros/releases/download/vX.X.X/geth-linux-amd64
chmod +x geth-linux-amd64
sudo mv geth-linux-amd64 /usr/local/bin/geth

# 3. Copier la clé du bootnode
mkdir -p /var/lib/ducros-bootnode
# Uploadez bootnode1.key, bootnode2.key, etc. sur chaque serveur
scp bootnode1.key user@server1:/var/lib/ducros-bootnode/bootnode.key

# 4. Créer le service systemd
sudo nano /etc/systemd/system/ducros-bootnode.service
```

### Fichier de service systemd

```ini
[Unit]
Description=Ducros Bootnode
After=network.target

[Service]
Type=simple
User=ducros
WorkingDirectory=/var/lib/ducros-bootnode
ExecStart=/usr/local/bin/geth bootnode \
  -nodekey=/var/lib/ducros-bootnode/bootnode.key \
  -addr=0.0.0.0:30303 \
  -verbosity=3
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Créer l'utilisateur et démarrer

```bash
# Créer utilisateur ducros
sudo useradd -r -s /bin/false ducros
sudo chown -R ducros:ducros /var/lib/ducros-bootnode

# Activer et démarrer le service
sudo systemctl daemon-reload
sudo systemctl enable ducros-bootnode
sudo systemctl start ducros-bootnode

# Vérifier le statut
sudo systemctl status ducros-bootnode
```

---

## 🔍 Étape 4 : Vérifier les Bootnodes

### Obtenir l'enode de chaque serveur

```bash
# Sur le serveur bootnode
geth bootnode -nodekey=/var/lib/ducros-bootnode/bootnode.key -writeaddress
```

Notez l'output, puis construisez l'enode complet :

```
enode://[public_key]@[IP_publique_du_serveur]:30303
```

**Exemple :**
```
enode://a1b2c3d4e5f6789...@95.217.123.45:30303
enode://f6e5d4c3b2a1098...@164.92.234.56:30303
enode://1234567890abcdef...@45.76.111.222:30303
```

### Tester la connectivité

```bash
# Depuis un autre serveur/machine
nc -zv 95.217.123.45 30303
# Should output: Connection to 95.217.123.45 30303 port [tcp/*] succeeded!
```

---

## 📝 Étape 5 : Mettre à Jour le Code

### Modifier `params/bootnodes_ducros.go`

```go
var DucrosBootnodes = []string{
	"enode://a1b2c3d4e5f6789abc...@95.217.123.45:30303",
	"enode://f6e5d4c3b2a1098def...@164.92.234.56:30303",
	"enode://1234567890abcdef12...@45.76.111.222:30303",
	"enode://9876543210fedcba98...@51.210.99.88:30303",
	"enode://abcdef0123456789ab...@172.105.77.66:30303",
}
```

### Recompiler et Redistribuer

```bash
# Recompiler geth avec les nouveaux bootnodes
make geth

# Créer une release
git add params/bootnodes_ducros.go
git commit -m "feat: Add production bootnodes for mainnet"
git tag v1.0.0
git push origin v1.0.0

# Distribuer le binaire compilé
```

---

## 🎯 Étape 6 : DNS (Optionnel mais Recommandé)

Au lieu d'utiliser des IPs directement, utilisez des noms de domaine :

```
bootnode1.ducros.network → 95.217.123.45
bootnode2.ducros.network → 164.92.234.56
bootnode3.ducros.network → 45.76.111.222
```

### Avantages

- ✅ Peut changer l'IP sans recompiler
- ✅ Plus professionnel
- ✅ Meilleure gestion DNS

### Configuration DNS

Dans votre registrar (ex: Cloudflare, Namecheap) :

```
Type: A
Name: bootnode1.ducros.network
Value: 95.217.123.45
TTL: 3600

Type: A
Name: bootnode2.ducros.network
Value: 164.92.234.56
TTL: 3600
```

### Code avec DNS

```go
var DucrosBootnodes = []string{
	"enode://a1b2c3...@bootnode1.ducros.network:30303",
	"enode://f6e5d4...@bootnode2.ducros.network:30303",
	"enode://123456...@bootnode3.ducros.network:30303",
}
```

---

## 🛡️ Sécurité des Bootnodes

### Firewall Configuration

```bash
# Autoriser uniquement le port 30303
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp
sudo ufw enable
```

### Monitoring

Installez un monitoring pour surveiller les bootnodes :

```bash
# Installer prometheus node exporter
wget https://github.com/prometheus/node_exporter/releases/download/v1.5.0/node_exporter-1.5.0.linux-amd64.tar.gz
tar -xzf node_exporter-1.5.0.linux-amd64.tar.gz
sudo mv node_exporter-1.5.0.linux-amd64/node_exporter /usr/local/bin/

# Créer service
sudo nano /etc/systemd/system/node_exporter.service
```

```ini
[Unit]
Description=Node Exporter

[Service]
User=ducros
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable node_exporter
sudo systemctl start node_exporter
```

---

## 📊 Vérification du Réseau

### Tester qu'un nœud peut se connecter

Sur une machine de test :

```bash
# Lancer geth sans bootnodes custom
./geth --datadir /tmp/test-node \
  --networkid 33669 \
  --nodiscover \
  --bootnodes "enode://a1b2c3...@bootnode1.ducros.network:30303"

# Dans les logs, vous devriez voir :
# INFO [XX-XX|XX:XX:XX] Successfully connected to bootnode
# INFO [XX-XX|XX:XX:XX] Peer connected                 id=a1b2c3...
```

---

## 🚨 Problèmes Courants

### Bootnode ne démarre pas

```bash
# Vérifier les logs
sudo journalctl -u ducros-bootnode -f

# Vérifier que le port est ouvert
sudo netstat -tulpn | grep 30303

# Vérifier les permissions
ls -la /var/lib/ducros-bootnode/
```

### Nœuds ne peuvent pas se connecter

```bash
# Tester la connectivité
telnet bootnode1.ducros.network 30303

# Vérifier le firewall
sudo ufw status

# Vérifier que le bootnode écoute
sudo ss -tulpn | grep 30303
```

### DNS ne résout pas

```bash
# Tester la résolution DNS
nslookup bootnode1.ducros.network
dig bootnode1.ducros.network

# Vérifier la propagation DNS
# https://www.whatsmydns.net/
```

---

## 📋 Checklist Avant Mainnet

Avant de lancer le mainnet, vérifiez :

- [ ] Au moins 3 bootnodes déployés
- [ ] Bootnodes dans différentes régions géographiques
- [ ] Tous les bootnodes accessibles publiquement
- [ ] Port 30303 ouvert (TCP et UDP)
- [ ] Services systemd configurés et activés
- [ ] Monitoring en place
- [ ] DNS configuré (si applicable)
- [ ] Enodes ajoutés dans `params/bootnodes_ducros.go`
- [ ] Code recompilé avec les nouveaux bootnodes
- [ ] Binaires distribués aux utilisateurs
- [ ] Documentation publiée

---

## 💡 Conseils Pro

1. **Redondance** : Ayez toujours 2× plus de bootnodes que nécessaire
2. **Geographic Distribution** : Répartissez dans au moins 3 continents
3. **Uptime** : Visez 99.9% d'uptime pour les bootnodes
4. **Backups** : Sauvegardez les clés des bootnodes
5. **Rotation** : Prévoyez de pouvoir remplacer un bootnode sans downtime

---

## 🆘 Support

Si vous avez des problèmes avec les bootnodes :

1. Vérifiez les logs : `journalctl -u ducros-bootnode`
2. Testez la connectivité réseau
3. Vérifiez le firewall
4. Consultez la documentation geth : https://geth.ethereum.org/docs/fundamentals/peer-to-peer

---

**Prêt à lancer vos bootnodes et votre mainnet !** 🚀
