package sqlite

import (
	"context"
	"database/sql"
)

// Migrate applies local runtime migrations.
//
// PowerSync-projected tables are created by the embedded extension schema at
// session start, not by replaying the reflected sqlc schema on every open.
func Migrate(ctx context.Context, db *sql.DB) error {
	_ = ctx
	_ = db
	return nil
}
