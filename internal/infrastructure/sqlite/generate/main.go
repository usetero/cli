package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	controlPlaneRoot, err := findControlPlaneRoot(repoRoot)
	if err != nil {
		return err
	}

	authoredSchemaPath := filepath.Join(controlPlaneRoot, "internal", "infra", "powersync", "schema.sql")
	schemaSQL, err := os.ReadFile(authoredSchemaPath)
	if err != nil {
		return fmt.Errorf("read control-plane powersync schema %s: %w", authoredSchemaPath, err)
	}
	authoredJSONBSchemaPath := filepath.Join(controlPlaneRoot, "internal", "infra", "powersync", "jsonb_schema.json")
	jsonbSchema, err := os.ReadFile(authoredJSONBSchemaPath)
	if err != nil {
		return fmt.Errorf("read control-plane powersync jsonb schema %s: %w", authoredJSONBSchemaPath, err)
	}
	authoredPowerSyncSchemaPath := filepath.Join(controlPlaneRoot, "internal", "infra", "powersync", "powersync_schema.json")
	powerSyncSchema, err := os.ReadFile(authoredPowerSyncSchemaPath)
	if err != nil {
		return fmt.Errorf("read control-plane powersync schema artifact %s: %w", authoredPowerSyncSchemaPath, err)
	}

	outputDir := filepath.Join(repoRoot, "internal", "infrastructure", "sqlite")
	schemaPath := filepath.Join(outputDir, "schema.sql")
	if err := os.WriteFile(schemaPath, schemaSQL, 0o644); err != nil {
		return fmt.Errorf("write schema.sql: %w", err)
	}
	jsonbSchemaPath := filepath.Join(outputDir, "jsonb_schema.json")
	if err := os.WriteFile(jsonbSchemaPath, jsonbSchema, 0o644); err != nil {
		return fmt.Errorf("write jsonb_schema.json: %w", err)
	}
	powerSyncSchemaPath := filepath.Join(outputDir, "powersync_schema.json")
	if err := os.WriteFile(powerSyncSchemaPath, powerSyncSchema, 0o644); err != nil {
		return fmt.Errorf("write powersync_schema.json: %w", err)
	}

	if err := generateJSONBTypes(repoRoot, jsonbSchema); err != nil {
		return fmt.Errorf("generate jsonb types: %w", err)
	}

	if err := runSQLC(ctx, repoRoot); err != nil {
		return fmt.Errorf("sqlc generate: %w", err)
	}

	fmt.Printf("Wrote %s\n", schemaPath)
	fmt.Printf("Wrote %s\n", jsonbSchemaPath)
	fmt.Printf("Wrote %s\n", powerSyncSchemaPath)
	return nil
}

func runSQLC(ctx context.Context, repoRoot string) error {
	sqlcPath := filepath.Join(repoRoot, "bin", "sqlc")
	cmd := exec.CommandContext(ctx, sqlcPath, "generate")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "sqlc.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root")
		}
		dir = parent
	}
}

func findControlPlaneRoot(repoRoot string) (string, error) {
	if override := os.Getenv("TERO_CONTROL_PLANE_ROOT"); override != "" {
		return override, nil
	}

	controlPlaneRoot := filepath.Join(filepath.Dir(repoRoot), "control-plane")
	if _, err := os.Stat(filepath.Join(controlPlaneRoot, "go.mod")); err == nil {
		return controlPlaneRoot, nil
	}

	return "", fmt.Errorf("could not find control-plane repo; set TERO_CONTROL_PLANE_ROOT")
}
