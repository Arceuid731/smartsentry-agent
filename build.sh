#!/bin/bash
# Script de compilation cross-platform pour SmartSentry Agent Installer

set -e

echo "🔨 Compilation SmartSentry Agent Installer"

# Nettoyer les anciens builds
rm -rf build/
mkdir -p build/

# Version (à récupérer depuis git tag ou version manuelle)
VERSION=${VERSION:-"v0.1.0"}

# Compiler pour Linux amd64
echo "📦 Compilation Linux amd64..."
(cd installer && GOOS=linux GOARCH=amd64 go build -ldflags "-X main.VERSION=$VERSION" -o ../build/smartsentry-installer-linux-amd64)

# Compiler pour Linux arm64
echo "📦 Compilation Linux arm64..."
(cd installer && GOOS=linux GOARCH=arm64 go build -ldflags "-X main.VERSION=$VERSION" -o ../build/smartsentry-installer-linux-arm64)

# Compiler pour Windows amd64
echo "📦 Compilation Windows amd64..."
(cd installer && GOOS=windows GOARCH=amd64 go build -ldflags "-X main.VERSION=$VERSION" -o ../build/smartsentry-installer-windows-amd64.exe)

# Compiler pour macOS amd64
echo "📦 Compilation macOS amd64..."
(cd installer && GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VERSION=$VERSION" -o ../build/smartsentry-installer-darwin-amd64)

# Compiler pour macOS arm64 (Apple Silicon)
echo "📦 Compilation macOS arm64..."
(cd installer && GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.VERSION=$VERSION" -o ../build/smartsentry-installer-darwin-arm64)

echo "✅ Compilation terminée ! Binaires disponibles dans ./build/"
ls -la build/
