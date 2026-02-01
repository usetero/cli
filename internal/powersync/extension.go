package powersync

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/usetero/cli/internal/sqlite"
)

//go:embed extensions/*.dylib extensions/*.so
var embeddedExtensions embed.FS

// LoadExtension loads the PowerSync extension into the database and applies the schema.
func LoadExtension(ctx context.Context, db sqlite.Database) error {
	extPath, err := ExtensionPath()
	if err != nil {
		return fmt.Errorf("get extension path: %w", err)
	}
	if err := db.DB().LoadExtension(ctx, extPath, "sqlite3_powersync_init"); err != nil {
		return fmt.Errorf("load extension: %w", err)
	}
	if _, err := db.DB().Exec(ctx, "SELECT powersync_replace_schema(?)", SchemaJSON()); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

// ExtensionPath returns the path to the PowerSync extension for the current platform.
// The extension is extracted from the embedded binary to a temporary location.
func ExtensionPath() (string, error) {
	filename, err := extensionFilename()
	if err != nil {
		return "", err
	}

	// Read embedded extension
	data, err := embeddedExtensions.ReadFile("extensions/" + filename)
	if err != nil {
		return "", fmt.Errorf("read embedded extension: %w", err)
	}

	// Write to temp directory
	// We use a consistent path so we don't create new files on every run
	tmpDir := filepath.Join(os.TempDir(), "tero-powersync")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("create temp directory: %w", err)
	}

	path := filepath.Join(tmpDir, filename)

	// Check if already extracted with correct size
	if info, err := os.Stat(path); err == nil && info.Size() == int64(len(data)) {
		return path, nil
	}

	// Write to temp file, then rename for atomicity
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o755); err != nil {
		return "", fmt.Errorf("write extension: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename extension: %w", err)
	}

	return path, nil
}

// extensionFilename returns the platform-specific extension filename.
func extensionFilename() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return "libpowersync_aarch64.macos.dylib", nil
		case "amd64":
			return "libpowersync_x64.macos.dylib", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "arm64":
			return "libpowersync_aarch64.linux.so", nil
		case "amd64":
			return "libpowersync_x64.linux.so", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}
