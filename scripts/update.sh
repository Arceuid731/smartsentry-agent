#!/bin/bash
# Script de mise à jour SmartSentry Agent pour Linux/macOS
# Usage: curl -sSL https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/update.sh | sudo bash

set -e  # Arrêter en cas d'erreur

REPO_URL="https://github.com/Arceuid731/smartsentry-agent"
LATEST_RELEASE_URL="$REPO_URL/releases/latest"
BINARY_NAME="smartsentry-installer-linux-amd64"
SERVICE_NAME="smartsentry-agent"
CONFIG_DIR="/etc/smartsentry-agent"
BACKUP_DIR="/etc/smartsentry-agent/backups"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction d'affichage avec couleurs
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Vérifier les privilèges root
if [[ $EUID -ne 0 ]]; then
   print_error "Ce script doit être exécuté avec sudo ou en tant que root"
   exit 1
fi

print_status "🔄 Mise à jour de SmartSentry Agent"

# Vérifier si l'agent est déjà installé
if ! systemctl is-enabled "$SERVICE_NAME" >/dev/null 2>&1; then
    print_error "SmartSentry Agent n'est pas installé. Utilisez le script d'installation à la place."
    print_status "Installation: curl -sSL https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.sh | sudo bash"
    exit 1
fi

# Détecter l'OS et l'architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) 
        print_error "Architecture non supportée : $ARCH"
        exit 1
        ;;
esac

# Ajuster le nom du binaire selon l'OS
if [[ "$OS" == "darwin" ]]; then
    BINARY_NAME="smartsentry-installer-darwin-$ARCH"
else
    BINARY_NAME="smartsentry-installer-linux-$ARCH"
fi

print_status "OS détecté: $OS, Architecture: $ARCH"

# Créer le répertoire de sauvegarde
mkdir -p "$BACKUP_DIR"

# Sauvegarder la configuration actuelle
print_status "💾 Sauvegarde de la configuration actuelle..."
BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/config_backup_$BACKUP_DATE.yaml"

if [[ -f "$CONFIG_DIR/config.yaml" ]]; then
    cp "$CONFIG_DIR/config.yaml" "$BACKUP_FILE"
    print_success "Configuration sauvegardée dans $BACKUP_FILE"
else
    print_warning "Aucune configuration trouvée à sauvegarder"
fi

# Arrêter le service
print_status "🛑 Arrêt du service SmartSentry Agent..."
if systemctl is-active --quiet "$SERVICE_NAME"; then
    systemctl stop "$SERVICE_NAME"
    print_success "Service arrêté"
else
    print_warning "Le service n'était pas en cours d'exécution"
fi

# Créer un répertoire temporaire
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

print_status "📥 Téléchargement de la nouvelle version..."

# Obtenir l'URL de la dernière release
DOWNLOAD_URL="$REPO_URL/releases/latest/download/$BINARY_NAME"

# Télécharger l'installateur
if command -v curl >/dev/null 2>&1; then
    curl -sSL "$DOWNLOAD_URL" -o "$TEMP_DIR/installer"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TEMP_DIR/installer"
else
    print_error "curl ou wget requis pour télécharger l'installateur"
    exit 1
fi

# Vérifier que le téléchargement a réussi
if [[ ! -f "$TEMP_DIR/installer" ]]; then
    print_error "Échec du téléchargement de l'installateur"
    exit 1
fi

# Rendre le binaire exécutable
chmod +x "$TEMP_DIR/installer"

print_status "🔧 Lancement de la mise à jour..."

# Exécuter l'installateur (il détectera automatiquement l'installation existante)
"$TEMP_DIR/installer"

print_status "🚀 Redémarrage du service..."
systemctl daemon-reload
systemctl start "$SERVICE_NAME"
systemctl enable "$SERVICE_NAME"

# Vérifier que le service est bien démarré
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    print_success "✅ Mise à jour terminée avec succès !"
    print_status "Le service SmartSentry Agent est maintenant actif avec la nouvelle version"
else
    print_error "❌ Problème lors du redémarrage du service"
    print_status "Vérifiez les logs avec : sudo journalctl -u $SERVICE_NAME -f"
    exit 1
fi

# Afficher les informations de debug
print_status "📊 Informations de débogage :"
echo "  • Les logs sont maintenant disponibles dans : /var/log/smartsentry-agent/"
echo "  • Logs du collecteur : /var/log/smartsentry-agent/collector.log"
echo "  • Logs d'erreurs : /var/log/smartsentry-agent/collector-error.log"

# Afficher les commandes utiles
echo ""
print_status "📋 Commandes utiles :"
if [[ "$OS" == "linux" ]]; then
    echo "  • Statut      : sudo systemctl status smartsentry-agent"
    echo "  • Logs système: sudo journalctl -u smartsentry-agent -f"
    echo "  • Logs agent  : sudo tail -f /var/log/smartsentry-agent/collector.log"
    echo "  • Redémarrer  : sudo systemctl restart smartsentry-agent"
    echo "  • Arrêter     : sudo systemctl stop smartsentry-agent"
fi

echo ""
print_status "📖 Documentation: $REPO_URL"
print_status "💾 Sauvegarde disponible: $BACKUP_FILE"