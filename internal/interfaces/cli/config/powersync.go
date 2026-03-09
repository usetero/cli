package config

// PowerSyncConfig configures the PowerSync service origin.
type PowerSyncConfig struct {
	Origin string `name:"powersync-origin" help:"PowerSync API origin. Do not include a path." env:"TERO_POWERSYNC_ORIGIN"`
}
