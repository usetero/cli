package config

// WorkOSConfig configures WorkOS auth settings.
type WorkOSConfig struct {
	ClientID string `name:"workos-client-id" help:"WorkOS OAuth client ID." env:"TERO_WORKOS_CLIENT_ID"`
}
