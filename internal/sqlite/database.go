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
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Database is the interface for application code.
// It provides type-safe access to domain data.
type Database interface {
	Messages() Messages
	Conversations() Conversations
	Subscribe() *Subscription
	Close() error
	// DB returns the underlying *DB for infrastructure code that needs
	// low-level access (powersync, raw queries, etc).
	DB() *DB
}

// DB wraps a SQLite database connection.
// It implements the Database interface.
type DB struct {
	db    *sql.DB
	path  string
	watch watchState
}

// Ensure DB implements Database.
var _ Database = (*DB)(nil)

// Open opens a SQLite database at the given path.
// The database file and parent directories are created if they don't exist.
func Open(ctx context.Context, path string) (*DB, error) {
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
	if err := db.PingContext(ctx); err != nil {
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

// DB returns itself. This implements Database.DB() and allows
// infrastructure code to access low-level methods.
func (d *DB) DB() *DB {
	return d
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.path
}

// Raw returns the underlying *sql.DB for advanced use cases.
func (d *DB) Raw() *sql.DB {
	return d.db
}

// Queries returns a Queries instance for running typed queries.
func (d *DB) Queries() *gen.Queries {
	return gen.New(d.db)
}

// Messages returns type-safe message operations.
func (d *DB) Messages() Messages {
	return &messagesImpl{queries: d.Queries()}
}

// Conversations returns type-safe conversation operations.
func (d *DB) Conversations() Conversations {
	return &conversationsImpl{queries: d.Queries()}
}

// Query executes a query and returns the results.
func (d *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row.
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows.
// If update hooks are installed, subscribers are notified of any table changes.
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := d.db.ExecContext(ctx, query, args...)
	if err == nil {
		d.checkForChanges()
	}
	return result, err
}

// Count returns the number of rows in the given table.
func (d *DB) Count(ctx context.Context, table string) (int64, error) {
	var count int64
	// Use quote identifier to prevent SQL injection
	err := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", table)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return count, nil
}

// LoadExtension loads a SQLite extension from the given path.
// This uses the go-sqlite3 driver's C API to properly enable and load extensions.
// The entryPoint can be empty to use the default, or specify a custom entry point
// like "sqlite3_powersync_init" for the PowerSync extension.
func (d *DB) LoadExtension(ctx context.Context, path, entryPoint string) error {
	conn, err := d.db.Conn(ctx)
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
