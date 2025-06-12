//go:build darwin

package main

import "fmt"

// installLinuxService stub pour macOS - la vraie implémentation est dans service_linux.go
func installLinuxService() error {
	return fmt.Errorf("installLinuxService n'est pas supporté sur macOS")
}

// installWindowsService stub pour macOS - la vraie implémentation est dans service_windows.go
func installWindowsService() error {
	return fmt.Errorf("installWindowsService n'est pas supporté sur macOS")
}
