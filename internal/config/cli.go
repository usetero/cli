package config

import (
	"net/url"
	"os"
)

const (
	productionEndpoint          = "https://api.usetero.com"
	productionPowerSyncEndpoint = "https://powersync.usetero.com"
	productionChatEndpoint      = "https://chat.usetero.com"
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
		WorkOSClientID:    getEnvOrDefault("WORKOS_CLIENT_ID", "client_01JQCC2D06JF9ASFA6GRHMFA3N"),
		Debug:             os.Getenv("TERO_DEBUG") == "true" || os.Getenv("TERO_DEBUG") == "1",
	}
}

// Environment returns the environment name for data isolation.
// Resolution order:
//  1. TERO_ENV env var (explicit: "local", "dev", "prd", or any custom value)
//  2. Production API endpoint → "prd"
//  3. Non-production → derive from URL host (fallback for custom setups)
func (c *CLIConfig) Environment() string {
	if env := os.Getenv("TERO_ENV"); env != "" {
		return env
	}
	if c.APIEndpoint == productionEndpoint {
		return "prd"
	}
	u, err := url.Parse(c.APIEndpoint)
	if err != nil {
		return "prd"
	}
	return u.Host
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
