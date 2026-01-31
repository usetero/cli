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
