package config

import (
	"os"
	"strings"
)

// CLIConfig holds configuration for the Tero CLI.
type CLIConfig struct {
	// APIEndpoint is the Tero control plane GraphQL endpoint
	APIEndpoint string

	// WorkOSClientID is the WorkOS OAuth client ID for authentication
	WorkOSClientID string

	// Debug enables debug logging
	Debug bool
}

// LoadCLIConfig loads CLI configuration from environment variables and defaults.
// Priority: environment variables > smart defaults based on version.
func LoadCLIConfig(version string) *CLIConfig {
	cfg := &CLIConfig{
		APIEndpoint:    getDefaultAPIEndpoint(version),
		WorkOSClientID: getDefaultWorkOSClientID(),
		Debug:          false,
	}

	// Override from environment variables if set
	if endpoint := os.Getenv("TERO_API_ENDPOINT"); endpoint != "" {
		cfg.APIEndpoint = endpoint
	}

	if clientID := os.Getenv("TERO_WORKOS_CLIENT_ID"); clientID != "" {
		cfg.WorkOSClientID = clientID
	}

	if debug := os.Getenv("TERO_DEBUG"); debug == "true" || debug == "1" {
		cfg.Debug = true
	}

	return cfg
}

// getDefaultAPIEndpoint returns the default API endpoint based on the build version.
// Development builds (version contains "dev" or "dirty") use localhost.
// Release builds use the production API.
func getDefaultAPIEndpoint(version string) string {
	// Development: local control plane
	if isDevelopmentBuild(version) {
		return "http://localhost:8081/graphql"
	}

	// Production: hosted control plane
	return "https://api.usetero.com/graphql"
}

// getDefaultWorkOSClientID returns the default WorkOS client ID.
// Production client ID is used by default.
// Set TERO_WORKOS_CLIENT_ID environment variable to override (e.g., for staging/dev).
func getDefaultWorkOSClientID() string {
	return "client_01JQCC2D06JF9ASFA6GRHMFA3N" // Production
}

// isDevelopmentBuild returns true if this is a development build.
func isDevelopmentBuild(version string) bool {
	if version == "" || version == "dev" {
		return true
	}

	// Git describe with uncommitted changes adds "-dirty"
	if strings.Contains(version, "dirty") {
		return true
	}

	// No version tag set during build
	if strings.HasPrefix(version, "v0.0.0") {
		return true
	}

	return false
}
