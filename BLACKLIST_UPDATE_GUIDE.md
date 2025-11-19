# Guide : Ajouter des Adresses Botnet à la Blacklist Après Déploiement

## 🎯 Vue d'Ensemble

Ajouter des adresses à la blacklist mining est une **modification consensus-breaking**. Cela signifie que tous les nœuds du réseau doivent mettre à jour leur code en même temps, sinon le réseau se divisera en deux chaînes.

---

## 📋 Processus Complet (5 Étapes)

### Étape 1 : Identification des Adresses Malveillantes

**Critères pour blacklister une adresse** :
- ✅ Preuve de mining par botnet (analyse de patterns)
- ✅ Adresse liée à malware connu
- ✅ Mining depuis des machines compromises
- ✅ Rapports de la communauté avec preuves
- ✅ Analyse forensique de trafic réseau

**Sources d'identification** :
```bash
# 1. Monitoring du réseau
./geth attach --exec "admin.peers" | grep "suspicious_pattern"

# 2. Analyse des patterns de mining
# - Beaucoup de petits miners (1-2 H/s chacun)
# - IPs résidentielles variées
# - Connexions courtes et fréquentes
# - Pas de communication normale (RPC, etc.)

# 3. Rapports de la communauté
# - Discord/Telegram/Forum
# - Utilisateurs rapportant mining non autorisé
# - Logs d'antivirus détectant le miner
```

**Exemple de collecte de données** :
```bash
# Script pour détecter les miners suspects
cat > detect_botnet.sh <<'EOF'
#!/bin/bash

# Récupère tous les miners actifs
miners=$(./geth attach --exec "eth.getBlock('latest').miner")

# Analyse le pattern de chaque mineur
for miner in $miners; do
    # Nombre de blocs minés
    blocks=$(./geth attach --exec "eth.getBlockNumber() - eth.getBlock('earliest', miner).number")

    # Hashrate estimé
    hashrate=$(calculate_hashrate $miner)

    # Si petit hashrate mais beaucoup d'IPs différentes = suspect
    if [ $hashrate -lt 100 ] && [ $unique_ips -gt 50 ]; then
        echo "⚠️ Suspect: $miner (hashrate: $hashrate, IPs: $unique_ips)"
    fi
done
EOF
```

---

### Étape 2 : Validation Communautaire

**Avant d'ajouter une adresse**, consulte la communauté :

1. **Publier un rapport** (GitHub Issue / Forum) :
   ```markdown
   # Proposition de Blacklist : 0xABC...DEF

   ## Preuves
   - Pattern de mining suspect (1000+ IPs différentes, 1-2 H/s chacune)
   - Rapports de 50+ utilisateurs (mining non autorisé)
   - Corrélation avec malware "DucrosMiner.exe" détecté par antivirus

   ## Impact
   - ~5% du hashrate total
   - Coût pour le réseau : 42 DCR/semaine détournés

   ## Proposition
   Ajouter 0xABC...DEF à la mining blacklist

   ## Vote
   👍 Pour blacklister
   👎 Contre blacklister
   💬 Besoin de plus d'infos
   ```

2. **Période de discussion** : 7-14 jours minimum

3. **Vote de la communauté** (si gouvernance en place)

---

### Étape 3 : Modification du Code

Une fois validé, modifie le code source :

#### A. Ouvrir `params/protocol_params.go`

```go
// Ligne ~230
var MiningBlacklist = map[common.Address]bool{
    // Exemple blacklisted addresses - replace with actual malicious miners
    // common.HexToAddress("0xMALICIOUS_MINER_ADDRESS_1"): true, // Reason: Confirmed botnet
}
```

#### B. Ajouter la nouvelle adresse avec DOCUMENTATION

