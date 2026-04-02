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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadCLIConfig()
	endpoint := cfg.PowerSyncEndpoint
	token := os.Getenv("POWERSYNC_API_TOKEN")
	outputDir := mustGetwd()

	var (
		schemaJSON string
		err        error
	)
	if token != "" {
		ctx := context.Background()
		fmt.Println("Fetching schema from PowerSync service...")
		schemaJSON, err = psapi.FetchSchemaJSON(ctx, endpoint, token)
		if err != nil {
			return fmt.Errorf("fetch schema: %w", err)
		}
	} else {
		fmt.Println("POWERSYNC_API_TOKEN not set; loading schema from local control-plane snapshot...")
		schemaJSON, err = loadSchemaJSONFromLocalSnapshot(outputDir)
		if err != nil {
			return fmt.Errorf("load local schema snapshot: %w", err)
		}
	}

	// Write schema.json to the extension directory (where it's embedded from)
	schemaPath := filepath.Join(outputDir, "extension", "schema.json")
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

func loadSchemaJSONFromLocalSnapshot(outputDir string) (string, error) {
	schemaPath, err := findLocalSchemaSnapshot(outputDir)
	if err != nil {
		return "", err
	}

	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", schemaPath, err)
	}

	tables, err := parseSchemaSQL(string(schemaSQL))
	if err != nil {
		return "", err
	}
	tables = psapi.ApplyClientIndexes(tables)

	data, err := json.Marshal(struct {
		Tables []psapi.SchemaTable `json:"tables"`
	}{Tables: tables})
	if err != nil {
		return "", fmt.Errorf("marshal schema: %w", err)
	}

	return string(data), nil
}

func findLocalSchemaSnapshot(outputDir string) (string, error) {
	candidates := []string{}
	if fromEnv := os.Getenv("POWERSYNC_SCHEMA_SQL"); fromEnv != "" {
		candidates = append(candidates, fromEnv)
	}
	if controlPlaneDir := os.Getenv("CONTROL_PLANE_DIR"); controlPlaneDir != "" {
		candidates = append(candidates, filepath.Join(controlPlaneDir, "internal", "infra", "powersync", "schema.sql"))
	}
	candidates = append(candidates,
		filepath.Join(filepath.Dir(outputDir), "control-plane", "internal", "infra", "powersync", "schema.sql"),
		filepath.Join(outputDir, "..", "..", "..", "..", "..", "..", "src", "tero", "control-plane", "internal", "infra", "powersync", "schema.sql"),
	)

	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not find control-plane PowerSync schema.sql; set POWERSYNC_API_TOKEN, POWERSYNC_SCHEMA_SQL, or CONTROL_PLANE_DIR")
}

var (
	createTablePattern = regexp.MustCompile(`(?ms)^CREATE TABLE (\w+) \(\n(.*?)\n\);`)
	columnPattern      = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s+(TEXT|INTEGER|REAL|BLOB)\b`)
)

func parseSchemaSQL(schemaSQL string) ([]psapi.SchemaTable, error) {
	matches := createTablePattern.FindAllStringSubmatch(schemaSQL, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no CREATE TABLE statements found")
	}

	tables := make([]psapi.SchemaTable, 0, len(matches))
	for _, match := range matches {
		table := psapi.SchemaTable{Name: match[1], Indexes: []psapi.SchemaIndex{}}
		for _, line := range strings.Split(match[2], "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			line = strings.TrimSuffix(line, ",")
			col := columnPattern.FindStringSubmatch(line)
			if len(col) == 0 {
				continue
			}
			table.Columns = append(table.Columns, psapi.SchemaColumn{
				Name: col[1],
				Type: strings.ToLower(col[2]),
			})
		}
		if len(table.Columns) == 0 {
			return nil, fmt.Errorf("table %s has no parsed columns", table.Name)
		}
		tables = append(tables, table)
	}

	return tables, nil
}
