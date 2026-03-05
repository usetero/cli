package session

import (
	"context"
	"errors"
	"time"

	"github.com/usetero/cli/internal/domains/tenancy"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

type EventKind string

const (
	EventStarting  EventKind = "starting"
	EventSyncing   EventKind = "syncing"
	EventReady     EventKind = "ready"
	EventStalled   EventKind = "stalled"
	EventRecovered EventKind = "recovered"
	EventStopped   EventKind = "stopped"
	EventError     EventKind = "error"
)

type Event struct {
	Kind           EventKind
	AccountID      tenancy.AccountID
	ProcessedCount int
	Table          psdb.TableName
	RowID          string
	StalledFor     time.Duration
	Err            error
}

func (s *Service) emit(ctx context.Context, event Event) {
	select {
	case s.events <- event:
	case <-ctx.Done():
	default:
	}
}

func (s *Service) forwardUploaderEvents(ctx context.Context, accountID tenancy.AccountID, events <-chan psuploader.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-events:
			if !ok {
				return
			}
			switch v := e.(type) {
			case psuploader.SyncingEvent:
				s.emit(ctx, Event{
					Kind:           EventSyncing,
					AccountID:      accountID,
					ProcessedCount: v.ProcessedCount,
				})
			case psuploader.StalledEvent:
				s.emit(ctx, Event{
					Kind:       EventStalled,
					AccountID:  accountID,
					StalledFor: v.StalledFor,
					Table:      v.Table,
					RowID:      v.RowID,
					Err:        v.Error,
				})
			case psuploader.RecoveredEvent:
				s.emit(ctx, Event{
					Kind:       EventRecovered,
					AccountID:  accountID,
					StalledFor: v.StalledFor,
				})
			}
		}
	}
}

func isNonFatalStopErr(err error) bool {
	return err == nil || errors.Is(err, context.Canceled)
}
