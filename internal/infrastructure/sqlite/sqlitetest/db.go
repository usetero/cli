package sqlitetest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Open opens a test database with app migrations applied.
func Open(t *testing.T) *sqlite.DB {
	t.Helper()
	sqlite.SetExtensionPath("")
	db, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
