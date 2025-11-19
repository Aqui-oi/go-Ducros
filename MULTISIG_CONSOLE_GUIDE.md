# Guide Pratique : Blacklist avec Multi-Sig 3/5 via Console

## 🎯 Vue d'Ensemble

Au lieu de faire un hard fork à chaque fois, tu peux utiliser un système multi-sig où **3 personnes sur 5** doivent approuver pour blacklister une adresse.

---

## Étape 1 : Déployer le Contrat Multi-Sig (Une Seule Fois)

### A. Préparer les Adresses des 5 Signers

```javascript
// Dans la console geth
var signer1 = "0xTA_PREMIERE_ADRESSE";     // Toi (fondateur)
var signer2 = "0xADRESSE_DEV_PRINCIPAL";   // Dev de confiance
var signer3 = "0xADRESSE_MEMBRE_1";        // Membre communauté
var signer4 = "0xADRESSE_MEMBRE_2";        // Membre communauté
var signer5 = "0xADRESSE_PARTENAIRE";      // Partenaire technique

var signers = [signer1, signer2, signer3, signer4, signer5];
var requiredSigs = 3;  // 3 signatures sur 5
```

### B. Compiler le Smart Contract

Sauvegarde ce fichier : `BlacklistMultiSig.sol`

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract BlacklistMultiSig {
    address[] public signers;
    mapping(address => bool) public isSigner;
    uint256 public requiredSignatures;

    mapping(address => bool) public blacklisted;
    mapping(address => uint256) public blacklistedSince;

    struct Proposal {
        address target;
        bool toBlacklist;
        string reason;
        uint256 createdAt;
        bool executed;
        mapping(address => bool) signed;
        uint256 signatureCount;
    }

    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalCount;

    event ProposalCreated(uint256 indexed id, address indexed target, bool toBlacklist);
    event ProposalSigned(uint256 indexed id, address indexed signer);
    event ProposalExecuted(uint256 indexed id, address indexed target);
    event AddressBlacklisted(address indexed target, uint256 timestamp);
    event AddressUnblacklisted(address indexed target, uint256 timestamp);

    modifier onlySigner() {
        require(isSigner[msg.sender], "Not a signer");
        _;
    }

    constructor(address[] memory _signers, uint256 _requiredSigs) {
        require(_signers.length >= _requiredSigs, "Not enough signers");

        for (uint256 i = 0; i < _signers.length; i++) {
            signers.push(_signers[i]);
            isSigner[_signers[i]] = true;
        }

        requiredSignatures = _requiredSigs;
    }

    function propose(address _target, bool _toBlacklist, string memory _reason)
        external onlySigner returns (uint256)
    {
        uint256 id = proposalCount++;
        Proposal storage prop = proposals[id];

        prop.target = _target;
        prop.toBlacklist = _toBlacklist;
        prop.reason = _reason;
        prop.createdAt = block.timestamp;
        prop.signed[msg.sender] = true;
        prop.signatureCount = 1;

        emit ProposalCreated(id, _target, _toBlacklist);

        if (requiredSignatures == 1) {
            _execute(id);
        }

        return id;
    }

    function sign(uint256 _id) external onlySigner {
        Proposal storage prop = proposals[_id];
        require(!prop.executed, "Already executed");
        require(!prop.signed[msg.sender], "Already signed");

        prop.signed[msg.sender] = true;
        prop.signatureCount++;

        emit ProposalSigned(_id, msg.sender);

        if (prop.signatureCount >= requiredSignatures) {
            _execute(_id);
        }
    }

    function _execute(uint256 _id) internal {
        Proposal storage prop = proposals[_id];
        require(!prop.executed, "Already executed");

        prop.executed = true;

        if (prop.toBlacklist) {
            blacklisted[prop.target] = true;
            blacklistedSince[prop.target] = block.timestamp;
            emit AddressBlacklisted(prop.target, block.timestamp);
        } else {
            blacklisted[prop.target] = false;
            emit AddressUnblacklisted(prop.target, block.timestamp);
        }

        emit ProposalExecuted(_id, prop.target);
    }

    function isBlacklisted(address _addr) external view returns (bool) {
        return blacklisted[_addr];
    }

    function getProposal(uint256 _id) external view returns (
        address target,
        bool toBlacklist,
        string memory reason,
        uint256 signatureCount,
        bool executed
    ) {
        Proposal storage prop = proposals[_id];
        return (prop.target, prop.toBlacklist, prop.reason, prop.signatureCount, prop.executed);
    }

    function hasSigned(uint256 _id, address _signer) external view returns (bool) {
        return proposals[_id].signed[_signer];
    }
}
```

### C. Compiler avec Solc

```bash
# Installe solc si pas déjà fait
npm install -g solc

