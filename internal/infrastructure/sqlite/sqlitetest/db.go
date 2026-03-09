package sqlitetest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Open opens a test database with the embedded PowerSync schema applied.
func Open(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("register powersync extension: %v", err)
	}
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := extension.ApplySchema(context.Background(), db); err != nil {
		_ = db.Close()
		t.Fatalf("apply powersync schema: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
