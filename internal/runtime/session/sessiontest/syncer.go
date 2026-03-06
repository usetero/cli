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
	StateValue     pssyncer.State
	AccountID      pssyncer.AccountID
	StartCalls     atomic.Int32
	NotifyUploadFn func(context.Context) error
}

var _ interface {
	Start(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error
	Stop()
	State() pssyncer.State
	IsReady() bool
	NotifyUploadCompleted(ctx context.Context) error
} = (*Syncer)(nil)

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
func (s *Syncer) State() pssyncer.State {
	if s.StateValue != nil {
		return s.StateValue
	}
	if s.Ready {
		return &pssyncer.Ready{}
	}
	return &pssyncer.Connecting{}
}
func (s *Syncer) IsReady() bool {
	return s.Ready
}
func (s *Syncer) NotifyUploadCompleted(ctx context.Context) error {
	if s.NotifyUploadFn != nil {
		return s.NotifyUploadFn(ctx)
	}
	return nil
}
