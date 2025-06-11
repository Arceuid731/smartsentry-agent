package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/Arceuid731/smartsentry-agent/internal/api"    // Ajustez le chemin
	"github.com/Arceuid731/smartsentry-agent/internal/config" // Ajustez le chemin
)

// Agent représente l'agent SmartSentry
type Agent struct {
	config    *config.AgentConfig
	apiClient *api.ClientAPI
	stopChan  chan struct{}
}

// New crée une nouvelle instance de l'agent
func New(cfg *config.AgentConfig) (*Agent, error) {
	apiClient := api.NewClientAPI(cfg)
	return &Agent{
		config:    cfg,
		apiClient: apiClient,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start démarre l'agent
func (a *Agent) Start() error {
	log.Println("Démarrage de l'agent SmartSentry...")

	activationResp, err := a.apiClient.ActivateAgent()
	if err != nil {
		return fmt.Errorf("échec de l'activation de l'agent: %w", err)
	}

	if !activationResp.Success {
		return fmt.Errorf("activation refusée par le serveur: %s", activationResp.Message)
	}

	log.Printf("Agent activé avec succès. AgentID: %s", activationResp.AgentID)
	a.config.AgentID = activationResp.AgentID         // Stocker l'ID de l'agent
	a.config.AccessToken = activationResp.AccessToken // Stocker le token d'accès

	// Boucle principale de l'agent (à développer)
	// Pour l'instant, juste une boucle qui logue toutes les 30 secondes
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Agent (ID: %s) en cours d'exécution...", a.config.AgentID)
			// Ici, vous ajouterez la collecte et l'envoi de métriques
		case <-a.stopChan:
			log.Println("Arrêt de l'agent...")
			return nil
		}
	}
}

// Stop arrête l'agent
func (a *Agent) Stop() {
	close(a.stopChan)
}
