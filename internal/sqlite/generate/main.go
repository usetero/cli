// Package main generates SQLite schema types from the PowerSync service.
//
// It fetches the database schema and sync rules from the PowerSync API,
// then generates Go types for each synced table.
//
// Usage:
//
//	doppler run -- go generate ./internal/sqlite
package main

import (
	"fmt"
	"os"
	"sort"
)

func main() {
	cfg := Config{
		URL:   os.Getenv("POWERSYNC_URL"),
		Token: os.Getenv("POWERSYNC_API_TOKEN"),
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Use: doppler run -- go generate ./internal/powersync")
		os.Exit(1)
	}

	client := NewClient(cfg)

	// Fetch schema (Postgres column types)
	schemaResp, err := client.FetchSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch schema: %v\n", err)
		os.Exit(1)
	}

	// Fetch sync rules (which tables/columns are synced)
	syncRulesResp, err := client.FetchSyncRules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch sync rules: %v\n", err)
		os.Exit(1)
	}

	// Build column type map from schema
	columnTypes := schemaResp.ColumnTypes()

	// Build synced tables from sync rules
	tables := syncRulesResp.Tables(columnTypes)

	// Sort tables by name for deterministic output
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	// Output to the current directory (go generate runs from the package dir)
	outputDir := mustGetwd()

	// Generate schema.go
	if err := WriteSchemaFile(outputDir, tables); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate schema.go: %v\n", err)
		os.Exit(1)
	}

	// Generate a file for each table type
	for _, table := range tables {
		if err := WriteTypeFile(outputDir, table); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate %s: %v\n", table.FileName, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Generated %d files in %s\n", len(tables)+1, outputDir)
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}
