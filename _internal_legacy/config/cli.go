package config

import (
	"os"
)

// environmentDefaults holds the well-known defaults for a named environment.
type environmentDefaults struct {
	APIOrigin       string
	PowerSyncOrigin string
	ChatOrigin      string
	WorkOSClientID  string
}

// environments maps TERO_ENV values to their default configuration.
// Unknown environments fall back to prd defaults.
var environments = map[string]environmentDefaults{
	"local": {
		APIOrigin:       "http://localhost:18081",
		PowerSyncOrigin: "http://localhost:18084",
		ChatOrigin:      "http://localhost:18083",
		WorkOSClientID:  "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
	},
	"dev": {
		APIOrigin:       "https://api.usetero.dev",
		PowerSyncOrigin: "https://powersync.usetero.dev",
		ChatOrigin:      "https://chat.usetero.dev",
		WorkOSClientID:  "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
	},
	"prd": {
		APIOrigin:       "https://api.usetero.com",
		PowerSyncOrigin: "https://powersync.usetero.com",
		ChatOrigin:      "https://chat.usetero.com",
		WorkOSClientID:  "client_01JQCC2D06JF9ASFA6GRHMFA3N",
	},
}

// CLIConfig holds configuration for the Tero CLI.
type CLIConfig struct {
	// Env is the environment name (local, dev, prd) from TERO_ENV.
	Env string

	// APIOrigin is the Tero control plane service origin.
	APIOrigin string

	// PowerSyncOrigin is the PowerSync service origin for local-first sync.
	PowerSyncOrigin string

	// ChatOrigin is the Chat API origin for message streaming.
	ChatOrigin string

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
		Env:             env,
		APIOrigin:       getEnvOrDefault("TERO_API_ORIGIN", defaults.APIOrigin),
		PowerSyncOrigin: getEnvOrDefault("TERO_POWERSYNC_ORIGIN", defaults.PowerSyncOrigin),
		ChatOrigin:      getEnvOrDefault("TERO_CHAT_ORIGIN", defaults.ChatOrigin),
		WorkOSClientID:  getEnvOrDefault("WORKOS_CLIENT_ID", defaults.WorkOSClientID),
		Debug:           os.Getenv("TERO_DEBUG") == "true" || os.Getenv("TERO_DEBUG") == "1",
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