# Compile le contrat
solc --abi --bin --optimize BlacklistMultiSig.sol -o build/

# Cela crée:
# - build/BlacklistMultiSig.abi
# - build/BlacklistMultiSig.bin
```

### D. Déployer depuis la Console Geth

```javascript
// 1. Lance la console geth
./build/bin/geth attach

// 2. Unlock ton compte
personal.unlockAccount(eth.accounts[0], "ton_mot_de_passe")

// 3. Charge l'ABI et le bytecode
var abi = [/* copie le contenu de BlacklistMultiSig.abi */];
var bytecode = "0x..."; // copie le contenu de BlacklistMultiSig.bin

// 4. Crée le contrat
var BlacklistContract = eth.contract(abi);

// 5. Définis les signers
var signers = [
    "0xSIGNER_1",
    "0xSIGNER_2",
    "0xSIGNER_3",
    "0xSIGNER_4",
    "0xSIGNER_5"
];

// 6. Déploie!
var blacklistInstance = BlacklistContract.new(
    signers,
    3,  // 3 signatures requises
    {
        from: eth.accounts[0],
        data: bytecode,
        gas: 3000000
    },
    function(error, contract) {
        if (!error) {
            if (contract.address) {
                console.log("✅ Contrat déployé à:", contract.address);
                console.log("⚠️  Sauvegarde cette adresse!");
            }
        } else {
            console.log("❌ Erreur:", error);
        }
    }
);

// Attends quelques secondes...
// Tu verras: "✅ Contrat déployé à: 0x1111111111111111111111111111111111111111"
```

---

## Étape 2 : Connecter au Contrat (Après Déploiement)

```javascript
// 1. Lance geth console
./build/bin/geth attach

// 2. Charge l'ABI
var abi = [/* même ABI que ci-dessus */];

// 3. Adresse du contrat déployé (celle que tu as sauvegardée)
var contractAddress = "0x1111111111111111111111111111111111111111";

// 4. Crée l'instance
var blacklist = eth.contract(abi).at(contractAddress);

// 5. Vérifie que ça fonctionne
console.log("Signers:", blacklist.signers(0), blacklist.signers(1), blacklist.signers(2));
console.log("Signatures requises:", blacklist.requiredSignatures());
console.log("Propositions:", blacklist.proposalCount());
```

---

## Étape 3 : Blacklister une Adresse (Signer 1)

### Scénario : Tu détectes un botnet à l'adresse 0x742d...

```javascript
// 1. Lance console en tant que Signer 1
./build/bin/geth attach

// 2. Connecte au contrat (comme ci-dessus)
var blacklist = eth.contract(abi).at(contractAddress);

// 3. Unlock ton compte
personal.unlockAccount(eth.accounts[0], "mot_de_passe")

// 4. Adresse à blacklister
var botnetAddress = "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb";
var reason = "Confirmed botnet - 1000+ infected machines - Evidence: https://github.com/project/issues/123";

// 5. Crée la proposition
var proposalTx = blacklist.propose(
    botnetAddress,
    true,  // true = blacklist, false = remove
    reason,
    {
        from: eth.accounts[0],
        gas: 300000
    }
);

console.log("✅ Proposition créée! Transaction:", proposalTx);

// 6. Attends confirmation (quelques secondes)
// Puis récupère le proposal ID
var proposalId = blacklist.proposalCount() - 1;
console.log("📋 Proposal ID:", proposalId);

// 7. Vérifie la proposition
var prop = blacklist.getProposal(proposalId);
console.log("Target:", prop[0]);
console.log("To blacklist:", prop[1]);
console.log("Reason:", prop[2]);
console.log("Signatures:", prop[3].toString(), "/ 3 required");
console.log("Executed:", prop[4]);
```

**Résultat** :
```
✅ Proposition créée! Transaction: 0xabc123...
📋 Proposal ID: 0
Target: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
To blacklist: true
Reason: Confirmed botnet - 1000+ infected machines...
Signatures: 1 / 3 required
Executed: false
```

---

## Étape 4 : Signer la Proposition (Signer 2)

### Le Signer 2 doit maintenant approuver

```javascript
// 1. Signer 2 lance sa console
./build/bin/geth attach

