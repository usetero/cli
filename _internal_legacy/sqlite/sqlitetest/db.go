package sqlitetest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/sqlite"
)

// OpenBareDB creates a temporary SQLite database WITHOUT the PowerSync schema.
// Use this only for low-level tests (extension loading, watch hooks with custom tables).
// For most tests, use dbtest.OpenTestDB() which includes the full schema.
func OpenBareDB(t *testing.T) sqlite.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := sqlite.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}
