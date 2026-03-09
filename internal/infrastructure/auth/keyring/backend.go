package keyring

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// BackendSystem stores secrets in the OS keychain.
	BackendSystem = "system"
	// BackendFile stores secrets in a file under the runtime environment directory.
	BackendFile = "file"

	EnvBackend = "TERO_SECRET_STORE"
	EnvPath    = "TERO_SECRET_STORE_PATH"
)

type backend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

func resolveBackend(env string) (backend, error) {
	switch os.Getenv(EnvBackend) {
	case "", BackendSystem:
		return &systemBackend{service: baseServiceName + ":" + env}, nil
	case BackendFile:
		path, err := resolveFilePath(env)
		if err != nil {
			return nil, err
		}
		return newFileBackend(path), nil
	default:
		return nil, fmt.Errorf("invalid secret store backend %q", os.Getenv(EnvBackend))
	}
}

func resolveFilePath(env string) (string, error) {
	if path := os.Getenv(EnvPath); path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tero", "environments", env, "secrets.json"), nil
}
