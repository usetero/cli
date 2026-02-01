package database

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// databaseOpenedMsg is sent internally when the database file is open.
type databaseOpenedMsg struct {
	db        sqlite.Database
	accountID string
}

// Database manages the local data layer: SQLite database, syncer, and uploader.
type Database struct {
	ctx             context.Context
	powersyncConfig *powersync.Config
	auth            auth.Auth
	conversations   api.Conversations
	messages        chat.Messages
	logger          log.Logger

	db       sqlite.Database
	syncer   *Syncer
	uploader *Uploader
}

// New creates a new database model.
func New(
	ctx context.Context,
	powersyncConfig *powersync.Config,
	auth auth.Auth,
	conversations api.Conversations,
	messages chat.Messages,
	logger log.Logger,
) *Database {
	return &Database{
		ctx:             ctx,
		powersyncConfig: powersyncConfig,
		auth:            auth,
		conversations:   conversations,
		messages:        messages,
		logger:          logger,
	}
}

// Start opens the database and starts the syncer and uploader.
func (d *Database) Start(accountID string) tea.Cmd {
	return func() tea.Msg {
		dbPath, err := d.powersyncConfig.DatabasePath(accountID)
		if err != nil {
			d.logger.Error("failed to get database path", "error", err)
			return nil
		}

		db, err := sqlite.Open(d.ctx, dbPath)
		if err != nil {
			d.logger.Error("failed to open database", "error", err)
			return nil
		}

		d.logger.Info("database opened", "path", dbPath)
		return databaseOpenedMsg{db: db, accountID: accountID}
	}
}

// Update handles database-related messages.
func (d *Database) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle internal lifecycle messages
	switch msg := msg.(type) {
	case databaseOpenedMsg:
		d.db = msg.db
		d.syncer = NewSyncer(d.ctx, d.powersyncConfig, d.auth, d.logger)
		d.uploader = NewUploader(d.ctx, d.db, d.conversations, d.messages, d.logger)
		return tea.Batch(
			d.syncer.Start(d.db, msg.accountID),
			d.uploader.Start(),
		)
	}

	// Delegate to syncer
	if d.syncer != nil {
		if cmd := d.syncer.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Delegate to uploader
	if d.uploader != nil {
		if cmd := d.uploader.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// DB returns the underlying database, or nil if not yet opened.
func (d *Database) DB() sqlite.Database {
	return d.db
}

// IsReady returns true if the database is open and syncer has completed first sync.
func (d *Database) IsReady() bool {
	return d.db != nil && d.syncer != nil && d.syncer.IsReady()
}

// Close closes the database and stops the syncer and uploader.
func (d *Database) Close() {
	if d.uploader != nil {
		d.uploader.Wait()
	}
	if d.syncer != nil {
		d.syncer.Stop()
	}
	if d.db != nil {
		d.db.Close()
	}
}
