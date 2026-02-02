package powersync

import (
	"context"
	"os"
	"testing"

	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestExtensionPath(t *testing.T) {
	t.Run("extracts extension to temp directory", func(t *testing.T) {
		path, err := ExtensionPath()
		if err != nil {
			t.Fatalf("ExtensionPath() error = %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected extension file to exist at %s, got err: %v", path, err)
		}

		if info.Size() == 0 {
			t.Error("expected extension file to have content")
		}

		perm := info.Mode().Perm()
		if perm&0100 == 0 {
			t.Errorf("expected extension to be executable, got permissions %o", perm)
		}
	})

	t.Run("returns same path on subsequent calls", func(t *testing.T) {
		path1, err := ExtensionPath()
		if err != nil {
			t.Fatalf("first ExtensionPath() error = %v", err)
		}

		path2, err := ExtensionPath()
		if err != nil {
			t.Fatalf("second ExtensionPath() error = %v", err)
		}

		if path1 != path2 {
			t.Errorf("expected same path, got %s and %s", path1, path2)
		}
	})
}

func TestExtensionFilename(t *testing.T) {
	t.Parallel()

	t.Run("returns platform-specific filename", func(t *testing.T) {
		t.Parallel()

		filename, err := extensionFilename()
		if err != nil {
			t.Fatalf("extensionFilename() error = %v", err)
		}

		if filename == "" {
			t.Error("expected non-empty filename")
		}

		hasDylib := len(filename) > 6 && filename[len(filename)-6:] == ".dylib"
		hasSo := len(filename) > 3 && filename[len(filename)-3:] == ".so"
		if !hasDylib && !hasSo {
			t.Errorf("expected filename to end in .dylib or .so, got %s", filename)
		}
	})
}

func TestRegisterExtension(t *testing.T) {
	t.Run("registers extension for automatic loading", func(t *testing.T) {
		// RegisterExtension should not error
		err := RegisterExtension()
		if err != nil {
			t.Fatalf("RegisterExtension() error = %v", err)
		}

		// After registration, Open() should automatically have extension loaded
		ctx := context.Background()
		db := sqlitetest.OpenTest(t)

		// Verify extension is loaded by calling a PowerSync function
		var result string
		row := db.QueryRow(ctx, "SELECT powersync_update_hooks('get')")
		if err := row.Scan(&result); err != nil {
			t.Fatalf("powersync function call failed (extension not loaded?): %v", err)
		}
	})
}

func TestApplySchema(t *testing.T) {
	t.Run("applies schema to database", func(t *testing.T) {
		// Ensure extension is registered
		if err := RegisterExtension(); err != nil {
			t.Fatalf("RegisterExtension() error = %v", err)
		}

		ctx := context.Background()
		db := sqlitetest.OpenTest(t)

		// Apply schema
		if err := ApplySchema(ctx, db); err != nil {
			t.Fatalf("ApplySchema() error = %v", err)
		}

		// Verify schema was applied by checking for a known table/view
		var count int
		row := db.QueryRow(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND sql LIKE '%powersync%'")
		if err := row.Scan(&count); err != nil {
			t.Fatalf("query sqlite_master error = %v", err)
		}

		if count == 0 {
			t.Error("expected PowerSync views to be created")
		}
	})
}

func TestUpdateHooks(t *testing.T) {
	t.Run("tracks table changes", func(t *testing.T) {
		if err := RegisterExtension(); err != nil {
			t.Fatalf("RegisterExtension() error = %v", err)
		}

		ctx := context.Background()
		db := sqlitetest.OpenTest(t)

		// Install update hooks
		_, err := db.Exec(ctx, "SELECT powersync_update_hooks('install')")
		if err != nil {
			t.Fatalf("install hooks error = %v", err)
		}

		// Create table and insert data
		_, err = db.Exec(ctx, "CREATE TABLE messages (id INTEGER PRIMARY KEY, content TEXT)")
		if err != nil {
			t.Fatalf("create table error = %v", err)
		}

		_, err = db.Exec(ctx, "INSERT INTO messages (content) VALUES ('hello')")
		if err != nil {
			t.Fatalf("insert error = %v", err)
		}

		// Get changed tables
		var result string
		row := db.QueryRow(ctx, "SELECT powersync_update_hooks('get')")
		if err := row.Scan(&result); err != nil {
			t.Fatalf("get hooks error = %v", err)
		}

		if result != `["messages"]` {
			t.Errorf("expected [\"messages\"], got %s", result)
		}

		// Second call should return empty (changes cleared)
		row = db.QueryRow(ctx, "SELECT powersync_update_hooks('get')")
		if err := row.Scan(&result); err != nil {
			t.Fatalf("second get hooks error = %v", err)
		}

		if result != `[]` {
			t.Errorf("expected [], got %s", result)
		}
	})
}
