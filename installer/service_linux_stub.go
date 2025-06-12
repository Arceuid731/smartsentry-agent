//go:build linux

package main

import "fmt"

// installWindowsService stub pour Linux - la vraie implémentation est dans service_windows.go
func installWindowsService() error {
	return fmt.Errorf("installWindowsService n'est pas supporté sur Linux")
}
