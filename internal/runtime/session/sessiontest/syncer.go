package sessiontest

import (
	"context"
	"sync/atomic"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type Syncer struct {
	StartErr       error
	Started        bool
	Stopped        bool
	Ready          bool
	AccountID      pssyncer.AccountID
	StartCalls     atomic.Int32
	NotifyUploadFn func(context.Context) error
}

func (s *Syncer) Start(_ context.Context, _ *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error {
	s.StartCalls.Add(1)
	s.AccountID = accountID
	if s.StartErr != nil {
		return s.StartErr
	}
	s.Started = true
	if s.Ready {
		if onFirstSync != nil {
			onFirstSync()
		}
	}
	return nil
}

func (s *Syncer) Stop() { s.Stopped = true }
func (s *Syncer) IsReady() bool {
	return s.Ready
}
func (s *Syncer) NotifyUploadCompleted(ctx context.Context) error {
	if s.NotifyUploadFn != nil {
		return s.NotifyUploadFn(ctx)
	}
	return nil
}
