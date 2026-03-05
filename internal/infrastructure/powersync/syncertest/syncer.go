package syncertest

import (
	"context"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Mock is a functional mock for PowerSync syncer lifecycle calls.
type Mock struct {
	StartFn                 func(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error
	StopFn                  func()
	IsReadyFn               func() bool
	NotifyUploadCompletedFn func(ctx context.Context) error
}

func (m *Mock) Start(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
	if m.StartFn == nil {
		if onFirstSync != nil {
			onFirstSync()
		}
		return nil
	}
	return m.StartFn(ctx, db, accountID, onFirstSync)
}

func (m *Mock) Stop() {
	if m.StopFn != nil {
		m.StopFn()
	}
}

func (m *Mock) IsReady() bool {
	if m.IsReadyFn == nil {
		return false
	}
	return m.IsReadyFn()
}

func (m *Mock) NotifyUploadCompleted(ctx context.Context) error {
	if m.NotifyUploadCompletedFn == nil {
		return nil
	}
	return m.NotifyUploadCompletedFn(ctx)
}
