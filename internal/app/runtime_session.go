package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/usetero/cli/internal/domain"
	psapi "github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/upload"
)

// openDatabase opens the SQLite database for the given account.
func (m *Model) openDatabase(accountID string) error {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			m.scope.Warn("failed to close previous database", "error", err)
		}
		m.db = nil
	}

	dbPath, err := m.storage.DatabasePath(accountID)
	if err != nil {
		return err
	}

	db, err := sqlite.Open(m.ctx, dbPath)
	if err != nil {
		return err
	}

	m.db = db
	m.scope.Info("database opened", "path", dbPath)
	return nil
}

// startSync starts the syncer and uploader with the open database.
func (m *Model) startSync(accountID string) error {
	if m.db == nil {
		return fmt.Errorf("database not open")
	}
	if m.sessionCancel != nil {
		m.shutdown()
		if err := m.openDatabase(accountID); err != nil {
			return err
		}
	}

	// Create a session context that is cancelled on shutdown.
	sessionCtx, cancel := context.WithCancel(m.ctx)
	m.sessionCancel = cancel

	if err := m.syncer.Start(sessionCtx, m.db, accountID, nil); err != nil {
		cancel()
		m.sessionCancel = nil
		return err
	}
	m.scope.Info("syncer started", "account_id", accountID)

	// Set account ID on services
	m.services.SetAccountID(domain.AccountID(accountID))

	// Create PowerSync API client for write checkpoints
	psClient := psapi.NewClient(m.cfg.PowerSyncEndpoint)

	// Create and start uploader
	m.uploader = upload.New(
		m.db,
		psClient,
		m.authService,
		m.services.Conversations,
		m.services.Messages,
		m.services.Services,
		m.services.Policies,
		m.scope,
		upload.WithBatchCompletedHook(func(ctx context.Context) error {
			return m.syncer.NotifyUploadCompleted(ctx)
		}),
	)
	go func() {
		if err := m.uploader.Run(sessionCtx); err != nil && !errors.Is(err, context.Canceled) {
			m.scope.Error("uploader error", "error", err)
		}
	}()
	m.scope.Info("uploader started", "account_id", accountID)

	return nil
}

func (m *Model) shutdown() {
	db := m.db

	if m.sessionCancel != nil {
		m.sessionCancel()
		m.sessionCancel = nil
	}
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if db != nil {
		if err := db.Close(); err != nil {
			m.scope.Warn("failed to close database", "error", err)
		}
	}
	m.db = nil
	m.uploader = nil
}
