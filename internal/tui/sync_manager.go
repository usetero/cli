package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/tui/onboarding/account"
)

// Re-export AccountSelectedMsg for use outside this package.
type AccountSelectedMsg = account.AccountSelectedMsg

// InitialSyncCompletedMsg is sent when the first sync completes successfully.
type InitialSyncCompletedMsg struct{}

// syncStartedMsg is sent when PowerSync has been initialized and started.
type syncStartedMsg struct {
	db  sqlite.Database
	err error
}

// initialSyncCompletedMsg is the internal message for initial sync completion.
type initialSyncCompletedMsg struct {
	err error
}

// Syncer abstracts the PowerSync sync operations for testing.
type Syncer interface {
	Start(ctx context.Context, db sqlite.Database, accountID, token string) error
	Stop()
	Status() powersync.Status
	WaitForFirstSync(ctx context.Context) error
}

// SyncManager handles PowerSync lifecycle.
// It listens for AccountSelectedMsg and starts syncing in the background.
type SyncManager struct {
	ctx            context.Context
	config         *powersync.Config
	sync           Syncer
	tokenRefresher powersync.TokenRefresher
	logger         log.Logger

	// State
	db sqlite.Database
}

// NewSyncManager creates a new sync manager for production use.
func NewSyncManager(ctx context.Context, config *powersync.Config, tokenRefresher powersync.TokenRefresher, logger log.Logger) *SyncManager {
	return &SyncManager{
		ctx:            ctx,
		config:         config,
		sync:           powersync.NewSync(config, tokenRefresher),
		tokenRefresher: tokenRefresher,
		logger:         logger,
	}
}

// NewSyncManagerForTest creates a sync manager with injected dependencies for testing.
func NewSyncManagerForTest(ctx context.Context, config *powersync.Config, sync Syncer, tokenRefresher powersync.TokenRefresher, logger log.Logger) *SyncManager {
	return &SyncManager{
		ctx:            ctx,
		config:         config,
		sync:           sync,
		tokenRefresher: tokenRefresher,
		logger:         logger,
	}
}

// SyncStartedMsgForTest creates a syncStartedMsg for testing.
func SyncStartedMsgForTest(db sqlite.Database, err error) tea.Msg {
	return syncStartedMsg{db: db, err: err}
}

// Update handles sync-related messages.
// Returns nil for messages it doesn't handle.
func (m *SyncManager) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case AccountSelectedMsg:
		m.logger.Info("account selected, starting sync", "accountID", msg.Account.ID)
		return m.startSync(msg.Account.ID)

	case syncStartedMsg:
		if msg.err != nil {
			m.logger.Error("failed to start sync", "error", msg.err)
			return nil
		}
		m.db = msg.db
		m.logger.Info("sync started, waiting for initial sync", "status", m.sync.Status())
		return m.waitForInitialSync()

	case initialSyncCompletedMsg:
		if msg.err != nil {
			m.logger.Error("initial sync failed", "error", msg.err)
			return nil
		}
		m.logger.Info("initial sync completed")
		return func() tea.Msg { return InitialSyncCompletedMsg{} }
	}

	return nil
}

// DB returns the database, or nil if sync hasn't started.
func (m *SyncManager) DB() sqlite.Database {
	return m.db
}

// Sync returns the PowerSync instance.
func (m *SyncManager) Sync() Syncer {
	return m.sync
}

// Shutdown stops sync and closes the database.
func (m *SyncManager) Shutdown() {
	if m.sync != nil {
		m.sync.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
}

// startSync opens the database and starts PowerSync in the background.
func (m *SyncManager) startSync(accountID string) tea.Cmd {
	return func() tea.Msg {
		// Get database path
		dbPath, err := m.config.DatabasePath(accountID)
		if err != nil {
			return syncStartedMsg{err: err}
		}

		// Open database
		db, err := sqlite.Open(dbPath)
		if err != nil {
			return syncStartedMsg{err: err}
		}

		// Get auth token
		token, err := m.tokenRefresher.GetAccessToken(m.ctx)
		if err != nil {
			db.Close()
			return syncStartedMsg{err: err}
		}

		// Start PowerSync
		if err := m.sync.Start(m.ctx, db, accountID, token); err != nil {
			db.Close()
			return syncStartedMsg{err: err}
		}

		return syncStartedMsg{db: db}
	}
}

// waitForInitialSync waits for the first sync to complete in the background.
func (m *SyncManager) waitForInitialSync() tea.Cmd {
	return func() tea.Msg {
		err := m.sync.WaitForFirstSync(m.ctx)
		return initialSyncCompletedMsg{err: err}
	}
}
