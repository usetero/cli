package sqlitetest

import (
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/sqlite"
)

// OpenTest creates a temporary SQLite database for testing.
// The database is automatically closed when the test completes.
func OpenTest(t *testing.T) *sqlite.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}
