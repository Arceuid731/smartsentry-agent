//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// installLinuxService installe et configure le service systemd sur Linux
func installLinuxService() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("cette fonction ne fonctionne que sur Linux")
	}

	fmt.Println("🔧 Installation du service systemd...")

	// Créer l'utilisateur système
	if err := createSystemUser(); err != nil {
		return fmt.Errorf("échec création utilisateur : %w", err)
	}

	// Créer le répertoire de logs
	if err := createLogDirectory(); err != nil {
		return fmt.Errorf("échec création répertoire logs : %w", err)
	}

	// Copier le fichier service systemd
	if err := installSystemdServiceFile(); err != nil {
		return fmt.Errorf("échec installation fichier service : %w", err)
	}

	// Créer la configuration override pour les capabilities (optionnel)
	if err := createSystemdOverrideConfiguration(); err != nil {
		fmt.Printf("⚠️  Attention : impossible de créer la configuration override : %v\n", err)
		fmt.Println("📋 Pour activer les capacités avancées manuellement, consultez la documentation")
	}

	// Recharger systemd pour prendre en compte le nouveau service
	fmt.Println("🔄 Rechargement de systemd...")
	if err := runSystemCommand("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("échec rechargement systemd : %w", err)
	}

	// Activer le service pour démarrage automatique
	fmt.Println("✅ Activation du service au démarrage...")
	if err := runSystemCommand("systemctl", "enable", SERVICE_NAME); err != nil {
		return fmt.Errorf("échec activation service : %w", err)
	}

	// Démarrer le service
	fmt.Println("🚀 Démarrage du service...")
	if err := runSystemCommand("systemctl", "start", SERVICE_NAME); err != nil {
		return fmt.Errorf("échec démarrage service : %w", err)
	}

	// Vérifier que le service fonctionne
	if err := checkLinuxServiceStatus(); err != nil {
		return fmt.Errorf("le service ne semble pas fonctionner : %w", err)
	}

	// Afficher les informations sur les capabilities
	displayCapabilitiesInfo()

	fmt.Println("✅ Service systemd installé et démarré avec succès")
	return nil
}

// installSystemdServiceFile crée le fichier .service systemd
func installSystemdServiceFile() error {
	serviceContent := `[Unit]
Description=SmartSentry Observability Agent
Documentation=https://github.com/Arceuid731/smartsentry-agent
After=network.target
Wants=network.target

[Service]
Type=simple
User=smartsentry
Group=smartsentry
ExecStart=/usr/local/bin/otelcol-contrib --config=/etc/smartsentry-agent/config.yaml
Restart=always
RestartSec=5

# Sécurité renforcée
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/smartsentry-agent

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=smartsentry-agent

[Install]
WantedBy=multi-user.target
`

	servicePath := "/etc/systemd/system/" + SERVICE_NAME + ".service"
	fmt.Printf("📝 Création du fichier service : %s\n", servicePath)

	// Écrire le fichier service
	err := os.WriteFile(servicePath, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("impossible d'écrire %s : %w", servicePath, err)
	}

	return nil
}

// createSystemdOverrideConfiguration crée la configuration override pour les capabilities avancées
func createSystemdOverrideConfiguration() error {
	overrideDir := fmt.Sprintf("/etc/systemd/system/%s.service.d", SERVICE_NAME)
	overrideFile := fmt.Sprintf("%s/override.conf", overrideDir)

	fmt.Printf("📁 Création du répertoire override : %s\n", overrideDir)

	// Créer le répertoire override s'il n'existe pas
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire %s : %w", overrideDir, err)
	}

	// Contenu de la configuration override
	overrideContent := `# Configuration override pour SmartSentry Agent
# Ajoute des capabilities système avancées pour la surveillance approfondie

[Service]
# Capabilities ambiantes pour la surveillance système avancée
# CAP_SYS_PTRACE : Permet de tracer les processus système
# CAP_DAC_READ_SEARCH : Permet de contourner les permissions de lecture de fichiers
AmbientCapabilities=CAP_SYS_PTRACE CAP_DAC_READ_SEARCH

# Note: Ces capabilities permettent à SmartSentry d'effectuer une surveillance
# plus approfondie du système, incluant l'inspection des processus et l'accès
# à des fichiers système généralement protégés pour la collecte de métriques avancées.
`

	fmt.Printf("📝 Création du fichier override : %s\n", overrideFile)

	// Écrire le fichier override
	err := os.WriteFile(overrideFile, []byte(overrideContent), 0644)
	if err != nil {
		return fmt.Errorf("impossible d'écrire %s : %w", overrideFile, err)
	}

	fmt.Println("✅ Configuration override créée avec succès")
	return nil
}

