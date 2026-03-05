package uploader

import (
	"time"

	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

// Event is emitted by uploader to report progress and health.
type Event interface {
	event()
}

type SyncingEvent struct {
	ProcessedCount int
}

func (SyncingEvent) event() {}

type StalledEvent struct {
	Error      error
	Table      psdb.TableName
	RowID      string
	StalledFor time.Duration
}

func (StalledEvent) event() {}

type RecoveredEvent struct {
	StalledFor time.Duration
}

func (RecoveredEvent) event() {}
