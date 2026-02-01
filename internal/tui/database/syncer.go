package database

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// syncReadyMsg is sent when PowerSync completes its first sync.
type syncReadyMsg struct {
	syncer powersync.Syncer
}

// syncTickMsg triggers a status poll.
type syncTickMsg struct{}

// Syncer manages the PowerSync lifecycle.
type Syncer struct {
	ctx    context.Context
	config *powersync.Config
	auth   auth.Auth
	logger log.Logger

	syncer        powersync.Syncer
	waiting       bool // true while waiting for first sync
	firstSyncDone bool // set by callback when first sync completes
}

// NewSyncer creates a new syncer model.
func NewSyncer(ctx context.Context, config *powersync.Config, auth auth.Auth, logger log.Logger) *Syncer {
	return &Syncer{
		ctx:    ctx,
		config: config,
		auth:   auth,
		logger: logger,
	}
}

// Start starts the sync process. Returns a command that starts syncing and begins status polling.
func (s *Syncer) Start(db sqlite.Database, accountID string) tea.Cmd {
	s.waiting = true
	s.firstSyncDone = false

	return func() tea.Msg {
		syncer := powersync.NewSync(s.config, s.auth, s.logger)
		err := syncer.Start(s.ctx, db, accountID, func() {
			// First sync complete - flag it for the next poll tick
			s.firstSyncDone = true
		})
		if err != nil {
			s.logger.Error("failed to start sync", "error", err)
			return powersync.StatusUpdateMsg{
				Status:    powersync.StatusError,
				LastError: err,
			}
		}

		s.logger.Info("sync started", "accountID", accountID)

		// Return the syncer immediately so we can poll its status
		return syncReadyMsg{syncer: syncer}
	}
}

// tickCmd returns a command that sends a tick after a delay.
func (s *Syncer) tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return syncTickMsg{}
	})
}

// Update handles sync-related messages.
func (s *Syncer) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case syncReadyMsg:
		s.syncer = msg.syncer
		s.waiting = true
		// Start polling for status updates
		return s.tickCmd()

	case syncTickMsg:
		if s.syncer == nil || !s.waiting {
			return nil
		}

		// Check if first sync completed via callback
		if s.firstSyncDone {
			s.waiting = false
			s.logger.Info("first sync complete")
			return func() tea.Msg { return powersync.SyncReadyMsg{} }
		}

		status := s.syncer.Status()
		syncStatus := s.syncer.SyncStatus()
		lastError := s.syncer.LastError()

		// Send status update and continue polling
		return tea.Batch(
			func() tea.Msg {
				return powersync.StatusUpdateMsg{
					Status:     status,
					SyncStatus: syncStatus,
					LastError:  lastError,
				}
			},
			s.tickCmd(),
		)

	case powersync.SyncStatusQueryMsg:
		// Sync step is asking if sync is ready
		if s.syncer != nil && !s.waiting {
			return func() tea.Msg { return powersync.SyncReadyMsg{} }
		}
		// If still waiting, return current status
		if s.syncer != nil {
			return func() tea.Msg {
				return powersync.StatusUpdateMsg{
					Status:     s.syncer.Status(),
					SyncStatus: s.syncer.SyncStatus(),
					LastError:  s.syncer.LastError(),
				}
			}
		}
	}

	return nil
}

// IsReady returns true if sync has completed.
func (s *Syncer) IsReady() bool {
	return s.syncer != nil && !s.waiting
}

// Stop stops the syncer.
func (s *Syncer) Stop() {
	if s.syncer != nil {
		s.syncer.Stop()
	}
}
