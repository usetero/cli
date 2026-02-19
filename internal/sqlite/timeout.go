package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// defaultQueryTimeout is applied to any query whose context has no deadline.
// Callers can override by setting their own deadline on the context.
const defaultQueryTimeout = 3 * time.Second

// Ensure timeoutDB implements gen.DBTX.
var _ gen.DBTX = (*timeoutDB)(nil)

// timeoutDB wraps a *sql.DB and applies defaultQueryTimeout to any context
// that doesn't already have a deadline set by the caller.
type timeoutDB struct {
	db *sql.DB
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultQueryTimeout)
}

func (t *timeoutDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return t.db.ExecContext(ctx, query, args...)
}

func (t *timeoutDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return t.db.PrepareContext(ctx, query)
}

func (t *timeoutDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Not canceled here — the caller iterates the returned *sql.Rows after
	// this method returns. The timeout still fires after defaultQueryTimeout
	// if the query is stuck; it just isn't eagerly cleaned up.
	ctx, _ = withTimeout(ctx) //nolint:lostcancel
	return t.db.QueryContext(ctx, query, args...)
}

func (t *timeoutDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Not canceled here — the caller calls Scan() on the returned *sql.Row
	// after this method returns, and Scan checks ctx.Err(). Same as QueryContext.
	ctx, _ = withTimeout(ctx) //nolint:lostcancel
	return t.db.QueryRowContext(ctx, query, args...)
}
