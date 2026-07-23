package config

import (
	"os"
)

// Config holds agent runtime configuration.
type Config struct {
	ServerURL string
	Token     string
}

// Load reads agent configuration from environment variables.
func Load() (*Config, error) {
	serverURL := os.Getenv("AGENT_SERVER_URL")
	if serverURL == "" {
		serverURL = "ws://localhost:3000/agents"
	}

	return &Config{
		ServerURL: serverURL,
		Token:     os.Getenv("AGENT_TOKEN"),
	}, nil
}
