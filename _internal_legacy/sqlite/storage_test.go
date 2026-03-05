package sqlite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/sqlite"
)

func TestStorageService_DatabasePath(t *testing.T) {
	t.Parallel()

	t.Run("returns path in databases directory", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-storage-path", "")
		storage := sqlite.NewStorageService(cfg)

		path, err := storage.DatabasePath("acc_123")
		if err != nil {
			t.Fatalf("DatabasePath() error = %v", err)
		}

		if !strings.Contains(path, "databases") {
			t.Errorf("path should contain 'databases' directory: %s", path)
		}
		if !strings.HasSuffix(path, "acc_123.sqlite") {
			t.Errorf("path should end with account ID and .sqlite extension: %s", path)
		}
	})

	t.Run("different accounts have different paths", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-storage-diff", "")
		storage := sqlite.NewStorageService(cfg)

		path1, _ := storage.DatabasePath("acc_1")
		path2, _ := storage.DatabasePath("acc_2")

		if path1 == path2 {
			t.Error("different accounts should have different paths")
		}
	})
}

func TestStorageService_ClearDatabase(t *testing.T) {
	t.Parallel()

	t.Run("removes existing database file", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-clear-db", "")
		storage := sqlite.NewStorageService(cfg)

		// Create the file first
		path, _ := storage.DatabasePath("acc_to_delete")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		// Verify it exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("file should exist before clear")
		}

		// Clear it
		err := storage.ClearDatabase("acc_to_delete")
		if err != nil {
			t.Fatalf("ClearDatabase() error = %v", err)
		}

		// Verify it's gone
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("file should not exist after clear")
		}

		// Cleanup
		t.Cleanup(func() {
			baseDir, _ := cfg.BaseDir()
			os.RemoveAll(baseDir)
		})
	})

	t.Run("succeeds when file does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-clear-nonexistent", "")
		storage := sqlite.NewStorageService(cfg)

		err := storage.ClearDatabase("nonexistent_account")

		if err != nil {
			t.Errorf("ClearDatabase() should not error for nonexistent file: %v", err)
		}
	})
}

func TestStorageService_Clear(t *testing.T) {
	t.Parallel()

	t.Run("removes all sqlite files", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-clear-all", "")
		storage := sqlite.NewStorageService(cfg)

		// Create multiple files
		path1, _ := storage.DatabasePath("acc_1")
		path2, _ := storage.DatabasePath("acc_2")
		if err := os.MkdirAll(filepath.Dir(path1), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path1, []byte("test1"), 0644); err != nil {
			t.Fatalf("write file 1: %v", err)
		}
		if err := os.WriteFile(path2, []byte("test2"), 0644); err != nil {
			t.Fatalf("write file 2: %v", err)
		}

		// Clear all
		err := storage.Clear()
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}

		// Verify both are gone
		if _, err := os.Stat(path1); !os.IsNotExist(err) {
			t.Error("file 1 should not exist after clear")
		}
		if _, err := os.Stat(path2); !os.IsNotExist(err) {
			t.Error("file 2 should not exist after clear")
		}

		// Cleanup
		t.Cleanup(func() {
			baseDir, _ := cfg.BaseDir()
			os.RemoveAll(baseDir)
		})
	})

	t.Run("succeeds when directory is empty", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.Load("test-clear-empty", "")
		storage := sqlite.NewStorageService(cfg)

		err := storage.Clear()

		if err != nil {
			t.Errorf("Clear() should not error for empty directory: %v", err)
		}
	})
}
