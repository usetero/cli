// Package sqlite provides the local SQLite database for the CLI.
// This is the primary data layer - PowerSync keeps it in sync with the server,
// but you query it directly like any SQLite database.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-sqlite3"
)

// Database is the interface for the local SQLite database.
type Database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Count(table string) (int64, error)
	LoadExtension(path, entryPoint string) error
	Close() error
}

// DB wraps a SQLite database connection.
// It implements the Database interface.
type DB struct {
	db   *sql.DB
	path string
}

// Ensure DB implements Database.
var _ Database = (*DB)(nil)

// Open opens a SQLite database at the given path.
// The database file and parent directories are created if they don't exist.
func Open(path string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{
		db:   db,
		path: path,
	}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.path
}

// Raw returns the underlying *sql.DB for advanced use cases.
func (d *DB) Raw() *sql.DB {
	return d.db
}

// Query executes a query and returns the results.
func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns at most one row.
func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Exec executes a query that doesn't return rows.
func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Count returns the number of rows in the given table.
func (d *DB) Count(table string) (int64, error) {
	var count int64
	// Use quote identifier to prevent SQL injection
	err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", table)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return count, nil
}

// LoadExtension loads a SQLite extension from the given path.
// This uses the go-sqlite3 driver's C API to properly enable and load extensions.
// The entryPoint can be empty to use the default, or specify a custom entry point
// like "sqlite3_powersync_init" for the PowerSync extension.
func (d *DB) LoadExtension(path, entryPoint string) error {
	conn, err := d.db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()

	return conn.Raw(func(driverConn any) error {
		sqliteConn, ok := driverConn.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type: %T", driverConn)
		}
		return sqliteConn.LoadExtension(path, entryPoint)
	})
}
