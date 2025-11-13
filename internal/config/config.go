package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config is the Tero CLI configuration stored as YAML.
// It implements the app.Store interface using a map-based structure for flexibility.
type Config struct {
	data map[string]interface{}
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

// Path returns the config file path (~/.tero/config.yaml)
func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tero", "config.yaml"), nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{data: make(map[string]interface{})}, nil // Empty config if file doesn't exist
	}
	if err != nil {
		return nil, err
	}

	var cfgData map[string]interface{}
	if err := yaml.Unmarshal(data, &cfgData); err != nil {
		return nil, err
	}

	return &Config{data: cfgData}, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	path, err := Path()
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