```go
var MiningBlacklist = map[common.Address]bool{
    // ═══════════════════════════════════════════════════════════════
    // MINING BLACKLIST - Updated: 2025-11-20
    // ═══════════════════════════════════════════════════════════════
    // Each entry must be documented with:
    // - Date added
    // - Reason for blacklisting
    // - Evidence/report link
    // - Estimated impact
    // ═══════════════════════════════════════════════════════════════

    // Added: 2025-11-20
    // Reason: Confirmed botnet operation (1000+ infected machines)
    // Evidence: https://github.com/yourproject/issues/123
    // Impact: ~5% hashrate (~42 DCR/week)
    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"): true,

    // Added: 2025-11-25
    // Reason: Malware "DucrosMiner.exe" using stolen compute
    // Evidence: https://github.com/yourproject/issues/456
    // Impact: ~2% hashrate (~16 DCR/week)
    common.HexToAddress("0x1234567890123456789012345678901234567890"): true,

    // Note: Blacklist only affects MINING rewards, NOT regular transactions
    // Blacklisted miners receive 0%, 100% goes to treasury
}
```

#### C. Sauvegarder et Formater

```bash
# Formate le code
gofmt -w params/protocol_params.go

# Vérifie la compilation
go build ./params
```

---

### Étape 4 : Coordination du Hard Fork

**CRITIQUE** : Tous les nœuds doivent update en même temps !

#### A. Choisir un Numéro de Bloc d'Activation

```go
// Dans params/config.go, ajoute une constante
const BlacklistUpdate1Block = 500000  // Bloc où la blacklist prend effet

// Puis dans consensus/randomx/consensus.go
func accumulateRewards(config *params.ChainConfig, stateDB vm.StateDB, header *types.Header, uncles []*types.Header) {
    // ... code existant ...

    // Check blacklist only AFTER activation block
    var isBlacklisted bool
    if header.Number.Uint64() >= params.BlacklistUpdate1Block {
        isBlacklisted = params.IsMinerBlacklisted(header.Coinbase)
    } else {
        isBlacklisted = false  // Blacklist not active yet
    }

    // ... reste du code ...
}
```

**Pourquoi un bloc d'activation ?**
- Donne du temps aux nœuds pour update
- Évite les splits de chaîne
- Permet une transition douce

**Comment choisir le bloc** :
```bash
# Bloc actuel
current_block=$(./geth attach --exec "eth.blockNumber")
echo "Bloc actuel: $current_block"

# Ajoute ~7 jours (13s/bloc * 6646 blocs/jour * 7 jours)
activation_block=$((current_block + 46522))
echo "Bloc d'activation: $activation_block"
```

#### B. Annoncer le Hard Fork

**Au moins 2 semaines à l'avance** :

1. **GitHub Release** :
   ```markdown
   # Ducros v1.2.0 - Blacklist Update #1

   ## ⚠️ MANDATORY UPDATE - HARD FORK

   ### Activation
   - Block: #500,000 (estimated: 2025-12-01 00:00 UTC)
   - All nodes MUST update before this block

   ### Changes
   - Added 2 addresses to mining blacklist (botnet operations)
   - See BLACKLIST_UPDATE_GUIDE.md for details

   ### Action Required
   All node operators must:
   1. Download v1.2.0
   2. Stop your node: `./geth stop`
   3. Replace binary: `cp geth-v1.2.0 ./geth`
   4. Restart: `./geth`

   ### Addresses Blacklisted
   - 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb (Botnet)
   - 0x1234567890123456789012345678901234567890 (Malware)

   ### Timeline
   - 2025-11-20: Announcement
   - 2025-11-27: Final reminder
   - 2025-12-01: Activation (block 500,000)
   ```

2. **Communication multi-canal** :
   - Discord/Telegram announcement
   - Twitter/X post
   - Email aux opérateurs de nœuds connus
   - Forum post

