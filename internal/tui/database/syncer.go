package database

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// syncReadyMsg is sent when PowerSync has started.
type syncReadyMsg struct{}

// syncTickMsg triggers a status poll.
type syncTickMsg struct{}

// Syncer manages the PowerSync lifecycle.
type Syncer struct {
	ctx     context.Context
	storage sqlite.Storage
	sync    powersync.Syncer
	logger  log.Logger

	waiting       bool // true while waiting for first sync
	firstSyncDone bool // set by callback when first sync completes
}

// NewSyncer creates a new syncer model.
func NewSyncer(ctx context.Context, storage sqlite.Storage, sync powersync.Syncer, logger log.Logger) *Syncer {
	return &Syncer{
		ctx:     ctx,
		storage: storage,
		sync:    sync,
		logger:  logger,
	}
}

// Start starts the sync process. Returns a command that starts syncing and begins status polling.
func (s *Syncer) Start(db sqlite.Database, accountID string) tea.Cmd {
	s.waiting = true
	s.firstSyncDone = false

	return func() tea.Msg {
		err := s.sync.Start(s.ctx, db, accountID, func() {
			// First sync complete - flag it for the next poll tick
			s.firstSyncDone = true
		})
		if err != nil {
			// If database is corrupt, delete it and retry once.
			// Sync will repopulate from the server.
			if errors.Is(err, powersync.ErrDatabaseCorrupt) {
				s.logger.Warn("database corrupt, resetting", "error", err)
				db.Close()

				if clearErr := s.storage.ClearDatabase(accountID); clearErr != nil {
					s.logger.Error("failed to clear corrupt database", "error", clearErr)
				}

				// Re-open and retry
				dbPath, _ := s.storage.DatabasePath(accountID)
				newDB, openErr := sqlite.Open(s.ctx, dbPath)
				if openErr != nil {
					s.logger.Error("failed to reopen database after reset", "error", openErr)
					return powersync.StatusUpdateMsg{
						Status:    powersync.StatusError,
						LastError: openErr,
					}
				}

				err = s.sync.Start(s.ctx, newDB, accountID, func() {
					s.firstSyncDone = true
				})
				if err != nil {
					s.logger.Error("failed to start sync after reset", "error", err)
					return powersync.StatusUpdateMsg{
						Status:    powersync.StatusError,
						LastError: err,
					}
				}

				s.logger.Info("sync started after database reset", "accountID", accountID)
				return syncReadyMsg{}
			}

			s.logger.Error("failed to start sync", "error", err)
			return powersync.StatusUpdateMsg{
				Status:    powersync.StatusError,
				LastError: err,
			}
		}

		s.logger.Info("sync started", "accountID", accountID)

		// Return ready so we can start polling status
		return syncReadyMsg{}
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
	switch msg.(type) {
	case syncReadyMsg:
		s.waiting = true
		// Start polling for status updates
		return s.tickCmd()

	case syncTickMsg:
		if !s.waiting {
			return nil
		}

		// Check if first sync completed via callback
		if s.firstSyncDone {
			s.waiting = false
			s.logger.Info("first sync complete")
			return func() tea.Msg { return powersync.SyncReadyMsg{} }
		}

		status := s.sync.Status()
		syncStatus := s.sync.SyncStatus()
		lastError := s.sync.LastError()

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
		if !s.waiting {
			return func() tea.Msg { return powersync.SyncReadyMsg{} }
		}
		// If still waiting, return current status
		return func() tea.Msg {
			return powersync.StatusUpdateMsg{
				Status:     s.sync.Status(),
				SyncStatus: s.sync.SyncStatus(),
				LastError:  s.sync.LastError(),
			}
		}
	}

	return nil
}

// IsReady returns true if sync has completed.
func (s *Syncer) IsReady() bool {
	return s.sync.IsRunning() && !s.waiting
}

// Stop stops the syncer.
func (s *Syncer) Stop() {
	s.sync.Stop()
}
