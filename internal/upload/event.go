package upload

import "time"

// Event is the interface for all upload events.
// Handlers can define their own event types that implement this interface.
type Event interface {
	uploadEvent()
}

// Core upload events - emitted by the uploader itself

// SyncingEvent is emitted when entries are being processed.
type SyncingEvent struct {
	ProcessedCount int
}

func (SyncingEvent) uploadEvent() {}

// StalledEvent is emitted when upload is blocked on a failing entry.
type StalledEvent struct {
	Error      error
	Table      string
	RowID      string
	StalledFor time.Duration
}

func (StalledEvent) uploadEvent() {}

// RecoveredEvent is emitted when upload recovers from a stalled state.
type RecoveredEvent struct {
	StalledFor time.Duration
}

func (RecoveredEvent) uploadEvent() {}
