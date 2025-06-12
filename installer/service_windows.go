//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// installWindowsService installe et configure le service Windows
func installWindowsService() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("cette fonction ne fonctionne que sur Windows")
	}

	fmt.Println("🔧 Installation du service Windows...")

	// Créer le répertoire de logs
	if err := createLogDirectory(); err != nil {
		return fmt.Errorf("échec création répertoire logs : %w", err)
	}

	// Chemin vers le binaire otelcol-contrib
	binaryPath := `C:\Program Files\SmartSentry\otelcol-contrib.exe`
	configPath, err := getConfigDirectory()
	if err != nil {
		return fmt.Errorf("impossible de déterminer le répertoire de config : %w", err)
	}
	configFile := filepath.Join(configPath, "config.yaml")

	// Commande pour créer le service Windows
	// sc create : crée un nouveau service
	serviceCmd := fmt.Sprintf(
		`sc create "%s" binPath= "\"%s\" --config=\"%s\"" start= auto DisplayName= "SmartSentry Observability Agent"`,
		SERVICE_NAME,
		binaryPath,
		configFile,
	)

	fmt.Println("📝 Création du service Windows...")
	if err := runWindowsCommand(serviceCmd); err != nil {
		return fmt.Errorf("échec création service Windows : %w", err)
	}

	// Configurer la description du service
	descCmd := fmt.Sprintf(
		`sc description "%s" "SmartSentry Agent collecte les métriques système et les envoie au Gateway SmartSentry"`,
		SERVICE_NAME,
	)
	runWindowsCommand(descCmd) // Ignorer l'erreur, c'est optionnel

	// Démarrer le service
	fmt.Println("🚀 Démarrage du service...")
	if err := runWindowsCommand(fmt.Sprintf(`sc start "%s"`, SERVICE_NAME)); err != nil {
		return fmt.Errorf("échec démarrage service : %w", err)
	}

	// Vérifier que le service fonctionne
	if err := checkWindowsServiceStatus(); err != nil {
		return fmt.Errorf("le service ne semble pas fonctionner : %w", err)
	}

	fmt.Println("✅ Service Windows installé et démarré avec succès")
	return nil
}

// checkWindowsServiceStatus vérifie que le service Windows fonctionne
func checkWindowsServiceStatus() error {
	fmt.Println("🔍 Vérification du statut du service...")

	// Utiliser sc query pour vérifier le statut
	cmd := exec.Command("sc", "query", SERVICE_NAME)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("impossible de vérifier le statut du service : %w", err)
	}

	outputStr := string(output)

	// Chercher "RUNNING" dans la sortie
	if strings.Contains(outputStr, "RUNNING") {
		fmt.Printf("✅ Service %s actif et en cours d'exécution\n", SERVICE_NAME)
		return nil
	} else if strings.Contains(outputStr, "STOPPED") {
		return fmt.Errorf("service arrêté")
	} else if strings.Contains(outputStr, "START_PENDING") {
		return fmt.Errorf("service en cours de démarrage")
	} else {
		return fmt.Errorf("statut du service indéterminé")
	}
}

// stopWindowsService arrête le service Windows
func stopWindowsService() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	fmt.Printf("🛑 Arrêt du service %s...\n", SERVICE_NAME)

	cmd := fmt.Sprintf(`sc stop "%s"`, SERVICE_NAME)
	if err := runWindowsCommand(cmd); err != nil {
		fmt.Printf("⚠️  Attention : impossible d'arrêter le service : %v\n", err)
	}

	return nil
}

// uninstallWindowsService désinstalle complètement le service Windows
func uninstallWindowsService() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	fmt.Printf("🗑️  Désinstallation du service %s...\n", SERVICE_NAME)

	// Arrêter le service
	stopWindowsService()

	// Supprimer le service
	cmd := fmt.Sprintf(`sc delete "%s"`, SERVICE_NAME)
	if err := runWindowsCommand(cmd); err != nil {
		fmt.Printf("⚠️  Attention : impossible de supprimer le service : %v\n", err)
	}

	fmt.Println("✅ Service désinstallé")
	return nil
}

// runWindowsCommand exécute une commande Windows via cmd
func runWindowsCommand(command string) error {
	cmd := exec.Command("cmd", "/C", command)

	output, err := cmd.CombinedOutput()

	if err != nil {
		// Afficher la sortie en cas d'erreur pour diagnostiquer
		if len(output) > 0 {
			fmt.Printf("Erreur commande '%s': %s\n", command, string(output))
		}
		return fmt.Errorf("commande '%s' échouée : %w", command, err)
	}

	return nil
}
