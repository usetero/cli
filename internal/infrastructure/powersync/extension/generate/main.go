package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/usetero/cli/internal/infrastructure/powersync/client"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	token := strings.TrimSpace(os.Getenv("POWERSYNC_API_TOKEN"))
	if token == "" {
		return fmt.Errorf("POWERSYNC_API_TOKEN is required")
	}

	cfg, err := resolveRuntimeConfig()
	if err != nil {
		return err
	}

	schemaJSON, err := client.FetchSchemaJSON(context.Background(), cfg.PowerSync.Origin, client.AccessToken(token))
	if err != nil {
		return fmt.Errorf("fetch schema: %w", err)
	}

	outputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	schemaPath := filepath.Join(outputDir, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(schemaJSON), 0o644); err != nil {
		return fmt.Errorf("write schema.json: %w", err)
	}

	fmt.Printf("Wrote %s\n", schemaPath)
	return nil
}

func resolveRuntimeConfig() (config.RuntimeConfig, error) {
	env := config.EnvironmentProfile(strings.TrimSpace(os.Getenv("TERO_ENV")))
	if env == "" {
		env = config.ProfileLocal
	}
	return config.Resolve(config.RuntimeConfig{
		Env: env,
		PowerSync: config.PowerSyncConfig{
			Origin: strings.TrimSpace(os.Getenv("TERO_POWERSYNC_ORIGIN")),
		},
	})
}
