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
	driversOnce       sync.Once
)

const (
	writeDriverName = "sqlite3_tero_write"
	readDriverName  = "sqlite3_tero_read"
)

// DB is the interface for application code.
// It provides type-safe access to domain data and raw query execution.
type DB interface {
	// Domain entities
	Conversations() Conversations
	LogEventPolicies() LogEventPolicies
	LogEvents() LogEvents
	Messages() Messages
	Services() Services

	// Aggregated statuses
	DatadogAccountStatuses() DatadogAccountStatuses
	ServiceStatuses() ServiceStatuses

	// Sync
	PendingUploadCounts(ctx context.Context) (map[Table]int64, error)
	Subscribe() *Subscription

	// Low-level
	Query(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) *sql.Row
	Exec(ctx context.Context, sql string, args ...any) (sql.Result, error)
	WithTx(ctx context.Context, fn func(tx *Tx) error) error
	Raw() *sql.DB     // Write pool — for PowerSync controller and direct writes
	ReadRaw() *sql.DB // Read pool — for query tool and direct reads
	Close() error
}

// database is the concrete implementation of DB.
type database struct {
	db     *sql.DB // write pool — WAL writer, PowerSync controller, mutations
	readDB *sql.DB // read pool — query_only, all domain reads
	path   string
	watch  watchState
}

// Ensure database implements DB.
var _ DB = (*database)(nil)

// SetExtensionPath configures the PowerSync extension to be loaded on every
// new database connection. This must be called before Open() to have effect.
// Typically called once at startup by the powersync package.
func SetExtensionPath(path string) {
	extensionPathOnce.Do(func() {
		extensionPath = path
	})
}

// Per-connection pragmas applied to every connection via driver hooks.
// These must be set per-connection because database/sql pools create
// connections on demand, and pragmas are connection-scoped.
var basePragmas = []string{
	"PRAGMA busy_timeout = 30000",      // Wait up to 30s for locks instead of failing immediately
	"PRAGMA synchronous = NORMAL",      // Safe with WAL, avoids fsync on every commit
	"PRAGMA cache_size = -51200",       // 50MB page cache (negative = KB)
	"PRAGMA temp_store = MEMORY",       // Keep temp tables in memory
	"PRAGMA recursive_triggers = TRUE", // Required by PowerSync extension for trigger chains
}

// registerDrivers registers the write and read-only SQLite drivers exactly once.
// Both drivers load the PowerSync extension (if configured) and apply base pragmas
// on every new connection. The read driver additionally sets query_only = ON.
func registerDrivers() {
	driversOnce.Do(func() {
		sql.Register(writeDriverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if extensionPath != "" {
					if err := conn.LoadExtension(extensionPath, "sqlite3_powersync_init"); err != nil {
						return err
					}
				}
				return execPragmas(conn, basePragmas)
			},
		})

		readPragmas := append(basePragmas, "PRAGMA query_only = ON")
		sql.Register(readDriverName, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if extensionPath != "" {
					if err := conn.LoadExtension(extensionPath, "sqlite3_powersync_init"); err != nil {
						return err
					}
				}
				return execPragmas(conn, readPragmas)
			},
		})
	})
}

// execPragmas runs a list of PRAGMA statements on a raw SQLite connection.
func execPragmas(conn *sqlite3.SQLiteConn, pragmas []string) error {
	for _, p := range pragmas {
		if _, err := conn.Exec(p, nil); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}

// Open opens a SQLite database at the given path with separate read and write
// connection pools. The write pool enables WAL mode for concurrent read access
// during writes. The read pool enforces query_only = ON via the driver hook.
//
// The database file and parent directories are created if they don't exist.
func Open(ctx context.Context, path string) (DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	registerDrivers()

	// Open write pool
	db, err := sql.Open(writeDriverName, path)
	if err != nil {
		return nil, fmt.Errorf("open write pool: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping write pool: %w", err)
	}
	// SQLite only allows one writer at a time. A single connection avoids
	// SQLITE_BUSY between writers and ensures powersync_update_hooks state
	// (which is per-connection) is shared across all write operations.
	db.SetMaxOpenConns(1)

	// Database-level pragmas (not per-connection — one exec is correct).
	// WAL allows concurrent readers during writes — the core fix for query blocking during sync.
	// journal_size_limit caps the WAL file at 6MB to prevent unbounded growth.
	for _, p := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA journal_size_limit = 6291456",
	} {
		if _, err := db.ExecContext(ctx, p); err != nil {
			db.Close()
			return nil, fmt.Errorf("%s: %w", p, err)
		}
	}

	// Open read pool — every connection has query_only = ON and busy_timeout
	// set automatically by the driver hook.
	readDB, err := sql.Open(readDriverName, path)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open read pool: %w", err)
	}
	if err := readDB.PingContext(ctx); err != nil {
		db.Close()
		readDB.Close()
		return nil, fmt.Errorf("ping read pool: %w", err)
	}

	d := &database{
		db:     db,
		readDB: readDB,
		path:   path,
	}

	// Install update hooks for change notifications (only works with PowerSync extension)
	if extensionPath != "" {
		if err := d.installUpdateHooks(ctx); err != nil {
			db.Close()
			readDB.Close()
			return nil, fmt.Errorf("install update hooks: %w", err)
		}
	}

	return d, nil
}

