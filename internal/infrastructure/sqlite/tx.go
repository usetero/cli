package sqlite

import (
	"context"
	"database/sql"
)

// InTx executes fn in a SQL transaction, rolling back on error and committing on success.
func InTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return wrapErr("begin tx", err)
	}
	defer tx.Rollback() //nolint:errcheck // Best-effort rollback; commit path handles success.

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return wrapErr("commit tx", err)
	}
	return nil
}
