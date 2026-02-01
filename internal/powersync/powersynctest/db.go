package powersynctest

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

// OpenTestDB creates a temporary SQLite database with the PowerSync extension loaded.
// The database is automatically closed when the test completes.
func OpenTestDB(t *testing.T) *sqlite.DB {
	t.Helper()

	ctx := context.Background()
	db := sqlitetest.OpenTest(t)

	extPath, err := powersync.ExtensionPath()
	if err != nil {
		t.Fatalf("ExtensionPath() error = %v", err)
	}

	if err := db.LoadExtension(ctx, extPath, "sqlite3_powersync_init"); err != nil {
		t.Fatalf("LoadExtension() error = %v", err)
	}

	return db
}

// OpenTestDBWithSchema creates a temporary SQLite database with the PowerSync extension
// loaded and schema initialized. Ready for sync operations.
func OpenTestDBWithSchema(t *testing.T) *sqlite.DB {
	t.Helper()

	ctx := context.Background()
	db := OpenTestDB(t)

	// Initialize with the embedded schema
	if _, err := db.Exec(ctx, "SELECT powersync_replace_schema(?)", powersync.SchemaJSON()); err != nil {
		t.Fatalf("powersync_replace_schema() error = %v", err)
	}

	return db
}

// InsertCrudEntry inserts a test entry into the ps_crud table.
func InsertCrudEntry(t *testing.T, db *sqlite.DB, id int64, txID *int64, data string) {
	t.Helper()

	ctx := context.Background()
	_, err := db.Exec(ctx, "INSERT INTO ps_crud (id, tx_id, data) VALUES (?, ?, ?)", id, txID, data)
	if err != nil {
		t.Fatalf("InsertCrudEntry() error = %v", err)
	}
}
