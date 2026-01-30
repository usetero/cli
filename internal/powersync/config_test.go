package powersync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/powersync"
)

func TestConfig_DataDir(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	t.Run("production uses .tero/data", func(t *testing.T) {
		t.Parallel()

		config := &powersync.Config{
			Endpoint:  "https://powersync.usetero.dev",
			Namespace: "", // production
		}

		dir, err := config.DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}

		expected := filepath.Join(homeDir, ".tero", "data")
		if dir != expected {
			t.Errorf("DataDir() = %q, want %q", dir, expected)
		}
	})

	t.Run("dev namespace uses .tero/{namespace}/data", func(t *testing.T) {
		t.Parallel()

		config := &powersync.Config{
			Endpoint:  "https://powersync.dev.usetero.dev",
			Namespace: "localhost:3000",
		}

		dir, err := config.DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}

		expected := filepath.Join(homeDir, ".tero", "localhost:3000", "data")
		if dir != expected {
			t.Errorf("DataDir() = %q, want %q", dir, expected)
		}
	})
}

func TestConfig_ExtensionDir(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	t.Run("shared across all environments", func(t *testing.T) {
		t.Parallel()

		// Test with production config
		prodConfig := &powersync.Config{Namespace: ""}
		prodDir, err := prodConfig.ExtensionDir()
		if err != nil {
			t.Fatalf("ExtensionDir() error = %v", err)
		}

		// Test with dev config
		devConfig := &powersync.Config{Namespace: "localhost:3000"}
		devDir, err := devConfig.ExtensionDir()
		if err != nil {
			t.Fatalf("ExtensionDir() error = %v", err)
		}

		// Should be the same
		if prodDir != devDir {
			t.Errorf("ExtensionDir should be shared: prod=%q, dev=%q", prodDir, devDir)
		}

		expected := filepath.Join(homeDir, ".tero", "extensions")
		if prodDir != expected {
			t.Errorf("ExtensionDir() = %q, want %q", prodDir, expected)
		}
	})
}

func TestConfig_DatabasePath(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	t.Run("production path", func(t *testing.T) {
		t.Parallel()

		config := &powersync.Config{Namespace: ""}
		path, err := config.DatabasePath("acc_123")
		if err != nil {
			t.Fatalf("DatabasePath() error = %v", err)
		}

		expected := filepath.Join(homeDir, ".tero", "data", "acc_123.sqlite")
		if path != expected {
			t.Errorf("DatabasePath() = %q, want %q", path, expected)
		}
	})

	t.Run("dev namespace path", func(t *testing.T) {
		t.Parallel()

		config := &powersync.Config{Namespace: "localhost:3000"}
		path, err := config.DatabasePath("acc_456")
		if err != nil {
			t.Fatalf("DatabasePath() error = %v", err)
		}

		expected := filepath.Join(homeDir, ".tero", "localhost:3000", "data", "acc_456.sqlite")
		if path != expected {
			t.Errorf("DatabasePath() = %q, want %q", path, expected)
		}
	})

	t.Run("ends with .sqlite", func(t *testing.T) {
		t.Parallel()

		config := &powersync.Config{}
		path, err := config.DatabasePath("test-account")
		if err != nil {
			t.Fatalf("DatabasePath() error = %v", err)
		}

		if !strings.HasSuffix(path, ".sqlite") {
			t.Errorf("DatabasePath() = %q, should end with .sqlite", path)
		}
	})
}
