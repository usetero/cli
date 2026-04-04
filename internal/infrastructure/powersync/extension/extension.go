package extension

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

//go:embed extensions/*.dylib extensions/*.so
var embeddedExtensions embed.FS

// Register configures SQLite to load the PowerSync extension on new connections.
func Register() error {
	extPath, err := Path()
	if err != nil {
		return fmt.Errorf("get extension path: %w", err)
	}
	sqlite.SetExtensionPath(extPath)
	return nil
}

// SchemaJSON returns the embedded PowerSync schema.
func SchemaJSON() string {
	return sqlite.PowerSyncSchemaJSON()
}

// ApplySchema applies the PowerSync schema to the database.
func ApplySchema(ctx context.Context, db *sqlite.DB) error {
	if _, err := db.Exec(ctx, "SELECT powersync_replace_schema(?)", SchemaJSON()); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

// Path returns the path to the PowerSync extension for the current platform.
func Path() (string, error) {
	filename, err := extensionFilename()
	if err != nil {
		return "", err
	}

	data, err := embeddedExtensions.ReadFile("extensions/" + filename)
	if err != nil {
		return "", fmt.Errorf("read embedded extension: %w", err)
	}

	tmpDir := filepath.Join(os.TempDir(), "tero-powersync")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("create temp directory: %w", err)
	}

	path := filepath.Join(tmpDir, filename)
	if info, err := os.Stat(path); err == nil && info.Size() == int64(len(data)) {
		return path, nil
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o755); err != nil {
		return "", fmt.Errorf("write extension: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("rename extension: %w", err)
	}

	return path, nil
}

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