// 2. Connecte au contrat
var abi = [/* même ABI */];
var contractAddress = "0x1111111111111111111111111111111111111111";
var blacklist = eth.contract(abi).at(contractAddress);

// 3. Vérifie la proposition
var proposalId = 0;  // L'ID de la proposition
var prop = blacklist.getProposal(proposalId);

console.log("📋 Proposition #" + proposalId);
console.log("   Target:", prop[0]);
console.log("   Action:", prop[1] ? "BLACKLIST" : "REMOVE");
console.log("   Reason:", prop[2]);
console.log("   Signatures:", prop[3].toString() + " / 3");

// 4. Vérifie si tu as déjà signé
var alreadySigned = blacklist.hasSigned(proposalId, eth.accounts[0]);
console.log("   Already signed:", alreadySigned);

// 5. Si d'accord, SIGNE
personal.unlockAccount(eth.accounts[0], "mot_de_passe")

var signTx = blacklist.sign(
    proposalId,
    {
        from: eth.accounts[0],
        gas: 200000
    }
);

console.log("✅ Signature ajoutée! Transaction:", signTx);

// 6. Vérifie les signatures
var prop2 = blacklist.getProposal(proposalId);
console.log("Signatures maintenant:", prop2[3].toString() + " / 3");
```

**Résultat** :
```
📋 Proposition #0
   Target: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
   Action: BLACKLIST
   Reason: Confirmed botnet - 1000+ infected machines...
   Signatures: 1 / 3
   Already signed: false

✅ Signature ajoutée! Transaction: 0xdef456...
Signatures maintenant: 2 / 3
```

---

## Étape 5 : Signature Finale (Signer 3) → Exécution Automatique!

```javascript
// 1. Signer 3 lance console
./build/bin/geth attach

// 2. Connecte au contrat
var blacklist = eth.contract(abi).at(contractAddress);

// 3. Vérifie la proposition
var proposalId = 0;
var prop = blacklist.getProposal(proposalId);
console.log("Signatures:", prop[3].toString() + " / 3");

// 4. Signe (dernière signature!)
personal.unlockAccount(eth.accounts[0], "mot_de_passe")

var signTx = blacklist.sign(proposalId, {
    from: eth.accounts[0],
    gas: 200000
});

console.log("✅ Dernière signature! Transaction:", signTx);

// 5. Attends quelques secondes, puis vérifie
setTimeout(function() {
    var prop2 = blacklist.getProposal(proposalId);
    console.log("Signatures:", prop2[3].toString());
    console.log("Executed:", prop2[4]);  // Devrait être TRUE!

    // Vérifie que l'adresse est blacklistée
    var isBlacklisted = blacklist.isBlacklisted("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb");
    console.log("🚫 Address blacklisted:", isBlacklisted);
}, 3000);
```

**Résultat** :
```
✅ Dernière signature! Transaction: 0xghi789...

(3 secondes plus tard)
Signatures: 3
Executed: true
🚫 Address blacklisted: true

🎉 L'adresse est maintenant blacklistée!
   Au prochain bloc, le consensus appliquera la blacklist.
```

---

## Étape 6 : Vérifier l'Impact

### Vérifier que l'adresse ne reçoit plus de rewards

```javascript
// Vérifie le solde avant/après mining
var botnet = "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb";

// Solde actuel
var balanceBefore = eth.getBalance(botnet);
console.log("Balance avant:", web3.fromWei(balanceBefore, "ether"), "DCR");

// Attends que l'adresse mine un bloc (ou force si c'est ton testnet)
// ...quelques minutes...

// Solde après
var balanceAfter = eth.getBalance(botnet);
console.log("Balance après:", web3.fromWei(balanceAfter, "ether"), "DCR");

// Différence
var reward = balanceAfter - balanceBefore;
if (reward == 0) {
    console.log("✅ SUCCÈS: Aucune mining reward reçue!");
    console.log("   L'adresse est bien blacklistée.");
} else {
    console.log("⚠️  WARNING: A reçu", web3.fromWei(reward, "ether"), "DCR");
    console.log("   La blacklist n'est peut-être pas encore active.");
}
```

---

## 📋 Récapitulatif : Process Complet

### Pour Blacklister une Adresse (3/5 Multi-Sig)

```
Jour 1, 09:00 - Signer 1 (Toi)
├─ Détecte botnet: 0x742d...
├─ blacklist.propose(0x742d..., true, "Reason")
├─ Coût: ~0.015 DCR
└─ Status: 1/3 signatures ⏳

