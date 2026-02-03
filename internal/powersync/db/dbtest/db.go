// Package dbtest provides test utilities for the powersync/db package.
package dbtest

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

// OpenTestDB creates a temporary SQLite database with the PowerSync extension
// loaded and schema initialized. Ready for testing.
// The database is automatically closed when the test completes.
func OpenTestDB(t *testing.T) sqlite.DB {
	t.Helper()

	// Extension is registered via extension.init()
	ctx := context.Background()
	db := sqlitetest.OpenBareDB(t)

	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("ApplySchema() error = %v", err)
	}

	return db
}

// InsertCrudEntry inserts a test entry into the ps_crud table.
func InsertCrudEntry(t *testing.T, db sqlite.DB, id int64, txID *int64, data string) {
	t.Helper()

	ctx := context.Background()
	_, err := db.Exec(ctx, "INSERT INTO ps_crud (id, tx_id, data) VALUES (?, ?, ?)", id, txID, data)
	if err != nil {
		t.Fatalf("InsertCrudEntry() error = %v", err)
	}
}
