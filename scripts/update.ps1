# Script de mise à jour SmartSentry Agent pour Windows
# Usage: Invoke-WebRequest -Uri "https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/update.ps1" -UseBasicParsing | Invoke-Expression

[CmdletBinding()]
param()

# Configuration
$RepoUrl = "https://github.com/Arceuid731/smartsentry-agent"
$BinaryName = "smartsentry-installer-windows-amd64.exe"
$ServiceName = "smartsentry-agent"
$ConfigDir = "C:\ProgramData\SmartSentry\Agent"
$BackupDir = "$ConfigDir\backups"

# Fonction d'affichage avec couleurs
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Vérifier les privilèges administrateur
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "Ce script doit être exécuté en tant qu'administrateur"
    exit 1
}

Write-Status "🔄 Mise à jour de SmartSentry Agent"

# Vérifier si l'agent est déjà installé
try {
    $service = Get-Service -Name $ServiceName -ErrorAction Stop
} catch {
    Write-Error "SmartSentry Agent n'est pas installé. Utilisez le script d'installation à la place."
    Write-Status "Installation: Invoke-WebRequest -Uri 'https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.ps1' -UseBasicParsing | Invoke-Expression"
    exit 1
}

# Créer le répertoire de sauvegarde
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

# Sauvegarder la configuration actuelle
Write-Status "💾 Sauvegarde de la configuration actuelle..."
$BackupDate = Get-Date -Format "yyyyMMdd_HHmmss"
$BackupFile = "$BackupDir\config_backup_$BackupDate.yaml"

if (Test-Path "$ConfigDir\config.yaml") {
    Copy-Item "$ConfigDir\config.yaml" $BackupFile
    Write-Success "Configuration sauvegardée dans $BackupFile"
} else {
    Write-Warning "Aucune configuration trouvée à sauvegarder"
}

# Arrêter le service
Write-Status "🛑 Arrêt du service SmartSentry Agent..."
if ($service.Status -eq "Running") {
    Stop-Service -Name $ServiceName -Force
    Write-Success "Service arrêté"
} else {
    Write-Warning "Le service n'était pas en cours d'exécution"
}

# Créer un répertoire temporaire
$TempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    Write-Status "📥 Téléchargement de la nouvelle version..."
    
    # URL de téléchargement
    $DownloadUrl = "$RepoUrl/releases/latest/download/$BinaryName"
    $InstallerPath = "$TempDir\installer.exe"
    
    # Télécharger l'installateur
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $InstallerPath -UseBasicParsing
    
    # Vérifier que le téléchargement a réussi
    if (-not (Test-Path $InstallerPath)) {
        Write-Error "Échec du téléchargement de l'installateur"
        exit 1
    }
    
    Write-Status "🔧 Lancement de la mise à jour..."
    
    # Exécuter l'installateur
    $process = Start-Process -FilePath $InstallerPath -Wait -PassThru
    
    if ($process.ExitCode -ne 0) {
        Write-Error "Échec de l'installation (code de sortie: $($process.ExitCode))"
        exit 1
    }
    
    Write-Status "🚀 Redémarrage du service..."
    Start-Service -Name $ServiceName
    
    # Vérifier que le service est bien démarré
    Start-Sleep -Seconds 2
    $service = Get-Service -Name $ServiceName
    
    if ($service.Status -eq "Running") {
        Write-Success "✅ Mise à jour terminée avec succès !"
        Write-Status "Le service SmartSentry Agent est maintenant actif avec la nouvelle version"
    } else {
        Write-Error "❌ Problème lors du redémarrage du service"
        Write-Status "Vérifiez les logs dans l'Event Viewer"
        exit 1
    }
    
    # Afficher les informations de debug
    Write-Status "📊 Informations de débogage :"
    Write-Host "  • Les logs sont maintenant disponibles dans : C:\ProgramData\SmartSentry\Agent\logs\"
    Write-Host "  • Logs du collecteur : C:\ProgramData\SmartSentry\Agent\logs\collector.log"
    Write-Host "  • Logs d'erreurs : C:\ProgramData\SmartSentry\Agent\logs\collector-error.log"
    
    # Afficher les commandes utiles
    Write-Host ""
    Write-Status "📋 Commandes utiles :"
    Write-Host "  • Statut      : Get-Service -Name 'smartsentry-agent'"
    Write-Host "  • Logs        : Event Viewer > Windows Logs > Application"
    Write-Host "  • Redémarrer  : Restart-Service -Name 'smartsentry-agent'"
    Write-Host "  • Arrêter     : Stop-Service -Name 'smartsentry-agent'"
    
    Write-Host ""
    Write-Status "📖 Documentation: $RepoUrl"
    Write-Status "💾 Sauvegarde disponible: $BackupFile"
    
} finally {
    # Nettoyer le répertoire temporaire
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force
    }
}