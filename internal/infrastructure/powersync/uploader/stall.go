package uploader

import (
	"time"

	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

type stallTracker struct {
	since *time.Time
	entry *psdb.Mutation
}

func (t *stallTracker) Stalled(entry *psdb.Mutation, err error) StalledEvent {
	if t.since == nil {
		now := time.Now()
		t.since = &now
	}
	if entry != nil {
		entryCopy := *entry
		t.entry = &entryCopy
	}

	event := StalledEvent{
		Error:      err,
		StalledFor: time.Since(*t.since),
	}
	if t.entry != nil {
		event.Table = t.entry.Table
		event.RowID = t.entry.RowID
	}
	return event
}

func (t *stallTracker) Recovered() *RecoveredEvent {
	if t.since == nil {
		return nil
	}

	event := &RecoveredEvent{StalledFor: time.Since(*t.since)}
	t.since = nil
	t.entry = nil
	return event
}
