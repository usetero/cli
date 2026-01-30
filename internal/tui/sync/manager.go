// Package sync handles PowerSync lifecycle for the TUI.
package sync

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

// CompletedMsg is sent when initial sync completes successfully.
type CompletedMsg struct {
	DB sqlite.Database
}

// syncStartedMsg is sent when PowerSync has been initialized and started.
type syncStartedMsg struct {
	db  sqlite.Database
	err error
}

// syncCompletedMsg is the internal message for initial sync completion.
type syncCompletedMsg struct {
	err error
}

// Syncer abstracts the PowerSync sync operations for testing.
type Syncer interface {
	Start(ctx context.Context, db sqlite.Database, accountID, token string) error
	Stop()
	Status() powersync.Status
	WaitForFirstSync(ctx context.Context) error
}

// Manager handles PowerSync lifecycle.
// It listens for AccountSelectedMsg and starts syncing in the background.
type Manager struct {
	ctx            context.Context
	config         *powersync.Config
	syncer         Syncer
	tokenRefresher powersync.TokenRefresher
	logger         log.Logger

	// State
	db sqlite.Database
}

// NewManager creates a new sync manager for production use.
func NewManager(ctx context.Context, config *powersync.Config, tokenRefresher powersync.TokenRefresher, logger log.Logger) *Manager {
	return &Manager{
		ctx:            ctx,
		config:         config,
		syncer:         powersync.NewSync(config, tokenRefresher),
		tokenRefresher: tokenRefresher,
		logger:         logger,
	}
}

// NewManagerForTest creates a sync manager with injected dependencies for testing.
func NewManagerForTest(ctx context.Context, config *powersync.Config, syncer Syncer, tokenRefresher powersync.TokenRefresher, logger log.Logger) *Manager {
	return &Manager{
		ctx:            ctx,
		config:         config,
		syncer:         syncer,
		tokenRefresher: tokenRefresher,
		logger:         logger,
	}
}

// StartedMsgForTest creates a syncStartedMsg for testing.
func StartedMsgForTest(db sqlite.Database, err error) tea.Msg {
	return syncStartedMsg{db: db, err: err}
}

// Update handles sync-related messages.
// Returns nil for messages it doesn't handle.
func (m *Manager) Update(msg tea.Msg) tea.Cmd {
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

		// Install update hooks for change notifications
		if err := m.db.InstallUpdateHooks(); err != nil {
			m.logger.Error("failed to install update hooks", "error", err)
			// Non-fatal: continue without change notifications
		}

		m.logger.Info("sync started, waiting for initial sync", "status", m.syncer.Status())
		return m.waitForInitialSync()

	case syncCompletedMsg:
		if msg.err != nil {
			m.logger.Error("initial sync failed", "error", msg.err)
			return nil
		}
		m.logger.Info("initial sync completed")
		return func() tea.Msg { return CompletedMsg{DB: m.db} }
	}

	return nil
}

// DB returns the database, or nil if sync hasn't started.
func (m *Manager) DB() sqlite.Database {
	return m.db
}

// Syncer returns the PowerSync instance.
func (m *Manager) Syncer() Syncer {
	return m.syncer
}

// Shutdown stops sync and closes the database.
func (m *Manager) Shutdown() {
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
}

// startSync opens the database and starts PowerSync in the background.
func (m *Manager) startSync(accountID string) tea.Cmd {
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
		if err := m.syncer.Start(m.ctx, db, accountID, token); err != nil {
			db.Close()
			return syncStartedMsg{err: err}
		}

		return syncStartedMsg{db: db}
	}
}

// waitForInitialSync waits for the first sync to complete in the background.
func (m *Manager) waitForInitialSync() tea.Cmd {
	return func() tea.Msg {
		err := m.syncer.WaitForFirstSync(m.ctx)
		return syncCompletedMsg{err: err}
	}
}
