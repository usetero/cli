package preferencestest

import (
	"context"

	domainprefs "github.com/usetero/cli/internal/domains/preferences"
)

// Store is an in-memory preferences store with optional failure hooks.
type Store struct {
	Snapshot  domainprefs.Snapshot
	LoadErr   error
	SaveErr   error
	SaveCalls int
}

var _ domainprefs.Store = (*Store)(nil)

func (s *Store) Load(context.Context) (domainprefs.Snapshot, error) {
	if s.LoadErr != nil {
		return domainprefs.Snapshot{}, s.LoadErr
	}
	return s.Snapshot, nil
}

func (s *Store) Save(_ context.Context, snapshot domainprefs.Snapshot) error {
	s.SaveCalls++
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.Snapshot = snapshot
	return nil
}
