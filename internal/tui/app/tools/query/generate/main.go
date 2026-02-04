// Generator for query tool schema.
//
// Reads the SQLite schema (source of truth for structure) and enriches it
// with column comments from the control plane's Postgres database.
//
// Usage:
//
//	go generate ./internal/tui/app/tools/query
//
// Requires TERO_CONTROL_PLANE_DATABASE_URL environment variable.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Get database URL from environment
	dbURL := os.Getenv("TERO_CONTROL_PLANE_DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("TERO_CONTROL_PLANE_DATABASE_URL not set")
	}

	// Find paths
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot := findRepoRoot(cwd)
	sqlitePath := filepath.Join(repoRoot, "internal/sqlite/schema.sql")
	outputPath := filepath.Join(cwd, "schema.sql")

	// Read SQLite schema (source of truth)
	schema, err := os.ReadFile(sqlitePath)
	if err != nil {
		return fmt.Errorf("read sqlite schema: %w", err)
	}

	// Get SQLite version
	sqliteVersion, err := getSQLiteVersion()
	if err != nil {
		return fmt.Errorf("get sqlite version: %w", err)
	}

	// Fetch comments from control plane database
	comments, err := fetchComments(dbURL)
	if err != nil {
		return fmt.Errorf("fetch comments: %w", err)
	}
	fmt.Printf("Fetched %d column comments from control plane\n", len(comments))

	// Annotate schema with comments
	annotated := annotateSchema(string(schema), comments, sqliteVersion)

	// Write output
	if err := os.WriteFile(outputPath, []byte(annotated), 0o644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}

	fmt.Printf("Wrote %s\n", outputPath)
	return nil
}

func findRepoRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repo root")
		}
		dir = parent
	}
}

func getSQLiteVersion() (string, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var version string
	err = db.QueryRowContext(context.Background(), "SELECT sqlite_version()").Scan(&version)
	return version, err
}
