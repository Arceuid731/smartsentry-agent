//go:build windows

package main

import "fmt"

// installLinuxService stub pour Windows - la vraie implémentation est dans service_linux.go
func installLinuxService() error {
	return fmt.Errorf("installLinuxService n'est pas supporté sur Windows")
}
