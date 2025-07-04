package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// setupConfiguration télécharge et installe la configuration de l'agent
// selon l'OS détecté, puis demande à l'utilisateur l'adresse du Gateway
func setupConfiguration() error {
	// Déterminer le répertoire de configuration selon l'OS
	configDir, err := getConfigDirectory()
	if err != nil {
		return fmt.Errorf("impossible de déterminer le répertoire de configuration : %w", err)
	}

	// Créer le répertoire de configuration s'il n'existe pas
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire %s : %w", configDir, err)
	}

	fmt.Printf("📁 Configuration dans : %s\n", configDir)

	// Télécharger la configuration par défaut selon l'OS
	configURL := getDefaultConfigURL()
	configPath := filepath.Join(configDir, "config.yaml")

	fmt.Printf("📥 Téléchargement de la configuration depuis : %s\n", configURL)
	if err := downloadFile(configURL, configPath); err != nil {
		return fmt.Errorf("échec du téléchargement de la configuration : %w", err)
	}

	// Demander l'adresse du SmartSentry Gateway à l'utilisateur
	gatewayURL, err := promptForGatewayURL()
	if err != nil {
		return fmt.Errorf("erreur lors de la saisie du Gateway : %w", err)
	}

	// Mettre à jour la configuration avec l'URL du Gateway
	if err := updateConfigWithGateway(configPath, gatewayURL); err != nil {
		return fmt.Errorf("impossible de mettre à jour la configuration : %w", err)
	}

	fmt.Println("✅ Configuration mise à jour avec succès")
	return nil
}

// getConfigDirectory retourne le répertoire de configuration selon l'OS
func getConfigDirectory() (string, error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		return "/etc/smartsentry-agent", nil
	case "windows":
		// Sur Windows, utiliser ProgramData
		programData := os.Getenv("ProgramData")
		if programData == "" {
			return "", fmt.Errorf("variable d'environnement ProgramData non définie")
		}
		return filepath.Join(programData, "SmartSentry", "Agent"), nil
	default:
		return "", fmt.Errorf("système d'exploitation non supporté : %s", runtime.GOOS)
	}
}

// getDefaultConfigURL retourne l'URL de la configuration par défaut selon l'OS
func getDefaultConfigURL() string {
	switch runtime.GOOS {
	case "linux", "darwin":
		return CONFIG_BASE_URL + "/linux-default-config.yaml"
	case "windows":
		return CONFIG_BASE_URL + "/windows-default-config.yaml"
	default:
		// Fallback vers Linux
		return CONFIG_BASE_URL + "/linux-default-config.yaml"
	}
}

// promptForGatewayURL demande à l'utilisateur l'adresse de son SmartSentry Gateway
func promptForGatewayURL() (string, error) {
	fmt.Println("\n🔧 Configuration du SmartSentry Gateway")
	fmt.Println("Entrez l'adresse de votre SmartSentry Gateway (ex: http://192.168.1.211:4318)")
	fmt.Println("Cette adresse correspond à l'IP de votre cluster k3s avec le port NodePort du Gateway.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("URL du Gateway : ")

		// Lecture avec gestion d'erreur améliorée
		input, err := reader.ReadString('\n')
		if err != nil {
			// Gestion spécifique des erreurs de lecture
			if err.Error() == "unexpected newline" || strings.Contains(err.Error(), "newline") {
				fmt.Println("❌ Saisie invalide. Veuillez réessayer.")
				continue
			}
			return "", fmt.Errorf("erreur lors de la saisie du Gateway : %w", err)
		}

		gatewayURL := strings.TrimSpace(input)

		// Validation de l'URL vide
		if gatewayURL == "" {
			fmt.Println("❌ L'URL du Gateway ne peut pas être vide. Veuillez réessayer.")
			continue
		}

		// Ajout automatique du schéma HTTP si absent
		if !strings.HasPrefix(gatewayURL, "http://") && !strings.HasPrefix(gatewayURL, "https://") {
			gatewayURL = "http://" + gatewayURL
		}

		// Validation de l'URL avec le package net/url
		parsedURL, err := url.Parse(gatewayURL)
		if err != nil {
			fmt.Printf("❌ URL invalide : %v. Veuillez réessayer.\n", err)
			continue
		}

		// Vérification que l'URL contient un host valide
		if parsedURL.Host == "" {
			fmt.Println("❌ L'URL doit contenir un host valide. Veuillez réessayer.")
			continue
		}

		fmt.Printf("✅ Gateway configuré : %s\n", gatewayURL)
		return gatewayURL, nil
	}
}

