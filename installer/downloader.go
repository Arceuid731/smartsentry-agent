package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// downloadOTelCollector télécharge et installe le binaire OpenTelemetry Collector
// selon l'OS et l'architecture détectés
func downloadOTelCollector() error {
	// Construire l'URL de téléchargement basée sur l'OS et l'architecture
	downloadURL, filename := getOTelDownloadInfo()

	fmt.Printf("📡 Téléchargement depuis : %s\n", downloadURL)

	// Télécharger l'archive
	tempFile := filepath.Join(os.TempDir(), filename)
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("échec du téléchargement : %w", err)
	}
	defer os.Remove(tempFile) // Nettoyer le fichier temporaire

	fmt.Println("📦 Extraction de l'archive...")

	// Extraire le binaire selon le type d'archive
	var binaryPath string
	var err error

	if strings.HasSuffix(filename, ".zip") {
		binaryPath, err = extractFromZip(tempFile)
	} else {
		binaryPath, err = extractFromTarGz(tempFile)
	}

	if err != nil {
		return fmt.Errorf("échec de l'extraction : %w", err)
	}

	// Installer le binaire dans le répertoire système approprié
	return installBinary(binaryPath)
}

// getOTelDownloadInfo retourne l'URL de téléchargement et le nom de fichier
// pour la version et plateforme actuelles
func getOTelDownloadInfo() (string, string) {
	baseURL := "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download"

	var osName, archName, ext string

	// Mapping des noms d'OS Go vers les noms utilisés par OpenTelemetry
	switch runtime.GOOS {
	case "linux":
		osName = "linux"
		ext = "tar.gz"
	case "windows":
		osName = "windows"
		ext = "zip"
	case "darwin":
		osName = "darwin"
		ext = "tar.gz"
	default:
		osName = runtime.GOOS
		ext = "tar.gz"
	}

	// Mapping des architectures
	switch runtime.GOARCH {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	case "386":
		archName = "386"
	default:
		archName = runtime.GOARCH
	}

	// Construire le nom du fichier
	filename := fmt.Sprintf("otelcol-contrib_%s_%s_%s.%s", OTEL_VERSION, osName, archName, ext)

	// URL complète
	url := fmt.Sprintf("%s/v%s/%s", baseURL, OTEL_VERSION, filename)

	return url, filename
}

// downloadFile télécharge un fichier depuis une URL vers un chemin local
func downloadFile(url, filepath string) error {
	// Créer le fichier de destination
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Faire la requête HTTP
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Vérifier le code de statut
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mauvais code de statut : %d", resp.StatusCode)
	}

	// Copier le contenu
	_, err = io.Copy(out, resp.Body)
	return err
}

// extractFromZip extrait le binaire otelcol-contrib depuis une archive ZIP (Windows)
func extractFromZip(zipPath string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// Chercher le binaire principal
	for _, file := range reader.File {
		if strings.Contains(file.Name, "otelcol-contrib") && !strings.Contains(file.Name, "/") {
			// Extraire dans un répertoire temporaire
			extractPath := filepath.Join(os.TempDir(), "otelcol-contrib.exe")

			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			outFile, err := os.Create(extractPath)
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return "", err
			}

			return extractPath, nil
		}
	}

	return "", fmt.Errorf("binaire otelcol-contrib non trouvé dans l'archive")
}

// extractFromTarGz extrait le binaire depuis une archive tar.gz (Linux/macOS)
func extractFromTarGz(tarPath string) (string, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Décompression gzip
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzReader.Close()

	// Lecture tar
	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Chercher le binaire principal
		if strings.Contains(header.Name, "otelcol-contrib") && !strings.Contains(header.Name, "/") {
			// Extraire dans un répertoire temporaire
			extractPath := filepath.Join(os.TempDir(), "otelcol-contrib")

			outFile, err := os.Create(extractPath)
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return "", err
			}

			// Rendre exécutable (Linux/macOS)
			err = os.Chmod(extractPath, 0755)
			if err != nil {
				return "", err
			}

			return extractPath, nil
		}
	}

	return "", fmt.Errorf("binaire otelcol-contrib non trouvé dans l'archive")
}

// installBinary copie le binaire extrait vers son emplacement final dans le système
func installBinary(sourcePath string) error {
	var destPath string

	switch runtime.GOOS {
	case "windows":
		// Sur Windows, installer dans Program Files
		destPath = `C:\Program Files\SmartSentry\otelcol-contrib.exe`
		// Créer le répertoire s'il n'existe pas
		if err := os.MkdirAll(`C:\Program Files\SmartSentry`, 0755); err != nil {
			return err
		}
	default:
		// Sur Linux/macOS, installer dans /usr/local/bin
		destPath = "/usr/local/bin/otelcol-contrib"
	}

	fmt.Printf("📁 Installation du binaire vers : %s\n", destPath)

	// Copier le fichier
	return copyFile(sourcePath, destPath)
}

// copyFile copie un fichier depuis src vers dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Préserver les permissions sur Unix
	if runtime.GOOS != "windows" {
		sourceInfo, err := os.Stat(src)
		if err != nil {
			return err
		}
		return os.Chmod(dst, sourceInfo.Mode())
	}

	return nil
}