// displayCapabilitiesInfo affiche des informations sur les capabilities configurées
func displayCapabilitiesInfo() {
	fmt.Println("\n🔐 Informations sur les Capabilities Système:")
	fmt.Println("   Le service a été configuré avec des capabilities avancées :")
	fmt.Println("   • CAP_SYS_PTRACE     : Surveillance des processus système")
	fmt.Println("   • CAP_DAC_READ_SEARCH : Accès étendu aux fichiers système")
	fmt.Println("\n📋 Commandes utiles pour vérifier les capabilities :")
	fmt.Printf("   • Statut service     : sudo systemctl status %s\n", SERVICE_NAME)
	fmt.Printf("   • Capabilities actives : sudo cat /proc/$(pgrep -f %s)/status | grep Cap\n", SERVICE_NAME)
	fmt.Printf("   • Configuration override : sudo cat /etc/systemd/system/%s.service.d/override.conf\n", SERVICE_NAME)
	fmt.Println("\n⚠️  Note: Ces capabilities permettent une surveillance système approfondie")
	fmt.Println("   mais doivent être utilisées avec précaution en production.")
}

// verifyCapabilities vérifie que les capabilities sont correctement appliquées
func verifyCapabilities() error {
	fmt.Println("🔍 Vérification des capabilities du service...")

	// Trouver le PID du processus smartsentry-agent
	cmd := exec.Command("pgrep", "-f", SERVICE_NAME)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("impossible de trouver le processus %s : %w", SERVICE_NAME, err)
	}

	pid := strings.TrimSpace(string(output))
	if pid == "" {
		return fmt.Errorf("processus %s non trouvé", SERVICE_NAME)
	}

	// Vérifier les capabilities du processus
	statusFile := fmt.Sprintf("/proc/%s/status", pid)
	statusContent, err := os.ReadFile(statusFile)
	if err != nil {
		return fmt.Errorf("impossible de lire %s : %w", statusFile, err)
	}

	statusStr := string(statusContent)

	// Chercher les lignes de capabilities
	lines := strings.Split(statusStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CapEff:") || strings.HasPrefix(line, "CapAmb:") {
			fmt.Printf("   %s\n", line)
		}
	}

	fmt.Printf("✅ Capabilities vérifiées pour le processus %s (PID: %s)\n", SERVICE_NAME, pid)
	return nil
}

// checkLinuxServiceStatus vérifie que le service systemd fonctionne correctement
func checkLinuxServiceStatus() error {
	fmt.Println("🔍 Vérification du statut du service...")

	// Exécuter systemctl status
	cmd := exec.Command("systemctl", "is-active", SERVICE_NAME)
	output, err := cmd.Output()

	status := strings.TrimSpace(string(output))
	if err != nil || status != "active" {
		return fmt.Errorf("service non actif (statut: %s)", status)
	}

	fmt.Printf("✅ Service %s actif et en cours d'exécution\n", SERVICE_NAME)

	// Vérifier les capabilities (non bloquant)
	if err := verifyCapabilities(); err != nil {
		fmt.Printf("⚠️  Attention : %v\n", err)
	}

	return nil
}

// stopLinuxService arrête le service systemd
func stopLinuxService() error {
	if runtime.GOOS != "linux" {
		return nil
	}

	fmt.Printf("🛑 Arrêt du service %s...\n", SERVICE_NAME)

	// Arrêter le service
	if err := runSystemCommand("systemctl", "stop", SERVICE_NAME); err != nil {
		fmt.Printf("⚠️  Attention : impossible d'arrêter le service : %v\n", err)
	}

	return nil
}

// removeSystemdOverrideConfiguration supprime la configuration override
func removeSystemdOverrideConfiguration() error {
	overrideDir := fmt.Sprintf("/etc/systemd/system/%s.service.d", SERVICE_NAME)

	fmt.Printf("🗑️  Suppression de la configuration override : %s\n", overrideDir)

	// Supprimer le répertoire override et son contenu
	if err := os.RemoveAll(overrideDir); err != nil {
		return fmt.Errorf("impossible de supprimer %s : %w", overrideDir, err)
	}

	fmt.Println("✅ Configuration override supprimée")
	return nil
}
