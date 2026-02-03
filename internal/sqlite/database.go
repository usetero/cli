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
	"sync"

	"github.com/mattn/go-sqlite3"
	"github.com/usetero/cli/internal/sqlite/gen"
)

var (
	// extensionPath is set by powersync.RegisterExtension() to enable
	// automatic extension loading on every new connection.
	extensionPath     string
	extensionPathOnce sync.Once
	driverRegistered  bool
)

// DB is the interface for application code.
// It provides type-safe access to domain data and raw query execution.
type DB interface {
	Messages() Messages
	Conversations() Conversations
	Subscribe() *Subscription
	Query(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) *sql.Row
	Exec(ctx context.Context, sql string, args ...any) (sql.Result, error)
	WithTx(ctx context.Context, fn func(tx *Tx) error) error
	Raw() *sql.DB
	Close() error
}

// database is the concrete implementation of DB.
type database struct {
	db    *sql.DB
	path  string
	watch watchState
}

// Ensure database implements DB.
var _ DB = (*database)(nil)

// SetExtensionPath configures the PowerSync extension to be loaded on every
// new database connection. This must be called before Open() to have effect.
// Typically called once at startup by the powersync package.
func SetExtensionPath(path string) {
	extensionPathOnce.Do(func() {
		extensionPath = path
		registerPowerSyncDriver()
	})
}

// registerPowerSyncDriver registers a custom SQLite driver that loads the
// PowerSync extension on every new connection.
func registerPowerSyncDriver() {
	if driverRegistered {
		return
	}
	sql.Register("sqlite3_powersync", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if extensionPath == "" {
				return nil
			}
			return conn.LoadExtension(extensionPath, "sqlite3_powersync_init")
		},
	})
	driverRegistered = true
}

// Open opens a SQLite database at the given path.
// The database file and parent directories are created if they don't exist.
// If SetExtensionPath() was called, the PowerSync extension is automatically
// loaded on every connection.
func Open(ctx context.Context, path string) (DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Use PowerSync driver if extension is configured, otherwise plain sqlite3
	driverName := "sqlite3"
	if extensionPath != "" {
		driverName = "sqlite3_powersync"
	}

	db, err := sql.Open(driverName, path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	d := &database{
		db:   db,
		path: path,
	}

	// Install update hooks for change notifications (only works with PowerSync extension)
	if extensionPath != "" {
		if err := d.installUpdateHooks(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("install update hooks: %w", err)
		}
	}

	return d, nil
}

// Close closes the database connection.
func (d *database) Close() error {
	return d.db.Close()
}

// Path returns the database file path.
func (d *database) Path() string {
	return d.path
}

// Raw returns the underlying *sql.DB for advanced use cases.
func (d *database) Raw() *sql.DB {
	return d.db
}

// Queries returns a Queries instance for running typed queries.
func (d *database) Queries() *gen.Queries {
	return gen.New(d.db)
}

// Messages returns type-safe message operations.
func (d *database) Messages() Messages {
	return &messagesImpl{queries: d.Queries()}
}

// Conversations returns type-safe conversation operations.
func (d *database) Conversations() Conversations {
	return &conversationsImpl{queries: d.Queries()}
}

// Query executes a query and returns the results.
func (d *database) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// BeginTx starts a transaction with the given options.
// Use opts.ReadOnly = true to prevent writes.
func (d *database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// QueryRow executes a query that returns at most one row.
func (d *database) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows.
func (d *database) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Count returns the number of rows in the given table.
func (d *database) Count(ctx context.Context, table string) (int64, error) {
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
func (d *database) LoadExtension(ctx context.Context, path, entryPoint string) error {
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

// Tx wraps a SQL transaction with convenience methods.
type Tx struct {
	tx *sql.Tx
}

// Exec executes a query that doesn't return rows within the transaction.
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row within the transaction.
func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Query executes a query and returns the results within the transaction.
func (t *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// WithTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
func (d *database) WithTx(ctx context.Context, fn func(tx *Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(&Tx{tx: tx}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %w (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
