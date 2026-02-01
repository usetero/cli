package config

import (
	"net/url"
	"os"
)

const (
	productionEndpoint          = "https://api.usetero.com/graphql"
	productionPowerSyncEndpoint = "https://powersync.usetero.com"
	productionChatEndpoint      = "https://api.usetero.com"
)

// CLIConfig holds configuration for the Tero CLI.
type CLIConfig struct {
	// APIEndpoint is the Tero control plane GraphQL endpoint
	APIEndpoint string

	// PowerSyncEndpoint is the PowerSync service endpoint for local-first sync
	PowerSyncEndpoint string

	// ChatEndpoint is the Chat API endpoint for message streaming
	ChatEndpoint string

	// WorkOSClientID is the WorkOS OAuth client ID for authentication
	WorkOSClientID string

	// Debug enables debug logging
	Debug bool
}

// LoadCLIConfig loads CLI configuration from environment variables and defaults.
func LoadCLIConfig() *CLIConfig {
	return &CLIConfig{
		APIEndpoint:       getEnvOrDefault("TERO_API_ENDPOINT", productionEndpoint),
		PowerSyncEndpoint: getEnvOrDefault("TERO_POWERSYNC_ENDPOINT", productionPowerSyncEndpoint),
		ChatEndpoint:      getEnvOrDefault("TERO_CHAT_ENDPOINT", productionChatEndpoint),
		WorkOSClientID:    getEnvOrDefault("TERO_WORKOS_CLIENT_ID", "client_01JQCC2D06JF9ASFA6GRHMFA3N"),
		Debug:             os.Getenv("TERO_DEBUG") == "true" || os.Getenv("TERO_DEBUG") == "1",
	}
}

// IsProduction returns true if the CLI is configured for production.
func (c *CLIConfig) IsProduction() bool {
	return c.APIEndpoint == productionEndpoint
}

// Namespace returns a namespace string for non-production environments.
// Returns empty string for production, host for other environments.
func (c *CLIConfig) Namespace() string {
	if c.IsProduction() {
		return ""
	}
	u, err := url.Parse(c.APIEndpoint)
	if err != nil {
		return ""
	}
	return u.Host
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
