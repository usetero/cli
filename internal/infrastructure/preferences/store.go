package preferences

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	domainprefs "github.com/usetero/cli/internal/domains/preferences"
)

// Store persists preferences in ~/.tero/environments/<env>/preferences.json.
type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(env string) (*Store, error) {
	if env == "" {
		panic("preferences store requires env")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Store{
		path: filepath.Join(home, ".tero", "environments", env, "preferences.json"),
	}, nil
}

func (s *Store) Load(_ context.Context) (domainprefs.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return domainprefs.Snapshot{}, nil
		}
		return domainprefs.Snapshot{}, err
	}
	var snapshot domainprefs.Snapshot
	if err := json.Unmarshal(b, &snapshot); err != nil {
		return domainprefs.Snapshot{}, fmt.Errorf("decode preferences: %w", err)
	}
	if snapshot.Role != "" && !snapshot.Role.Valid() {
		snapshot.Role = ""
	}
	return snapshot, nil
}

func (s *Store) Save(_ context.Context, snapshot domainprefs.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, b, 0o600); err != nil {
		return err
	}
	return nil
}
