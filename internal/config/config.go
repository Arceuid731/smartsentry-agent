package config

import (
	"errors"
	"os"
)

// AgentConfig contient la configuration de l'agent
type AgentConfig struct {
	ActivationToken string
	SaaSEndpoint    string
	AgentID         string // Sera rempli après l'activation
	AccessToken     string // Sera rempli après l'activation
}

// LoadConfig charge la configuration à partir des variables d'environnement ou des arguments
func LoadConfig() (*AgentConfig, error) {
	token := os.Getenv("SS_ACTIVATION_TOKEN")
	endpoint := os.Getenv("SS_SAAS_ENDPOINT")

	if token == "" {
		return nil, errors.New("SS_ACTIVATION_TOKEN variable d'environnement requise")
	}
	if endpoint == "" {
		return nil, errors.New("SS_SAAS_ENDPOINT variable d'environnement requise")
	}

	return &AgentConfig{
		ActivationToken: token,
		SaaSEndpoint:    endpoint,
	}, nil
}