// Close closes both the read and write connection pools.
func (d *database) Close() error {
	readErr := d.readDB.Close()
	writeErr := d.db.Close()
	if writeErr != nil {
		return writeErr
	}
	return readErr
}

// Path returns the database file path.
func (d *database) Path() string {
	return d.path
}

// ---------------------------------------------------------------------------
// Raw pool access
// ---------------------------------------------------------------------------

// Raw returns the write pool for direct access.
// Used by the PowerSync controller which needs write access.
func (d *database) Raw() *sql.DB {
	return d.db
}

// ReadRaw returns the read pool for direct access.
// Used by the query tool for user-initiated SQL queries.
func (d *database) ReadRaw() *sql.DB {
	return d.readDB
}

// ---------------------------------------------------------------------------
// Typed query access (sqlc generated)
// ---------------------------------------------------------------------------

// ReadQueries returns a Queries instance backed by the read pool.
func (d *database) ReadQueries() *gen.Queries {
	return gen.New(d.readDB)
}

// WriteQueries returns a Queries instance backed by the write pool.
// Writes kick the watcher so subscribers are notified immediately.
func (d *database) WriteQueries() *gen.Queries {
	return gen.New(&kickingDB{db: d.db, kick: d.kickWatcher})
}

// ---------------------------------------------------------------------------
// Domain entity factories
// ---------------------------------------------------------------------------

// Messages returns type-safe message operations.
// Uses both pools: reads from readDB, writes (create/update) from writeDB.
func (d *database) Messages() Messages {
	return &messagesImpl{read: d.ReadQueries(), write: d.WriteQueries()}
}

// Conversations returns type-safe conversation operations.
// Uses both pools: reads from readDB, writes (create/update) from writeDB.
func (d *database) Conversations() Conversations {
	return &conversationsImpl{read: d.ReadQueries(), write: d.WriteQueries()}
}

// DatadogAccountStatuses returns type-safe Datadog account status operations.
func (d *database) DatadogAccountStatuses() DatadogAccountStatuses {
	return &datadogAccountStatusesImpl{queries: d.ReadQueries()}
}

// ServiceStatuses returns type-safe service status operations.
func (d *database) ServiceStatuses() ServiceStatuses {
	return &serviceStatusesImpl{queries: d.ReadQueries()}
}

// Services returns type-safe service operations.
func (d *database) Services() Services {
	return &servicesImpl{queries: d.ReadQueries()}
}

// LogEvents returns type-safe log event operations.
func (d *database) LogEvents() LogEvents {
	return &logEventsImpl{queries: d.ReadQueries()}
}

// LogEventPolicies returns type-safe log event policy operations.
func (d *database) LogEventPolicies() LogEventPolicies {
	return &logEventPoliciesImpl{queries: d.ReadQueries()}
}

// ---------------------------------------------------------------------------
// Low-level query methods
// ---------------------------------------------------------------------------

// Query executes a read query via the read pool.
func (d *database) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.readDB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row via the read pool.
func (d *database) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.readDB.QueryRowContext(ctx, query, args...)
}

// Exec executes a statement via the write pool and kicks the watcher so
// subscribers are notified without waiting for the next poll tick.
func (d *database) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := d.db.ExecContext(ctx, query, args...)
	if err == nil {
		d.kickWatcher()
	}
	return result, err
}

// BeginTx starts a transaction on the write pool.
func (d *database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Count returns the number of rows in the given table via the read pool.
func (d *database) Count(ctx context.Context, table string) (int64, error) {
	var count int64
	// Use quote identifier to prevent SQL injection
	err := d.readDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", table)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return count, nil
}

// PendingUploadCounts returns pending upload counts grouped by entity table via the read pool.
func (d *database) PendingUploadCounts(ctx context.Context) (map[Table]int64, error) {
	rows, err := d.readDB.QueryContext(ctx,
		"SELECT json_extract(data, '$.type') AS entity, COUNT(*) AS cnt FROM ps_crud GROUP BY 1")
	if err != nil {
		return nil, fmt.Errorf("count pending uploads: %w", err)
	}
	defer rows.Close()

	counts := make(map[Table]int64)
	for rows.Next() {
		var entity string
		var count int64
		if err := rows.Scan(&entity, &count); err != nil {
			return nil, fmt.Errorf("scan pending upload count: %w", err)
		}
		counts[Table(entity)] = count
	}
	return counts, rows.Err()
}

// LoadExtension loads a SQLite extension on the write pool.
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

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

// Tx wraps a SQL transaction with convenience methods.
type Tx struct {
	tx *sql.Tx
}

// Exec executes a statement within the transaction.
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row within the transaction.
func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Query executes a query within the transaction.
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

// WithTx executes a function within a database transaction on the write pool.
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
