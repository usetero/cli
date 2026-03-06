package session

import (
	"context"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
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

// SyncState returns the current PowerSync lifecycle state for UI rendering.
func (s *Service) SyncState() pssyncer.State {
	s.mu.RLock()
	syncer := s.syncer
	running := s.state.Running
	s.mu.RUnlock()
	if !running || syncer == nil {
		return &pssyncer.Disconnected{}
	}
	return syncer.State()
}

// Status returns the TUI-facing projection for lifecycle + sync progress.
func (s *Service) Status() Status {
	s.mu.RLock()
	running := s.state.Running
	scope := s.scope
	syncer := s.syncer
	s.mu.RUnlock()

	status := Status{
		Running: running,
		Scope:   scope,
		Sync:    &pssyncer.Disconnected{},
	}
	if !running || syncer == nil {
		return status
	}
	status.Ready = syncer.IsReady()
	status.Sync = syncer.State()
	if status.Sync == nil {
		status.Sync = &pssyncer.Disconnected{}
	}
	return status
}

// DB returns the active session database, or nil when not running.
func (s *Service) DB() *sqlite.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}
