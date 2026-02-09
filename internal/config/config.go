package config

import (
	"os"
	"path/filepath"

	"github.com/usetero/cli/internal/domain"
	"gopkg.in/yaml.v3"
)

// Config is a YAML-backed key-value store.
// It implements the preferences.Store interface using a map-based structure.
type Config struct {
	data  map[string]any
	path  string // absolute path to the YAML file
	env   string
	orgID domain.OrganizationID
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
	if v, ok := c.data[key].([]any); ok {
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
func baseDir(env string, orgID domain.OrganizationID) (string, error) {
	dir, err := envDir(env)
	if err != nil {
		return "", err
	}
	if orgID != "" {
		return filepath.Join(dir, "orgs", orgID.String()), nil
	}
	return dir, nil
}

// BaseDir returns the base directory for all Tero data.
func (c *Config) BaseDir() (string, error) {
	return baseDir(c.env, c.orgID)
}

// ConfigPath returns the config file path for the given env and orgID.
func ConfigPath(env string, orgID domain.OrganizationID) (string, error) {
	dir, err := baseDir(env, orgID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// UserPreferencesPath returns the env-level user preferences path.
func UserPreferencesPath(env string) (string, error) {
	dir, err := envDir(env)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "preferences.yaml"), nil
}

// OrgPreferencesPath returns the org-level preferences path.
func OrgPreferencesPath(env string, orgID domain.OrganizationID) (string, error) {
	dir, err := baseDir(env, orgID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "preferences.yaml"), nil
}

// Load reads an org-scoped config from disk, migrating from old layouts if needed.
func Load(env string, orgID domain.OrganizationID) (*Config, error) {
	migrate(env)

	path, err := ConfigPath(env, orgID)
	if err != nil {
		return nil, err
	}

	return loadFromPath(path, env, orgID)
}

// LoadUserPreferences reads the env-level user preferences from disk.
func LoadUserPreferences(env string) (*Config, error) {
	migrate(env)

	path, err := UserPreferencesPath(env)
	if err != nil {
		return nil, err
	}

	return loadFromPath(path, env, "")
}

// LoadOrgPreferences reads org-scoped preferences from disk.
func LoadOrgPreferences(env string, orgID domain.OrganizationID) (*Config, error) {
	migrate(env)

	path, err := OrgPreferencesPath(env, orgID)
	if err != nil {
		return nil, err
	}

	return loadFromPath(path, env, orgID)
}

// loadFromPath reads a YAML config from the given path.
func loadFromPath(path, env string, orgID domain.OrganizationID) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{data: make(map[string]any), path: path, env: env, orgID: orgID}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfgData map[string]any
	if err := yaml.Unmarshal(data, &cfgData); err != nil {
		return nil, err
	}

	return &Config{data: cfgData, path: path, env: env, orgID: orgID}, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c.data)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0o600)
}

// Clear removes the config file from disk.
func (c *Config) Clear() error {
	err := os.Remove(c.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ActiveOrgID reads the active org ID from env-level user preferences.
// Returns "" if no active org is set (first run or after reset).
func ActiveOrgID(env string) domain.OrganizationID {
	path, err := UserPreferencesPath(env)
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var prefs map[string]any
	if err := yaml.Unmarshal(data, &prefs); err != nil {
		return ""
	}

	if id, ok := prefs["active_org_id"].(string); ok {
		return domain.OrganizationID(id)
	}
	return ""
}

// SetActiveOrgID writes the active org ID to env-level user preferences.
func SetActiveOrgID(env string, orgID domain.OrganizationID) error {
	cfg, err := LoadUserPreferences(env)
	if err != nil {
		return err
	}
	cfg.Set("active_org_id", orgID.String())
	return cfg.Save()
}

// migrate is called by Load/LoadPreferences before reading config files.
func migrate(env string) {
	Migrate(env)
}
