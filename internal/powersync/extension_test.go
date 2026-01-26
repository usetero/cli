package powersync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/sqlite"
)

func TestExtensionPath(t *testing.T) {
	// Note: not parallel - shares cached extension path state
	t.Run("extracts extension to temp directory", func(t *testing.T) {

		// Act
		path, err := ExtensionPath()
		if err != nil {
			t.Fatalf("ExtensionPath() error = %v", err)
		}

		// Assert: file exists
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected extension file to exist at %s, got err: %v", path, err)
		}

		// Assert: file is not empty
		if info.Size() == 0 {
			t.Error("expected extension file to have content")
		}

		// Assert: file is executable
		perm := info.Mode().Perm()
		if perm&0100 == 0 {
			t.Errorf("expected extension to be executable, got permissions %o", perm)
		}
	})

	t.Run("returns same path on subsequent calls", func(t *testing.T) {

		// Act
		path1, err := ExtensionPath()
		if err != nil {
			t.Fatalf("first ExtensionPath() error = %v", err)
		}

		path2, err := ExtensionPath()
		if err != nil {
			t.Fatalf("second ExtensionPath() error = %v", err)
		}

		// Assert: same path returned (cached, not re-extracted)
		if path1 != path2 {
			t.Errorf("expected same path, got %s and %s", path1, path2)
		}
	})
}

func TestExtensionFilename(t *testing.T) {
	t.Parallel()

	t.Run("returns platform-specific filename", func(t *testing.T) {
		t.Parallel()

		// Act
		filename, err := extensionFilename()
		if err != nil {
			t.Fatalf("extensionFilename() error = %v", err)
		}

		// Assert: filename is not empty and has expected extension
		if filename == "" {
			t.Error("expected non-empty filename")
		}

		// Should end in .dylib (macOS) or .so (Linux)
		hasDylib := len(filename) > 6 && filename[len(filename)-6:] == ".dylib"
		hasSo := len(filename) > 3 && filename[len(filename)-3:] == ".so"
		if !hasDylib && !hasSo {
			t.Errorf("expected filename to end in .dylib or .so, got %s", filename)
		}
	})
}

func TestLoadExtension(t *testing.T) {
	t.Parallel()

	t.Run("loads into SQLite database", func(t *testing.T) {
		t.Parallel()

		// Arrange: get extension path
		extPath, err := ExtensionPath()
		if err != nil {
			t.Fatalf("ExtensionPath() error = %v", err)
		}

		// Arrange: open a database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := sqlite.Open(dbPath)
		if err != nil {
			t.Fatalf("sqlite.Open() error = %v", err)
		}
		defer db.Close()

		// Act
		err = db.LoadExtension(extPath, "sqlite3_powersync_init")

		// Assert
		if err != nil {
			t.Errorf("LoadExtension() error = %v", err)
		}
	})
}
