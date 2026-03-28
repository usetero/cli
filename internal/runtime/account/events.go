package account

import (
	"context"
	"time"

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
	Scope          Scope
	ProcessedCount int
	Table          psdb.TableName
	RowID          string
	StalledFor     time.Duration
	Err            error
}

func (r *Runtime) Events() <-chan Event {
	return r.events
}

func (r *Runtime) emit(ctx context.Context, event Event) {
	select {
	case r.events <- event:
	case <-ctx.Done():
	default:
	}
}

func (r *Runtime) forwardUploaderEvents(ctx context.Context, events <-chan psuploader.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			switch value := event.(type) {
			case psuploader.SyncingEvent:
				r.emit(ctx, Event{
					Kind:           EventSyncing,
					Scope:          r.scope,
					ProcessedCount: value.ProcessedCount,
				})
			case psuploader.StalledEvent:
				r.emit(ctx, Event{
					Kind:       EventStalled,
					Scope:      r.scope,
					StalledFor: value.StalledFor,
					Table:      value.Table,
					RowID:      value.RowID,
					Err:        value.Error,
				})
			case psuploader.RecoveredEvent:
				r.emit(ctx, Event{
					Kind:       EventRecovered,
					Scope:      r.scope,
					StalledFor: value.StalledFor,
				})
			}
		}
	}
}