3. **Monitoring** :
   ```bash
   # Script pour vérifier combien de nœuds ont update
   cat > check_update_status.sh <<'EOF'
   #!/bin/bash

   total_peers=$(./geth attach --exec "admin.peers.length")
   updated_peers=0

   for peer in $(./geth attach --exec "admin.peers"); do
       version=$(echo $peer | jq -r '.name')
       if [[ $version == *"v1.2.0"* ]]; then
           ((updated_peers++))
       fi
   done

   percentage=$((updated_peers * 100 / total_peers))
   echo "Peers updated: $updated_peers/$total_peers ($percentage%)"

   if [ $percentage -lt 80 ]; then
       echo "⚠️ WARNING: Less than 80% updated!"
   fi
   EOF

   chmod +x check_update_status.sh
   ./check_update_status.sh
   ```

#### C. Créer la Release

```bash
# Tag git
git tag -a v1.2.0-blacklist-update-1 -m "Blacklist Update #1: Add 2 botnet addresses"

# Push tag
git push origin v1.2.0-blacklist-update-1

# Compile pour différentes plateformes
make all
# Crée: geth-linux-amd64, geth-windows-amd64, geth-darwin-amd64

# Upload sur GitHub Releases avec checksums
sha256sum geth-* > checksums.txt
```

---

### Étape 5 : Déploiement et Surveillance

#### A. Jour du Hard Fork

**Avant l'activation** :
```bash
# Vérifier que la majorité a update
current=$(./geth attach --exec "eth.blockNumber")
activation=500000
blocks_remaining=$((activation - current))

echo "Blocks until activation: $blocks_remaining"
echo "Estimated time: $((blocks_remaining * 13 / 3600)) hours"

# Vérifier % de nœuds updated
./check_update_status.sh
```

**Pendant l'activation** (bloc 500,000) :
```bash
# Monitoring en temps réel
./geth attach --exec "
    eth.subscribe('newHeads', function(error, blockHeader) {
        if (blockHeader.number >= 500000) {
            console.log('🎉 Hard fork activated at block:', blockHeader.number);

            // Vérifie que les adresses blacklistées ne reçoivent plus de rewards
            var block = eth.getBlock(blockHeader.number, true);
            console.log('Miner:', block.miner);
            console.log('Is blacklisted:',
                block.miner == '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' ||
                block.miner == '0x1234567890123456789012345678901234567890'
            );
        }
    });
"
```

#### B. Post-Activation

**Surveillance pendant 48h** :
```bash
# Script de monitoring post-fork
cat > monitor_post_fork.sh <<'EOF'
#!/bin/bash

echo "Monitoring post-fork..."

# Vérifie qu'il n'y a pas de split de chaîne
for i in {1..48}; do
    # Vérifie le consensus
    latest=$(./geth attach --exec "eth.getBlock('latest').hash")
    peer_count=$(./geth attach --exec "admin.peers.length")

    echo "[Hour $i] Latest block hash: $latest"
    echo "[Hour $i] Peer count: $peer_count"

    # Si le nombre de peers chute brutalement = possible split
    if [ $peer_count -lt 10 ]; then
        echo "🚨 ALERT: Peer count dropped to $peer_count!"
        # Envoie notification
    fi

    # Vérifie que les adresses blacklistées ne reçoivent pas de rewards
    recent_blocks=$(./geth attach --exec "
        var blacklisted = ['0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb', '0x1234567890123456789012345678901234567890'];
        var violations = 0;
        for (var i = 0; i < 100; i++) {
            var block = eth.getBlock(eth.blockNumber - i);
            if (blacklisted.includes(block.miner)) {
                console.log('⚠️ Blacklisted miner still mining:', block.miner, 'at block', block.number);
                violations++;
            }
        }
        violations;
    ")

    if [ "$recent_blocks" != "0" ]; then
        echo "🚨 ALERT: Blacklisted addresses still mining!"
    fi

    sleep 3600  # 1 heure
done
EOF

chmod +x monitor_post_fork.sh
./monitor_post_fork.sh &
```

