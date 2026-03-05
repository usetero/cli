package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mattn/go-sqlite3"
)

// DB is a thin SQLite kernel wrapper.
type DB struct {
	db   *sql.DB
	path string
}

var (
	driverOnce    sync.Once
	extensionMu   sync.RWMutex
	extensionPath string
)

const driverName = "sqlite3_tero"

// SetExtensionPath configures the SQLite extension path loaded for new connections.
func SetExtensionPath(path string) {
	extensionMu.Lock()
	defer extensionMu.Unlock()
	extensionPath = path
}

func currentExtensionPath() string {
	extensionMu.RLock()
	defer extensionMu.RUnlock()
	return extensionPath
}

func registerDriver() {
	driverOnce.Do(func() {
		sql.Register(driverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if ext := currentExtensionPath(); ext != "" {
					if err := conn.LoadExtension(ext, "sqlite3_powersync_init"); err != nil {
						return err
					}
				}
				return nil
			},
		})
	})
}

// Open opens a SQLite database, applies core pragmas, and runs migrations.
func Open(ctx context.Context, path string) (*DB, error) {
	return open(ctx, path, true)
}

// OpenBare opens a SQLite database with core pragmas but without running app migrations.
// This is useful for extension-level tests that need a pristine database.
func OpenBare(ctx context.Context, path string) (*DB, error) {
	return open(ctx, path, false)
}

func open(ctx context.Context, path string, runMigrations bool) (*DB, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, wrapErr("mkdir", err)
	}

	registerDriver()

	raw, err := sql.Open(driverName, path)
	if err != nil {
		return nil, wrapErr("open", err)
	}
	if err := raw.PingContext(ctx); err != nil {
		raw.Close()
		return nil, wrapErr("ping", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 30000",
		"PRAGMA synchronous = NORMAL",
	} {
		if _, err := raw.ExecContext(ctx, pragma); err != nil {
			raw.Close()
			return nil, wrapErr("pragma", err)
		}
	}

	if runMigrations {
		if err := Migrate(ctx, raw); err != nil {
			raw.Close()
			return nil, err
		}
	}

	return &DB{db: raw, path: path}, nil
}

// Query executes a SQL query against the database.
func (d *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("database not open")
	}
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SQL query returning at most one row.
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a SQL statement.
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("database not open")
	}
	return d.db.ExecContext(ctx, query, args...)
}

// Path returns the opened SQLite file path.
func (d *DB) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// Raw returns the underlying *sql.DB.
func (d *DB) Raw() *sql.DB {
	if d == nil {
		return nil
	}
	return d.db
}

// Close closes the underlying DB.
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return wrapErr("close", d.db.Close())
}

// WithTx executes fn in a transaction.
func (d *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database not open")
	}
	return InTx(ctx, d.db, fn)
}
