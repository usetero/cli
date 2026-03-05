package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
)

// DatabasePath is an absolute path to an account-scoped SQLite database.
type DatabasePath string

func (p DatabasePath) String() string { return string(p) }
func (p DatabasePath) Validate() error {
	if p == "" {
		return fmt.Errorf("database path is required")
	}
	return nil
}

// AccountID identifies an account-scoped local database path.
type AccountID string

func (id AccountID) String() string { return string(id) }
func (id AccountID) Validate() error {
	if id == "" {
		return fmt.Errorf("account id is required")
	}
	return nil
}

// Storage resolves SQLite database file paths.
type Storage struct {
	BaseDir string
	Env     string
	OrgID   string
}

// NewDefaultStorage returns a storage resolver under ~/.tero/environments.
func NewDefaultStorage(env, orgID string) (Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Storage{}, err
	}
	return Storage{
		BaseDir: filepath.Join(home, ".tero", "environments"),
		Env:     env,
		OrgID:   orgID,
	}, nil
}

// DatabasePath returns the account-scoped SQLite path.
func (s Storage) DatabasePath(accountID AccountID) (DatabasePath, error) {
	if s.Env == "" {
		return "", fmt.Errorf("env is required")
	}
	if s.OrgID == "" {
		return "", fmt.Errorf("org id is required")
	}
	if err := accountID.Validate(); err != nil {
		return "", err
	}
	return DatabasePath(filepath.Join(s.BaseDir, s.Env, "orgs", s.OrgID, "accounts", accountID.String(), "tero.sqlite")), nil
}