**Vérifier l'impact** :
```bash
# Analyse de l'effet de la blacklist
./geth attach --exec "
    var blacklisted = ['0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb'];
    var beforeFork = 499900;
    var afterFork = 500100;

    // Compte les blocs minés AVANT
    var blocksBefore = 0;
    for (var i = beforeFork; i < 500000; i++) {
        if (eth.getBlock(i).miner == blacklisted[0]) {
            blocksBefore++;
        }
    }

    // Compte les blocs minés APRÈS
    var blocksAfter = 0;
    for (var i = 500000; i < afterFork; i++) {
        if (eth.getBlock(i).miner == blacklisted[0]) {
            blocksAfter++;
        }
    }

    console.log('Blocks mined BEFORE fork:', blocksBefore);
    console.log('Blocks mined AFTER fork:', blocksAfter);
    console.log('Success:', blocksAfter == 0 ? '✅' : '❌');
"
```

---

## 🔄 Mises à Jour Régulières

### Stratégie Long Terme

Pour éviter trop de hard forks, envisage des **mises à jour programmées** :

```go
// Dans params/config.go
const (
    BlacklistUpdate1Block = 500000   // 2025-12-01
    BlacklistUpdate2Block = 1000000  // 2026-06-01
    BlacklistUpdate3Block = 1500000  // 2027-01-01
    // etc.
)

// Permet de planifier les blacklist updates tous les 6 mois
// La communauté sait que ces dates sont les "maintenance windows"
```

**Avantages** :
- ✅ La communauté sait à l'avance quand update
- ✅ Moins de coordination nécessaire
- ✅ On peut grouper plusieurs adresses par update
- ✅ Processus plus fluide

---

## 📊 Checklist de Déploiement

Utilise cette checklist pour chaque ajout à la blacklist :

### Phase 1 : Préparation (Semaine -4 à -2)
- [ ] Identifier l'adresse malveillante
- [ ] Collecter les preuves
- [ ] Créer un rapport public (GitHub Issue)
- [ ] Ouvrir discussion communautaire
- [ ] Voter (si gouvernance en place)

### Phase 2 : Développement (Semaine -2)
- [ ] Modifier `params/protocol_params.go`
- [ ] Ajouter documentation dans le code
- [ ] Implémenter bloc d'activation
- [ ] Tester sur testnet local
- [ ] Compiler pour toutes les plateformes
- [ ] Générer checksums

### Phase 3 : Annonce (Semaine -2 à -1)
- [ ] Créer GitHub Release (draft)
- [ ] Rédiger annonce détaillée
- [ ] Poster sur tous les canaux (Discord, Twitter, Forum)
- [ ] Envoyer emails aux opérateurs de nœuds
- [ ] Mettre à jour documentation

### Phase 4 : Surveillance Pré-Activation (Semaine -1)
- [ ] Monitorer % de nœuds updated
- [ ] Rappels réguliers (J-7, J-3, J-1)
- [ ] Vérifier que >80% ont update
- [ ] Préparer monitoring temps réel

### Phase 5 : Activation (Jour J)
- [ ] Monitoring en temps réel
- [ ] Vérifier que le fork s'active correctement
- [ ] Vérifier qu'il n'y a pas de split
- [ ] Confirmer que la blacklist fonctionne
- [ ] Poster mise à jour "Success" sur les canaux

### Phase 6 : Post-Activation (J+1 à J+7)
- [ ] Monitoring continu 48h
- [ ] Vérifier impact sur hashrate
- [ ] Vérifier que blacklistés ne minent plus
- [ ] Analyser feedback de la communauté
- [ ] Documenter leçons apprises

---

## ⚠️ Gestion des Urgences

### Si Split de Chaîne Détecté

