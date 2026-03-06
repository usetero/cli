package config

import "github.com/usetero/cli/internal/infrastructure/logging"

// LoggingConfig controls logger verbosity.
type LoggingConfig struct {
	Level logging.Level `name:"log-level" help:"Log level." env:"TERO_LOG_LEVEL"`
}
