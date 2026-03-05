package config

// PowerSyncConfig configures the PowerSync endpoint.
type PowerSyncConfig struct {
	URL string `name:"powersync-url" help:"PowerSync API URL." env:"TERO_POWERSYNC_URL"`
}