```bash
# 1. Identifier le split
./geth attach --exec "
    var myHash = eth.getBlock('latest').hash;
    var peerHashes = admin.peers.map(p => p.latestBlock);
    console.log('My hash:', myHash);
    console.log('Peer hashes:', peerHashes);
    // Si différent = split!
"

# 2. Communication d'urgence
echo "🚨 CHAIN SPLIT DETECTED - ALL NODES STOP MINING" | mail -s "URGENT" ops@ducros.network

# 3. Coordonner résolution
# Option A: Rollback à avant le fork
# Option B: Push emergency fix
# Option C: Accepter une des deux chaînes
```

### Si Adresse Innocente Blacklistée par Erreur

1. **Arrêter immédiatement le déploiement**
2. **Annonce publique avec excuses**
3. **Nouveau hard fork pour retirer l'adresse**
4. **Compensation potentielle (airdrop, etc.)**

---

## 🔍 Exemple Complet (Timeline)

```
2025-11-20: Détection botnet (0x742d...)
├─ Collecte preuves pendant 3 jours
└─ 500 IPs différentes, 2 H/s chacune

2025-11-23: Publication rapport GitHub Issue #123
├─ Preuves documentées
└─ Vote communautaire ouvert

2025-11-30: Vote fermé (85% pour blacklist)
└─ Décision: Blacklister l'adresse

2025-12-01: Modification du code
├─ Ajout à MiningBlacklist
├─ Bloc activation: 500,000 (estimé 2025-12-20)
└─ Commit + tag v1.2.0

2025-12-02: Annonce hard fork
├─ GitHub Release published
├─ Discord/Twitter announcements
├─ Email à 50 node operators
└─ Documentation updated

2025-12-09: Reminder 1 (J-11)
└─ 45% des nœuds ont update

2025-12-16: Reminder 2 (J-4)
└─ 78% des nœuds ont update

2025-12-19: Final reminder (J-1)
└─ 92% des nœuds ont update ✅

2025-12-20: ACTIVATION (Bloc 500,000)
├─ 00:00 UTC: Fork activated
├─ Monitoring en temps réel
├─ Blacklist confirmed working ✅
└─ No chain split detected ✅

2025-12-21: Post-mortem
├─ Hashrate stable
├─ Adresse 0x742d... ne mine plus
├─ Treasury reçoit +5 DCR/jour extra
└─ Success! 🎉
```

---

## 📚 Ressources

### Scripts Utiles

Tous dans le dossier `scripts/blacklist/` :
- `detect_botnet.sh` - Détection automatique de patterns suspects
- `check_update_status.sh` - Vérifie % de nœuds updated
- `monitor_post_fork.sh` - Surveillance post-activation
- `emergency_rollback.sh` - Rollback en cas de problème

### Documentation Externe

- EIP-1 (Ethereum Improvement Proposal process)
- Bitcoin BIP process
- Monero hard fork history

---

## ❓ FAQ

**Q: Combien de temps avant le déploiement ?**
A: Minimum 2 semaines. Recommandé: 4 semaines.

**Q: Que se passe-t-il si certains nœuds ne mettent pas à jour ?**
A: Ils resteront sur l'ancienne chaîne et se sépareront du réseau principal.

**Q: Peut-on retirer une adresse de la blacklist ?**
A: Oui, même processus (hard fork) mais en retirant l'adresse.

**Q: Combien d'adresses peut-on ajouter à la fois ?**
A: Pas de limite technique. Recommandé: grouper plusieurs adresses par update.

**Q: La blacklist affecte-t-elle les transactions ?**
A: Non! Seulement les mining rewards. Les adresses peuvent toujours envoyer/recevoir des transactions normales.

---

## 🎯 Résumé

**Processus simplifié** :
1. **Identifier** l'adresse malveillante
2. **Valider** avec la communauté
3. **Modifier** le code
4. **Coordonner** le hard fork
5. **Déployer** et surveiller

**Temps total** : ~4 semaines du début à la fin

**Fréquence recommandée** : Updates trimestrielles groupées (plutôt qu'une par adresse)
