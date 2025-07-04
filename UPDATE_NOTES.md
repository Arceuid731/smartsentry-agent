# SmartSentry Agent - Notes de Mise à Jour

## Version mise à jour - OpenTelemetry Collector 0.129.0

### Résumé des modifications

#### 1. Mise à jour de la version OpenTelemetry Collector
- **Version précédente** : 0.128.0
- **Version actuelle** : 0.129.0
- **Distribution** : opentelemetry-collector-contrib
- **Fichier modifié** : `installer/main.go`

#### 2. Configuration améliorée du débogage
- **Nouveau** : Logs dans des fichiers séparés au lieu de journalctl uniquement
- **Répertoire** : `/var/log/smartsentry-agent/`
- **Fichiers de logs** :
  - `collector.log` : Logs normaux du collecteur
  - `collector-error.log` : Logs d'erreurs
- **Configuration** : `configs/linux-default-config.yaml` - section `service.telemetry.logs`

#### 3. Amélioration des métriques mémoire
- **Problème résolu** : Métriques `system.memory.usage` incomplètes
- **Configuration** : Métriques mémoire et CPU explicitement définies
- **Accès** : Utilisation optimisée de `/proc/meminfo` pour les détails complets

#### 4. Correction du script de build
- **Problème résolu** : Erreurs de chemins lors de la compilation
- **Modification** : Utilisation de sous-shells pour les changements de répertoire
- **Résultat** : Compilation cross-platform fonctionnelle

#### 5. Nouveaux scripts de mise à jour
- **Linux/macOS** : `scripts/update.sh`
- **Windows** : `scripts/update.ps1`
- **Fonctionnalités** :
  - Sauvegarde automatique de la configuration
  - Arrêt/redémarrage du service
  - Vérification de l'installation existante
  - Gestion des erreurs robuste

### Processus de mise à jour

#### Installation initiale (one-liner)
```bash
# Linux/macOS
curl -sSL https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.sh | sudo bash

# Windows
Invoke-WebRequest -Uri "https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.ps1" -UseBasicParsing | Invoke-Expression
```

#### Mise à jour (nouveaux scripts)
```bash
# Linux/macOS
curl -sSL https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/update.sh | sudo bash

# Windows
Invoke-WebRequest -Uri "https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/update.ps1" -UseBasicParsing | Invoke-Expression
```

### Nouvelles fonctionnalités de débogage

#### Accès aux logs
```bash
# Logs du collecteur (nouveaux fichiers)
sudo tail -f /var/log/smartsentry-agent/collector.log
sudo tail -f /var/log/smartsentry-agent/collector-error.log

# Logs système (toujours disponibles)
sudo journalctl -u smartsentry-agent -f
```

#### Vérification des métriques
```bash
# Statut du service
sudo systemctl status smartsentry-agent

# Vérification des capabilities
sudo cat /proc/$(pgrep -f smartsentry-agent)/status | grep Cap

# Test accès /proc/meminfo
sudo -u smartsentry cat /proc/meminfo
```

### Métriques améliorées

#### Métriques mémoire maintenant collectées
- `system.memory.usage` : Utilisation détaillée de la mémoire
- `system.memory.utilization` : Pourcentage d'utilisation
- Accès complet aux données `/proc/meminfo`

#### Métriques CPU étendues
- `system.cpu.time` : Temps CPU détaillé
- `system.cpu.utilization` : Pourcentage d'utilisation

### Permissions et sécurité

#### Permissions SystemD
- **ReadWritePaths** : `/var/log/smartsentry-agent`
- **Capabilities** : `CAP_SYS_PTRACE`, `CAP_DAC_READ_SEARCH`
- **Utilisateur** : `smartsentry` (non-root)

### Fichiers de sauvegarde

#### Lors de la mise à jour
- **Répertoire** : `/etc/smartsentry-agent/backups/`
- **Format** : `config_backup_YYYYMMDD_HHMMSS.yaml`
- **Automatique** : Sauvegarde avant chaque mise à jour

### Commandes utiles

#### Gestion du service
```bash
# Statut
sudo systemctl status smartsentry-agent

# Redémarrage
sudo systemctl restart smartsentry-agent

# Arrêt
sudo systemctl stop smartsentry-agent

# Logs en temps réel
sudo journalctl -u smartsentry-agent -f
```

#### Débogage avancé
```bash
# Logs détaillés du collecteur
sudo tail -f /var/log/smartsentry-agent/collector.log

# Logs d'erreurs uniquement
sudo tail -f /var/log/smartsentry-agent/collector-error.log

# Configuration actuelle
sudo cat /etc/smartsentry-agent/config.yaml

# Vérification des métriques en local (si port exposé)
curl -s http://localhost:8888/metrics
```

### Prochaines étapes

1. **Tester la mise à jour** sur une machine de test
2. **Déployer sur les serveurs de production** via le script de mise à jour
3. **Vérifier la collecte des métriques** améliorées
4. **Monitorer les logs** dans les nouveaux fichiers
5. **Créer une release GitHub** avec les binaires compilés

### Troubleshooting

#### Si la mise à jour échoue
1. Vérifier les logs : `sudo journalctl -u smartsentry-agent -f`
2. Restaurer la configuration : `sudo cp /etc/smartsentry-agent/backups/config_backup_*.yaml /etc/smartsentry-agent/config.yaml`
3. Redémarrer le service : `sudo systemctl restart smartsentry-agent`

#### Si les métriques mémoire sont toujours incomplètes
1. Vérifier les permissions : `sudo -u smartsentry cat /proc/meminfo`
2. Vérifier la configuration : `sudo cat /etc/smartsentry-agent/config.yaml`
3. Consulter les logs d'erreurs : `sudo tail -f /var/log/smartsentry-agent/collector-error.log`