// updateConfigWithGateway lit le fichier de config, remplace l'endpoint et le sauvegarde
func updateConfigWithGateway(configPath, gatewayURL string) error {
	// Lire le contenu du fichier
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("impossible de lire %s : %w", configPath, err)
	}

	// Remplacer le placeholder par l'URL réelle
	// Le placeholder dans les configs par défaut est : REMPLACE-PAR-IP-GATEWAY:30080
	configStr := string(content)

	// Plusieurs patterns possibles selon les configs
	replacements := map[string]string{
		"http://REMPLACE-PAR-IP-GATEWAY:30080": gatewayURL,
		"REMPLACE-PAR-IP-GATEWAY:30080":        strings.TrimPrefix(strings.TrimPrefix(gatewayURL, "http://"), "https://"),
		"IPDEVOTREVMK3S:30080":                 strings.TrimPrefix(strings.TrimPrefix(gatewayURL, "http://"), "https://"),
		"httpIPDEVOTREVMK3S30080":              gatewayURL,
	}

	// Appliquer les remplacements
	for old, new := range replacements {
		configStr = strings.ReplaceAll(configStr, old, new)
	}

	// Écrire le fichier mis à jour
	err = os.WriteFile(configPath, []byte(configStr), 0644)
	if err != nil {
		return fmt.Errorf("impossible d'écrire %s : %w", configPath, err)
	}

	return nil
}

// createSystemUser crée un utilisateur système dédié pour l'agent (Linux uniquement)
func createSystemUser() error {
	if runtime.GOOS != "linux" {
		// Sur Windows et macOS, on utilise le système de services par défaut
		return nil
	}

	fmt.Println("👤 Création de l'utilisateur système 'smartsentry'...")

	// Vérifier si l'utilisateur existe déjà
	if userExists("smartsentry") {
		fmt.Println("✅ L'utilisateur 'smartsentry' existe déjà")
		return nil
	}

	// Créer l'utilisateur système avec un shell non-interactif
	cmd := fmt.Sprintf("useradd --system --no-create-home --shell /usr/sbin/nologin smartsentry")
	if err := runSystemCommand("bash", "-c", cmd); err != nil {
		return fmt.Errorf("impossible de créer l'utilisateur système : %w", err)
	}

	fmt.Println("✅ Utilisateur 'smartsentry' créé avec succès")
	return nil
}

// userExists vérifie si un utilisateur système existe
func userExists(username string) bool {
	cmd := fmt.Sprintf("id %s", username)
	err := runSystemCommand("bash", "-c", cmd)
	return err == nil // Si la commande réussit, l'utilisateur existe
}

// createLogDirectory crée le répertoire de logs avec les bonnes permissions
func createLogDirectory() error {
	var logDir string

	switch runtime.GOOS {
	case "linux":
		logDir = "/var/log/smartsentry-agent"
	case "windows":
		programData := os.Getenv("ProgramData")
		if programData == "" {
			return fmt.Errorf("variable ProgramData non définie")
		}
		logDir = filepath.Join(programData, "SmartSentry", "Agent", "Logs")
	case "darwin":
		logDir = "/var/log/smartsentry-agent"
	default:
		return fmt.Errorf("OS non supporté pour les logs : %s", runtime.GOOS)
	}

	fmt.Printf("📝 Création du répertoire de logs : %s\n", logDir)

	// Créer le répertoire avec les bonnes permissions
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer %s : %w", logDir, err)
	}

	// Sur Linux, changer le propriétaire vers l'utilisateur smartsentry
	if runtime.GOOS == "linux" {
		cmd := fmt.Sprintf("chown smartsentry:smartsentry %s", logDir)
		if err := runSystemCommand("bash", "-c", cmd); err != nil {
			fmt.Printf("⚠️  Attention : impossible de changer le propriétaire de %s : %v\n", logDir, err)
		}
	}

	return nil
}

// runSystemCommand exécute une commande système et affiche l'erreur si échec
func runSystemCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			fmt.Printf("Erreur commande '%s %s': %s\n", name, strings.Join(args, " "), string(output))
		}
		return fmt.Errorf("commande '%s %s' échouée : %w", name, strings.Join(args, " "), err)
	}
	return nil
}
