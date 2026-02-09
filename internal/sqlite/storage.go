package sqlite

import (
	"os"
	"path/filepath"

	"github.com/usetero/cli/internal/config"
)

// Storage provides database file operations.
type Storage interface {
	DatabasePath(accountID string) (string, error)
	ClearDatabase(accountID string) error
	Clear() error
}

// StorageService implements Storage.
type StorageService struct {
	config *config.Config
}

// Ensure StorageService implements Storage.
var _ Storage = (*StorageService)(nil)

// NewStorageService creates a new storage service.
func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{config: cfg}
}

// dataDir returns the directory for storing database files.
func (s *StorageService) dataDir() (string, error) {
	baseDir, err := s.config.BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "databases"), nil
}

// DatabasePath returns the path to the SQLite database for a specific account.
func (s *StorageService) DatabasePath(accountID string) (string, error) {
	dataDir, err := s.dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, accountID+".sqlite"), nil
}

// ClearDatabase removes the SQLite database file for a specific account.
func (s *StorageService) ClearDatabase(accountID string) error {
	dbPath, err := s.DatabasePath(accountID)
	if err != nil {
		return err
	}
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Clear removes all SQLite database files.
func (s *StorageService) Clear() error {
	dataDir, err := s.dataDir()
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
