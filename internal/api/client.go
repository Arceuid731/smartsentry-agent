package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Arceuid731/smartsentry-agent/internal/config" // Ajustez le chemin
)

// ActivationRequest est la structure pour la requête d'activation
type ActivationRequest struct {
	Token      string `json:"token"`
	InstanceID string `json:"instance_id"` // Un ID unique pour cette instance d'agent
	Hostname   string `json:"hostname"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Version    string `json:"version"` // Version de l'agent
}

// ActivationResponse est la structure pour la réponse d'activation
type ActivationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AgentID      string `json:"agent_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// ClientAPI gère la communication avec l'API SaaS de SmartSentry
type ClientAPI struct {
	httpClient *http.Client
	config     *config.AgentConfig
}

// NewClientAPI crée un nouveau client API
func NewClientAPI(cfg *config.AgentConfig) *ClientAPI {
	return &ClientAPI{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		config:     cfg,
	}
}

// ActivateAgent tente d'activer l'agent auprès du backend SaaS
func (c *ClientAPI) ActivateAgent() (*ActivationResponse, error) {
	hostname, _ := os.Hostname()
	instanceID, _ := os.Hostname() // Pourrait être plus unique, e.g., machine-id

	payload := ActivationRequest{
		Token:      c.config.ActivationToken,
		InstanceID: instanceID, // Utiliser un ID plus robuste en production
		Hostname:   hostname,
		OS:         "linux", // À rendre dynamique plus tard
		Arch:       "amd64", // À rendre dynamique plus tard
		Version:    "0.1.0", // Version de l'agent
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erreur de marshalling JSON: %w", err)
	}

	reqURL := fmt.Sprintf("%s/api/v1/agents/activate", c.config.SaaSEndpoint)
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erreur de création de la requête: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur d'envoi de la requête d'activation: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur de lecture de la réponse: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("échec de l'activation, statut: %d, réponse: %s", resp.StatusCode, string(body))
	}

	var activationResp ActivationResponse
	if err := json.Unmarshal(body, &activationResp); err != nil {
		return nil, fmt.Errorf("erreur de unmarshalling de la réponse JSON: %w", err)
	}

	return &activationResp, nil
}
