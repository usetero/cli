package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDB_Count(t *testing.T) {
	t.Parallel()

	t.Run("returns zero for empty table", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		_, err = db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatalf("CREATE TABLE error = %v", err)
		}

		count, err := db.Count("items")
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}

		if count != 0 {
			t.Errorf("Count() = %d, want 0", count)
		}
	})

	t.Run("returns correct count", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		_, err = db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("CREATE TABLE error = %v", err)
		}

		_, err = db.Exec("INSERT INTO items (name) VALUES ('a'), ('b'), ('c')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		count, err := db.Count("items")
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}

		if count != 3 {
			t.Errorf("Count() = %d, want 3", count)
		}
	})

	t.Run("returns error for nonexistent table", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		_, err = db.Count("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent table, got nil")
		}
	})
}

func TestDB_Path(t *testing.T) {
	t.Parallel()

	t.Run("returns the database path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "mydb.sqlite")
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		if db.Path() != dbPath {
			t.Errorf("Path() = %q, want %q", db.Path(), dbPath)
		}
	})
}

func TestDB_Queries(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil Queries", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		q := db.Queries()
		if q == nil {
			t.Error("Queries() returned nil")
		}
	})
}

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("creates parent directories if they do not exist", func(t *testing.T) {
		t.Parallel()

		// Arrange: create a temp dir and a nested path that doesn't exist
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "nested", "dirs", "test.sqlite")

		// Verify the parent doesn't exist yet
		parentDir := filepath.Dir(dbPath)
		if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
			t.Fatalf("expected parent dir to not exist, got err: %v", err)
		}

		// Act
		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		// Assert: parent directory was created
		info, err := os.Stat(parentDir)
		if err != nil {
			t.Fatalf("expected parent dir to exist, got err: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected parent to be a directory")
		}

		// Assert: directory has correct permissions (0700)
		perm := info.Mode().Perm()
		if perm != 0700 {
			t.Errorf("expected permissions 0700, got %o", perm)
		}
	})

	t.Run("opens existing database", func(t *testing.T) {
		t.Parallel()

		// Arrange: create a database first
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "existing.sqlite")

		db1, err := Open(dbPath)
		if err != nil {
			t.Fatalf("first Open() error = %v", err)
		}

		// Create a table to verify it's a real database
		_, err = db1.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatalf("CREATE TABLE error = %v", err)
		}
		db1.Close()

		// Act: open the same database again
		db2, err := Open(dbPath)
		if err != nil {
			t.Fatalf("second Open() error = %v", err)
		}
		defer db2.Close()

		// Assert: table exists
		var count int64
		err = db2.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		if err != nil {
			t.Errorf("expected table to exist, got err: %v", err)
		}
	})
}
