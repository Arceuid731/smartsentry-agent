#!/bin/bash
# Script d'installation SmartSentry Agent pour Linux/macOS
# Usage: curl -sSL https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.sh | sudo bash

set -e  # Arrêter en cas d'erreur

REPO_URL="https://github.com/Arceuid731/smartsentry-agent"
LATEST_RELEASE_URL="$REPO_URL/releases/latest"
BINARY_NAME="smartsentry-installer-linux-amd64"

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

print_status "🚀 Installation de SmartSentry Agent"

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

# Créer un répertoire temporaire
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

print_status "📥 Téléchargement de l'installateur..."

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

print_status "🔧 Lancement de l'installation..."

# Exécuter l'installateur
"$TEMP_DIR/installer"

print_success "✅ Installation terminée !"
print_status "Le service SmartSentry Agent est maintenant actif"

# Afficher les commandes utiles
echo ""
print_status "📋 Commandes utiles :"
if [[ "$OS" == "linux" ]]; then
    echo "  • Statut      : sudo systemctl status smartsentry-agent"
    echo "  • Logs        : sudo journalctl -u smartsentry-agent -f"
    echo "  • Redémarrer  : sudo systemctl restart smartsentry-agent"
    echo "  • Arrêter     : sudo systemctl stop smartsentry-agent"
fi

echo ""
print_status "📖 Documentation: $REPO_URL"
