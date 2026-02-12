package config

import (
	"os"
)

// environmentDefaults holds the well-known defaults for a named environment.
type environmentDefaults struct {
	APIEndpoint       string
	PowerSyncEndpoint string
	ChatEndpoint      string
	WorkOSClientID    string
}

// environments maps TERO_ENV values to their default configuration.
// Unknown environments fall back to prd defaults.
var environments = map[string]environmentDefaults{
	"local": {
		APIEndpoint:       "http://localhost:8081",
		PowerSyncEndpoint: "http://localhost:8084",
		ChatEndpoint:      "http://localhost:8083",
		WorkOSClientID:    "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
	},
	"dev": {
		APIEndpoint:       "https://api.usetero.dev",
		PowerSyncEndpoint: "https://powersync.usetero.dev",
		ChatEndpoint:      "https://chat.usetero.dev",
		WorkOSClientID:    "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
	},
	"prd": {
		APIEndpoint:       "https://api.usetero.com",
		PowerSyncEndpoint: "https://powersync.usetero.com",
		ChatEndpoint:      "https://chat.usetero.com",
		WorkOSClientID:    "client_01JQCC2D06JF9ASFA6GRHMFA3N",
	},
}

// CLIConfig holds configuration for the Tero CLI.
type CLIConfig struct {
	// Env is the environment name (local, dev, prd) from TERO_ENV.
	Env string

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
// Resolution: TERO_ENV selects per-environment defaults, then individual env vars override.
func LoadCLIConfig() *CLIConfig {
	env := os.Getenv("TERO_ENV")
	if env == "" {
		env = "prd"
	}

	defaults := environments["prd"]
	if d, ok := environments[env]; ok {
		defaults = d
	}

	return &CLIConfig{
		Env:               env,
		APIEndpoint:       getEnvOrDefault("TERO_API_ENDPOINT", defaults.APIEndpoint),
		PowerSyncEndpoint: getEnvOrDefault("TERO_POWERSYNC_ENDPOINT", defaults.PowerSyncEndpoint),
		ChatEndpoint:      getEnvOrDefault("TERO_CHAT_ENDPOINT", defaults.ChatEndpoint),
		WorkOSClientID:    getEnvOrDefault("WORKOS_CLIENT_ID", defaults.WorkOSClientID),
		Debug:             os.Getenv("TERO_DEBUG") == "true" || os.Getenv("TERO_DEBUG") == "1",
	}
}

// Environment returns the environment name for data isolation.
func (c *CLIConfig) Environment() string {
	return c.Env
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
