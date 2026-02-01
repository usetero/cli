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
	}

	u.handlers["conversations"] = newConversationHandler(conversations, logger)
	u.handlers["messages"] = newMessageHandler(messages, db, logger)

	return u
}

// Run starts the upload loop. It blocks until the context is cancelled.
func (u *Uploader) Run(ctx context.Context) error {
	u.logger.Info("upload loop started")
	defer u.logger.Info("upload loop stopped")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Process all pending entries
		processed, err := u.processQueue(ctx)
		if err != nil {
			u.logger.Error("error processing upload queue", "error", err)
			// Wait before retrying on error
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(u.retryDelay):
			}
			continue
		}

		// If we processed entries, check for more immediately
		if processed > 0 {
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

	return fmt.Errorf("upload failed after %d retries: %w", u.maxRetries, lastErr)
}
