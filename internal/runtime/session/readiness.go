package session

import (
	"context"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Ready implements onboarding ReadinessService style checks.
func (s *Service) Ready(_ context.Context) (bool, error) {
	return s.IsReady(), nil
}

// IsReady reports whether initial sync has completed.
func (s *Service) IsReady() bool {
	s.mu.RLock()
	syncer := s.syncer
	running := s.state.Running
	s.mu.RUnlock()
	if !running || syncer == nil {
		return false
	}
	return syncer.IsReady()
}

// DB returns the active session database, or nil when not running.
func (s *Service) DB() *sqlite.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}
