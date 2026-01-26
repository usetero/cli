// Package main generates the PowerSync schema from the PowerSync service.
//
// It fetches the schema from the admin API and writes schema.json for embedding.
//
// Usage:
//
//	doppler run -- go generate ./internal/powersync
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/usetero/cli/internal/powersync"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	endpoint := os.Getenv("TERO_POWERSYNC_ENDPOINT")
	token := os.Getenv("POWERSYNC_API_TOKEN")

	if endpoint == "" {
		return fmt.Errorf("TERO_POWERSYNC_ENDPOINT is required\nUse: doppler run -- go generate ./internal/powersync")
	}
	if token == "" {
		return fmt.Errorf("POWERSYNC_API_TOKEN is required\nUse: doppler run -- go generate ./internal/powersync")
	}

	ctx := context.Background()

	fmt.Println("Fetching schema from PowerSync service...")

	schemaJSON, err := powersync.FetchSchemaJSON(ctx, endpoint, token)
	if err != nil {
		return fmt.Errorf("fetch schema: %w", err)
	}

	// Write schema.json to the powersync package directory (where go generate runs from)
	outputDir := mustGetwd()
	schemaPath := filepath.Join(outputDir, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(schemaJSON), 0o644); err != nil {
		return fmt.Errorf("write schema.json: %w", err)
	}

	fmt.Printf("Wrote %s\n", schemaPath)
	return nil
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}
