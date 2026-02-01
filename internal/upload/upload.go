// Package upload handles uploading local changes to the backend.
// It watches the PowerSync CRUD queue and dispatches entries to appropriate services.
package upload

import (
	"context"
	"fmt"
	"time"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// Status represents the current state of the upload queue.
type Status int

const (
	StatusIdle      Status = iota // Queue is empty, waiting
	StatusSyncing                 // Actively uploading entries
	StatusStalled                 // Upload failed, retrying
	StatusRecovered               // Recovered from stalled state
)

// Event is sent when the upload status changes.
type Event struct {
	Status         Status
	Error          error
	Table          string        // Table of the problematic entry (when stalled)
	RowID          string        // Row ID of the problematic entry (when stalled)
	StalledFor     time.Duration // How long we've been stalled
	ProcessedCount int           // Number of entries processed (when syncing)
}

// handler processes CRUD entries for a specific table.
type handler interface {
	Handle(ctx context.Context, entry *powersync.CrudEntry) error
}

// Uploader watches the CRUD queue and uploads changes to the backend.
type Uploader struct {
	queue    *powersync.CrudQueue
	handlers map[string]handler
	logger   log.Logger

	// Configuration
	pollInterval time.Duration
	retryDelay   time.Duration
	maxRetries   int

	// Event channel for status updates
	events chan Event

	// State tracking for visibility
	stalledSince *time.Time
	stalledEntry *powersync.CrudEntry
}

// New creates a new uploader with all handlers wired up.
func New(db sqlite.Database, conversations api.Conversations, messages chat.Messages, logger log.Logger) *Uploader {
	u := &Uploader{
		queue:        powersync.NewCrudQueue(db),
		handlers:     make(map[string]handler),
		logger:       logger,
		pollInterval: 100 * time.Millisecond,
		retryDelay:   1 * time.Second,
		maxRetries:   3,
		events:       make(chan Event, 10), // Buffered to avoid blocking uploader
	}

	u.handlers["conversations"] = newConversationHandler(conversations, logger)
	u.handlers["messages"] = newMessageHandler(messages, db, logger)

	return u
}

// Events returns the channel for receiving upload status events.
// Consumers should read from this channel to track upload progress.
func (u *Uploader) Events() <-chan Event {
	return u.events
}

// Run starts the upload loop. It blocks until the context is cancelled.
func (u *Uploader) Run(ctx context.Context) error {
	u.logger.Info("upload loop started")
	defer func() {
		u.logger.Info("upload loop stopped")
		close(u.events)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Process all pending entries
		processed, err := u.processQueue(ctx)
		if err != nil {
			// Track stalled state for visibility
			if u.stalledSince == nil {
				now := time.Now()
				u.stalledSince = &now
				u.logger.Warn("upload queue stalled", "error", err)
			} else {
				u.logger.Debug("upload queue still stalled",
					"stalledFor", time.Since(*u.stalledSince).Round(time.Second),
					"error", err,
				)
			}

			// Emit stalled event
			u.emit(ctx, Event{
				Status:     StatusStalled,
				Error:      err,
				Table:      u.stalledEntryTable(),
				RowID:      u.stalledEntryRowID(),
				StalledFor: time.Since(*u.stalledSince),
			})

			// Wait before retrying on error
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(u.retryDelay):
			}
			continue
		}

		// If we were stalled, log and emit recovery
		if u.stalledSince != nil {
			stalledFor := time.Since(*u.stalledSince)
			u.logger.Info("upload queue recovered",
				"stalledFor", stalledFor.Round(time.Second),
			)
			u.emit(ctx, Event{
				Status:     StatusRecovered,
				StalledFor: stalledFor,
			})
			u.stalledSince = nil
			u.stalledEntry = nil
		}

		// If we processed entries, emit syncing event and check for more
		if processed > 0 {
			u.emit(ctx, Event{
				Status:         StatusSyncing,
				ProcessedCount: processed,
			})
			continue
		}

		// No entries - wait before polling again
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(u.pollInterval):
		}
	}
}

// emit sends an event to the channel without blocking.
func (u *Uploader) emit(ctx context.Context, event Event) {
	select {
	case u.events <- event:
	case <-ctx.Done():
	default:
		// Channel full, drop event (logging already captures this)
	}
}

// stalledEntryTable returns the table of the stalled entry, if any.
func (u *Uploader) stalledEntryTable() string {
	if u.stalledEntry != nil {
		return u.stalledEntry.Table
	}
	return ""
}

// stalledEntryRowID returns the row ID of the stalled entry, if any.
func (u *Uploader) stalledEntryRowID() string {
	if u.stalledEntry != nil {
		return u.stalledEntry.RowID
	}
	return ""
}

// processQueue processes all pending CRUD entries.
// Returns the number of entries processed.
func (u *Uploader) processQueue(ctx context.Context) (int, error) {
	processed := 0

	for {
		entry, err := u.queue.GetNextEntry(ctx)
		if err != nil {
			return processed, fmt.Errorf("get next entry: %w", err)
		}
		if entry == nil {
			// Queue is empty
			return processed, nil
		}

		if err := u.processEntry(ctx, entry); err != nil {
			return processed, fmt.Errorf("process entry %d: %w", entry.ID, err)
		}

		processed++
	}
}

// processEntry processes a single CRUD entry with retries.
func (u *Uploader) processEntry(ctx context.Context, entry *powersync.CrudEntry) error {
	handler, ok := u.handlers[entry.Table]
	if !ok {
		u.logger.Warn("no handler for table, skipping",
			"table", entry.Table,
			"id", entry.ID,
			"op", entry.Op,
		)
		// Delete unhandled entries to avoid blocking the queue
		return u.queue.DeleteEntry(ctx, entry.ID)
	}

	var lastErr error
	for attempt := 0; attempt <= u.maxRetries; attempt++ {
		if attempt > 0 {
			u.logger.Debug("retrying upload",
				"table", entry.Table,
				"id", entry.ID,
				"attempt", attempt,
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(u.retryDelay * time.Duration(attempt)):
			}
		}

		if err := handler.Handle(ctx, entry); err != nil {
			lastErr = err
			u.logger.Warn("upload failed",
				"table", entry.Table,
				"id", entry.ID,
				"attempt", attempt,
				"error", err,
			)
			continue
		}

		// Success - delete from queue
		if err := u.queue.DeleteEntry(ctx, entry.ID); err != nil {
			return fmt.Errorf("delete entry after upload: %w", err)
		}

		u.logger.Debug("uploaded entry",
			"table", entry.Table,
			"rowId", entry.RowID,
			"op", entry.Op,
		)
		return nil
	}

	// Log at Error level - this entry is now blocking the entire queue
	u.logger.Error("upload blocked",
		"table", entry.Table,
		"rowId", entry.RowID,
		"op", entry.Op,
		"attempts", u.maxRetries+1,
		"error", lastErr,
	)
	u.stalledEntry = entry

	return fmt.Errorf("upload failed after %d retries: %w", u.maxRetries, lastErr)
}
