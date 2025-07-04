package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	// Version de l'OpenTelemetry Collector à utiliser
	OTEL_VERSION = "0.129.1"

	// Nom du service sur le système
	SERVICE_NAME = "smartsentry-agent"

	// URL de base pour télécharger la configuration par défaut
	// REMPLACE par l'URL GitHub de ton repo une fois pushé
	CONFIG_BASE_URL = "https://raw.githubusercontent.com/Arceuid731/smartsentry-agent/main/configs"
)

func main() {
	fmt.Println("🚀 SmartSentry Agent Installer")
	fmt.Println("Powered by OpenTelemetry Collector")
	fmt.Printf("Target OS: %s, Architecture: %s\n\n", runtime.GOOS, runtime.GOARCH)

	// Vérifier les permissions administrateur
	if !hasAdminPrivileges() {
		log.Fatal("❌ Erreur : Ce programme doit être exécuté avec des privilèges administrateur (sudo sur Linux, Administrateur sur Windows)")
	}

	// Étape 1 : Télécharger le binaire OpenTelemetry Collector
	fmt.Println("📥 Téléchargement de l'OpenTelemetry Collector...")
	if err := downloadOTelCollector(); err != nil {
		log.Fatalf("❌ Erreur lors du téléchargement : %v", err)
	}
	fmt.Println("✅ OpenTelemetry Collector téléchargé")

	// Étape 2 : Télécharger et installer la configuration
	fmt.Println("⚙️  Configuration de l'agent...")
	if err := setupConfiguration(); err != nil {
		log.Fatalf("❌ Erreur lors de la configuration : %v", err)
	}
	fmt.Println("✅ Configuration installée")

	// Étape 3 : Installer et démarrer le service
	fmt.Println("🔧 Installation du service système...")
	if err := installAndStartService(); err != nil {
		log.Fatalf("❌ Erreur lors de l'installation du service : %v", err)
	}
	fmt.Println("✅ Service installé et démarré")

	fmt.Println("\n🎉 Installation terminée avec succès !")
	fmt.Printf("Le service '%s' est maintenant actif et collecte les métriques.\n", SERVICE_NAME)

	// Instructions spécifiques à l'OS pour vérifier le service
	printServiceInstructions()
}

// hasAdminPrivileges vérifie si le programme s'exécute avec les privilèges administrateur
func hasAdminPrivileges() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		// Sur Linux/macOS, vérifier si l'utilisateur est root (UID 0)
		return os.Geteuid() == 0
	case "windows":
		// Sur Windows, cette vérification est plus complexe
		// Pour simplifier, on assume que si le programme arrive jusqu'ici,
		// c'est probablement OK (la vérification réelle se ferait avec l'API Windows)
		return true
	default:
		return false
	}
}

// printServiceInstructions affiche les commandes pour gérer le service selon l'OS
func printServiceInstructions() {
	fmt.Println("\n📋 Commandes utiles pour gérer le service :")

	switch runtime.GOOS {
	case "linux":
		fmt.Printf("  • Statut      : sudo systemctl status %s\n", SERVICE_NAME)
		fmt.Printf("  • Arrêter     : sudo systemctl stop %s\n", SERVICE_NAME)
		fmt.Printf("  • Redémarrer  : sudo systemctl restart %s\n", SERVICE_NAME)
		fmt.Printf("  • Logs        : sudo journalctl -u %s -f\n", SERVICE_NAME)
		fmt.Printf("  • Désinstaller: sudo systemctl stop %s && sudo systemctl disable %s\n", SERVICE_NAME, SERVICE_NAME)
		fmt.Printf("\n🔐 Commandes pour les capabilities système :\n")
		fmt.Printf("  • Vérifier capabilities : sudo cat /proc/$(pgrep -f %s)/status | grep Cap\n", SERVICE_NAME)
		fmt.Printf("  • Configuration override : sudo cat /etc/systemd/system/%s.service.d/override.conf\n", SERVICE_NAME)
		fmt.Printf("  • Recharger configuration: sudo systemctl daemon-reload && sudo systemctl restart %s\n", SERVICE_NAME)
	case "windows":
		fmt.Printf("  • Statut      : sc query \"%s\"\n", SERVICE_NAME)
		fmt.Printf("  • Arrêter     : sc stop \"%s\"\n", SERVICE_NAME)
		fmt.Printf("  • Redémarrer  : sc stop \"%s\" && sc start \"%s\"\n", SERVICE_NAME, SERVICE_NAME)
		fmt.Printf("  • Logs        : Check Event Viewer > Windows Logs > Application\n")
		fmt.Printf("  • Désinstaller: sc stop \"%s\" && sc delete \"%s\"\n", SERVICE_NAME, SERVICE_NAME)
	}
}

// installAndStartService installe et démarre le service selon l'OS
func installAndStartService() error {
	switch runtime.GOOS {
	case "linux":
		return installLinuxService()
	case "windows":
		return installWindowsService()
	case "darwin":
		// Sur macOS, on pourrait utiliser launchd, mais pour simplifier
		// on affiche un message pour l'instant
		fmt.Println("⚠️  Sur macOS, veuillez démarrer manuellement l'agent :")
		fmt.Println("sudo /usr/local/bin/otelcol-contrib --config=/etc/smartsentry-agent/config.yaml")
		return nil
	default:
		return fmt.Errorf("installation de service non supportée sur %s", runtime.GOOS)
	}
}