Jour 1, 14:00 - Signer 2 (Dev)
├─ Vérifie la proposition
├─ blacklist.sign(proposalId)
├─ Coût: ~0.005 DCR
└─ Status: 2/3 signatures ⏳

Jour 1, 18:00 - Signer 3 (Communauté)
├─ Vérifie la proposition
├─ blacklist.sign(proposalId)
├─ Coût: ~0.005 DCR
└─ ✅ EXÉCUTION AUTOMATIQUE!

Jour 1, 18:01 - Consensus
├─ Prochain bloc miné
├─ Consensus lit blacklist.isBlacklisted(0x742d...)
├─ Résultat: true
└─ Reward: 0% mineur, 100% trésorerie

Total:
├─ Temps: <24 heures 🚀
├─ Coût: ~0.025 DCR 💰
├─ Permanent: Oui ✅
└─ Réversible: Oui (même process pour retirer) ✅
```

---

## 🔄 Retirer une Adresse de la Blacklist

Si tu as fait une erreur ou si l'adresse est légitime :

```javascript
// Même process, mais toBlacklist = false

// Signer 1
blacklist.propose(
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    false,  // false = retirer de la blacklist
    "False positive - verified legitimate miner",
    {from: eth.accounts[0], gas: 300000}
);

// Signers 2 et 3
blacklist.sign(proposalId, {from: eth.accounts[0], gas: 200000});

// Après 3 signatures → Retiré automatiquement!
```

---

## 🛠️ Scripts Utiles

### Script: Vérifier toutes les propositions

```javascript
// check_proposals.js
var proposalCount = blacklist.proposalCount();
console.log("Total proposals:", proposalCount.toString());

for (var i = 0; i < proposalCount; i++) {
    var prop = blacklist.getProposal(i);
    console.log("\n📋 Proposal #" + i);
    console.log("   Target:", prop[0]);
    console.log("   Action:", prop[1] ? "BLACKLIST" : "REMOVE");
    console.log("   Reason:", prop[2]);
    console.log("   Signatures:", prop[3].toString() + " / 3");
    console.log("   Executed:", prop[4] ? "✅" : "⏳");
}
```

### Script: Lister tous les signers

```javascript
// list_signers.js
console.log("Signers:");
for (var i = 0; i < 5; i++) {
    var signer = blacklist.signers(i);
    var isSigner = blacklist.isSigner(signer);
    console.log((i+1) + ".", signer, isSigner ? "✅" : "❌");
}
```

### Script: Vérifier si une adresse est blacklistée

```javascript
// check_blacklist.js
function checkBlacklist(addr) {
    var isBlacklisted = blacklist.isBlacklisted(addr);
    console.log("Address:", addr);
    console.log("Blacklisted:", isBlacklisted ? "🚫 YES" : "✅ NO");

    if (isBlacklisted) {
        var since = blacklist.blacklistedSince(addr);
        var date = new Date(since * 1000);
        console.log("Since:", date.toLocaleString());
    }
}

// Usage
checkBlacklist("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb");
```

---

## ⚠️ Points Importants

### 1. Garde les Clés Privées des Signers en Sécurité!

```
Signer 1 (Toi): Cold wallet + backup
Signer 2-3: Multi-sig hardware wallets recommandés
Signer 4-5: Membres de confiance avec bonnes pratiques sécurité
```

### 2. Communication entre Signers

Avant de signer, les signers doivent:
- ✅ Vérifier les preuves (GitHub issue, logs, etc.)
- ✅ Confirmer l'impact (hashrate, DCR détourné)
- ✅ Discuter sur Discord/Telegram
- ✅ Consensus avant de signer

### 3. Coût en Gas

```
Action                  Gas      Coût (si 1 gwei)
propose()              ~150k     ~0.015 DCR
sign()                 ~50k      ~0.005 DCR
Total (3 sigs)         ~250k     ~0.025 DCR
```

### 4. Hard Fork Initial Nécessaire

⚠️ Ce système nécessite **UN hard fork initial** pour que le consensus lise le smart contract.

Après ce hard fork unique, **plus jamais de hard fork** pour les blacklists!

---

## 🎯 Avantages vs Hard Fork

| Critère | Hard Fork | Multi-Sig 3/5 |
|---------|-----------|---------------|
| Temps | 6 semaines | <24 heures |
| Coordination | Tous les nodes | 3 personnes |
| Coût | Gratuit | 0.025 DCR |
| Réversible | Non | Oui |
| Permanent | Oui | Oui |

**Multi-Sig = Meilleur compromis ! 🎉**
