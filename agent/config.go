package main

import (
	"fmt"
	"os"
	"runtime"
)

// Config holds the agent configuration
type Config struct {
	ServerURL         string
	AgentToken        string
	AgentSecretToken  string
	AgentVersion      string
	AgentArchitecture string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() (*Config, error) {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		return nil, fmt.Errorf("SERVER_URL environment variable is required")
	}

	agentToken := os.Getenv("AGENT_TOKEN")
	if agentToken == "" {
		return nil, fmt.Errorf("AGENT_TOKEN environment variable is required")
	}

	agentSecretToken := os.Getenv("AGENT_SECRET_TOKEN")
	if agentSecretToken == "" {
		// Fall back to AGENT_TOKEN if AGENT_SECRET_TOKEN is not set
		agentSecretToken = agentToken
	}

	return &Config{
		ServerURL:         serverURL,
		AgentToken:        agentToken,
		AgentSecretToken:  agentSecretToken,
		AgentVersion:      "1.0.0",
		AgentArchitecture: fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
	}, nil
}

