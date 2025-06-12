# Script d'installation SmartSentry Agent pour Windows
# Usage: Invoke-WebRequest -Uri "https://github.com/Arceuid731/smartsentry-agent/raw/main/scripts/install.ps1" | Invoke-Expression

# Vérifier les privilèges administrateur
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))
{
    Write-Host "❌ Ce script doit être exécuté en tant qu'Administrateur" -ForegroundColor Red
    Write-Host "Clic droit sur PowerShell -> 'Exécuter en tant qu'administrateur'" -ForegroundColor Yellow
    exit 1
}

$RepoUrl = "https://github.com/Arceuid731/smartsentry-agent"
$BinaryName = "smartsentry-installer-windows-amd64.exe"
$DownloadUrl = "$RepoUrl/releases/latest/download/$BinaryName"

Write-Host "🚀 Installation de SmartSentry Agent" -ForegroundColor Blue

# Créer un répertoire temporaire
$TempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    Write-Host "📥 Téléchargement de l'installateur..." -ForegroundColor Blue
    
    # Télécharger l'installateur
    $InstallerPath = Join-Path $TempDir "installer.exe"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $InstallerPath
    
    if (-not (Test-Path $InstallerPath)) {
        Write-Host "❌ Échec du téléchargement de l'installateur" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "🔧 Lancement de l'installation..." -ForegroundColor Blue
    
    # Exécuter l'installateur
    & $InstallerPath
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Erreur lors de l'installation" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ Installation terminée !" -ForegroundColor Green
    Write-Host "Le service SmartSentry Agent est maintenant actif" -ForegroundColor Blue
    
    # Afficher les commandes utiles
    Write-Host ""
    Write-Host "📋 Commandes utiles :" -ForegroundColor Blue
    Write-Host "  • Statut      : sc query `"smartsentry-agent`""
    Write-Host "  • Logs        : Event Viewer > Windows Logs > Application"
    Write-Host "  • Redémarrer  : sc stop `"smartsentry-agent`" && sc start `"smartsentry-agent`""
    Write-Host "  • Arrêter     : sc stop `"smartsentry-agent`""
    
    Write-Host ""
    Write-Host "📖 Documentation: $RepoUrl" -ForegroundColor Blue
}
finally {
    # Nettoyer le répertoire temporaire
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
