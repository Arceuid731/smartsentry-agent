package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arceuid731/smartsentry-agent/internal/agent"  // Ajustez le chemin
	"github.com/Arceuid731/smartsentry-agent/internal/config" // Ajustez le chemin
)

func main() {
	log.Println("Initialisation de l'agent SmartSentry...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erreur de chargement de la configuration: %v", err)
	}

	smartSentryAgent, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Erreur de création de l'agent: %v", err)
	}

	// Gérer l'arrêt propre
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := smartSentryAgent.Start(); err != nil {
			log.Fatalf("Erreur de démarrage de l'agent: %v", err)
		}
	}()

	<-sigChan // Attendre un signal d'arrêt
	log.Println("Signal d'arrêt reçu.")
	smartSentryAgent.Stop()
	log.Println("Agent SmartSentry arrêté.")
}
