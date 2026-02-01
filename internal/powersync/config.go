package powersync

import (
	"os"
	"path/filepath"
)

// Config holds configuration for connecting to PowerSync.
type Config struct {
	// Endpoint is the PowerSync service URL (e.g., https://powersync.usetero.dev)
	Endpoint string

	// Namespace is the environment namespace (empty for production, host for dev)
	Namespace string
}

// DataDir returns the directory for storing sync data.
// For production: ~/.tero/data/
// For other environments: ~/.tero/{namespace}/data/
func (c *Config) DataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if c.Namespace == "" {
		return filepath.Join(homeDir, ".tero", "data"), nil
	}
	return filepath.Join(homeDir, ".tero", c.Namespace, "data"), nil
}

// ExtensionDir returns the directory for caching the PowerSync extension.
// This is shared across all environments: ~/.tero/extensions/
func (c *Config) ExtensionDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tero", "extensions"), nil
}

// DatabasePath returns the path to the SQLite database for a specific account.
func (c *Config) DatabasePath(accountID string) (string, error) {
	dataDir, err := c.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, accountID+".sqlite"), nil
}

// Clear removes all SQLite database files for this environment.
func (c *Config) Clear() error {
	dataDir, err := c.DataDir()
	if err != nil {
		return err
	}

	files, _ := filepath.Glob(filepath.Join(dataDir, "*.sqlite"))
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
