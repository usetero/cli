package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the Tero CLI configuration stored as YAML.
// It implements the app.Store interface using a map-based structure for flexibility.
// Each org gets its own config file at ~/.tero/environments/{env}/orgs/{orgID}/config.yaml.
type Config struct {
	data  map[string]interface{}
	env   string
	orgID string
}

// Get retrieves a string value by key
func (c *Config) Get(key string) string {
	if v, ok := c.data[key].(string); ok {
		return v
	}
	return ""
}

// Set stores a string value by key
func (c *Config) Set(key string, value string) {
	c.data[key] = value
}

// GetBool retrieves a boolean value by key
func (c *Config) GetBool(key string) bool {
	if v, ok := c.data[key].(bool); ok {
		return v
	}
	return false
}

// SetBool stores a boolean value by key
func (c *Config) SetBool(key string, value bool) {
	c.data[key] = value
}

// GetList retrieves a list of strings by key
func (c *Config) GetList(key string) []string {
	if v, ok := c.data[key].([]interface{}); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if v, ok := c.data[key].([]string); ok {
		return v
	}
	return nil
}

// SetList stores a list of strings by key
func (c *Config) SetList(key string, values []string) {
	c.data[key] = values
}

// envDir returns the environment-level directory: ~/.tero/environments/{env}/.
func envDir(env string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tero", "environments", env), nil
}

// baseDir returns the base directory for org-scoped data.
// With orgID: ~/.tero/environments/{env}/orgs/{orgID}/
// Without orgID: ~/.tero/environments/{env}/ (fallback for first run)
func baseDir(env, orgID string) (string, error) {
	dir, err := envDir(env)
	if err != nil {
		return "", err
	}
	if orgID != "" {
		return filepath.Join(dir, "orgs", orgID), nil
	}
	return dir, nil
}

// BaseDir returns the base directory for all Tero data.
func (c *Config) BaseDir() (string, error) {
	return baseDir(c.env, c.orgID)
}

// ConfigPath returns the config file path for the given env and orgID.
func ConfigPath(env, orgID string) (string, error) {
	dir, err := baseDir(env, orgID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config from disk, migrating from the old layout if needed.
func Load(env, orgID string) (*Config, error) {
	migrate(env)

	path, err := ConfigPath(env, orgID)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{data: make(map[string]interface{}), env: env, orgID: orgID}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfgData map[string]interface{}
	if err := yaml.Unmarshal(data, &cfgData); err != nil {
		return nil, err
	}

	return &Config{data: cfgData, env: env, orgID: orgID}, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	path, err := ConfigPath(c.env, c.orgID)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c.data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// Clear removes the config file from disk
func (c *Config) Clear() error {
	path, err := ConfigPath(c.env, c.orgID)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ActiveOrgID reads the active org ID for the given environment.
// Returns "" if no active org is set (first run or after reset).
func ActiveOrgID(env string) string {
	dir, err := envDir(env)
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(dir, "active.yaml"))
	if err != nil {
		return ""
	}

	var active struct {
		OrgID string `yaml:"org_id"`
	}
	if err := yaml.Unmarshal(data, &active); err != nil {
		return ""
	}
	return active.OrgID
}

// SetActiveOrgID writes the active org ID for the given environment.
func SetActiveOrgID(env, orgID string) error {
	dir, err := envDir(env)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	active := struct {
		OrgID string `yaml:"org_id"`
	}{OrgID: orgID}

	data, err := yaml.Marshal(active)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "active.yaml"), data, 0o600)
}

// migrate runs all migration steps in order.
func migrate(env string) {
	migrateToEnvLayout(env)
	migrateToOrgLayout(env)
}

// migrateToEnvLayout moves data from the old layout to the environments/ layout.
// Old layout: ~/.tero/config.yaml (production), ~/.tero/{host}/config.yaml (others)
// New layout: ~/.tero/environments/{env}/config.yaml
// This is idempotent — if the new layout already exists, it's a no-op.
func migrateToEnvLayout(env string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	teroDir := filepath.Join(homeDir, ".tero")
	envDir, err := envDir(env)
	if err != nil {
		return
	}

	// Already migrated — new layout exists.
	if _, err := os.Stat(filepath.Join(envDir, "config.yaml")); err == nil {
		return
	}

	// Determine old directory based on environment name.
	var oldDir string
	if env == "prd" {
		// Production was at ~/.tero/ (root level).
		oldDir = teroDir
	} else {
		// Non-production was at ~/.tero/{env}/ (env was the host string).
		oldDir = filepath.Join(teroDir, env)
	}

	oldConfig := filepath.Join(oldDir, "config.yaml")
	if _, err := os.Stat(oldConfig); os.IsNotExist(err) {
		return // Nothing to migrate.
	}

	// Create new directory.
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		return
	}

	// Move config.yaml.
	_ = os.Rename(oldConfig, filepath.Join(envDir, "config.yaml"))

	// Move databases (old: data/, new: databases/).
	oldDataDir := filepath.Join(oldDir, "data")
	newDBDir := filepath.Join(envDir, "databases")
	if files, _ := filepath.Glob(filepath.Join(oldDataDir, "*.sqlite")); len(files) > 0 {
		_ = os.MkdirAll(newDBDir, 0o755)
		for _, f := range files {
			_ = os.Rename(f, filepath.Join(newDBDir, filepath.Base(f)))
		}
		_ = os.Remove(oldDataDir) // Remove empty old data dir.
	}

	// Clean up stale files at root.
	_ = os.Remove(filepath.Join(teroDir, "tero.db"))

	// Remove old non-production directory if empty.
	if env != "prd" {
		_ = os.Remove(oldDir)
	}
}

// migrateToOrgLayout moves data from the flat env layout to per-org layout.
// Old layout: ~/.tero/environments/{env}/config.yaml (with default_org_id inside)
// New layout: ~/.tero/environments/{env}/orgs/{orgID}/config.yaml + active.yaml
// Idempotent: skips if orgs/ directory already exists or if no default_org_id is set.
func migrateToOrgLayout(env string) {
	dir, err := envDir(env)
	if err != nil {
		return
	}

	// Already migrated — orgs/ directory exists.
	if _, err := os.Stat(filepath.Join(dir, "orgs")); err == nil {
		return
	}

	// Read the env-level config to get default_org_id.
	configPath := filepath.Join(dir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // No config to migrate.
	}

	var cfgData map[string]interface{}
	if err := yaml.Unmarshal(data, &cfgData); err != nil {
		return
	}

	orgID, _ := cfgData["default_org_id"].(string)
	if orgID == "" {
		return // Never completed onboarding — nothing to move.
	}

	// Create org directory.
	orgDir := filepath.Join(dir, "orgs", orgID)
	if err := os.MkdirAll(orgDir, 0o755); err != nil {
		return
	}

	// Move config.yaml → orgs/{orgID}/config.yaml
	_ = os.Rename(configPath, filepath.Join(orgDir, "config.yaml"))

	// Move databases/ → orgs/{orgID}/databases/
	oldDBDir := filepath.Join(dir, "databases")
	newDBDir := filepath.Join(orgDir, "databases")
	if files, _ := filepath.Glob(filepath.Join(oldDBDir, "*.sqlite")); len(files) > 0 {
		_ = os.MkdirAll(newDBDir, 0o755)
		for _, f := range files {
			_ = os.Rename(f, filepath.Join(newDBDir, filepath.Base(f)))
		}
		_ = os.Remove(oldDBDir) // Remove empty old databases dir.
	}

	// Write active.yaml with the org ID.
	_ = SetActiveOrgID(env, orgID)
}
