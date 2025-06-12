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
