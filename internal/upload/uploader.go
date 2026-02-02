// Package upload handles uploading local changes to the backend.
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

const (
	defaultPollInterval = 100 * time.Millisecond
	defaultRetryDelay   = 1 * time.Second
	defaultMaxRetries   = 3
)

// TokenRefresher provides access tokens for authentication.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Uploader watches the CRUD queue and uploads changes to the backend.
type Uploader struct {
	db             sqlite.Database
	queue          *powersync.CrudQueue
	client         powersync.Client
	tokenRefresher TokenRefresher
	handlers       map[string]Handler
	logger         log.Logger

	// Configuration
	pollInterval time.Duration
	retryDelay   time.Duration
	maxRetries   int

	// Event channel for status updates
	events chan Event

	// State tracking
	stalledSince *time.Time
	stalledEntry *powersync.CrudEntry
}

// New creates a new uploader.
func New(
	db sqlite.Database,
	client powersync.Client,
	tokenRefresher TokenRefresher,
	conversations api.Conversations,
	messages chat.Messages,
	logger log.Logger,
) *Uploader {
	return &Uploader{
		db:             db,
		queue:          powersync.NewCrudQueue(db),
		client:         client,
		tokenRefresher: tokenRefresher,
		handlers: map[string]Handler{
			sqlite.TableConversations: newConversationHandler(conversations, logger),
			sqlite.TableMessages:      newMessageHandler(messages, db, logger),
		},
		logger:       logger,
		pollInterval: defaultPollInterval,
		retryDelay:   defaultRetryDelay,
		maxRetries:   defaultMaxRetries,
		events:       make(chan Event, 10),
	}
}

// Events returns the channel for receiving upload status events.
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

		// Refresh token before each upload cycle
		token, err := u.tokenRefresher.GetAccessToken(ctx)
		if err != nil {
			u.logger.Warn("failed to get access token", "error", err)
			u.wait(ctx, u.retryDelay)
			continue
		}
		u.client.SetToken(token)

		// Process all pending entries
		processed, err := u.uploadAll(ctx)
		if err != nil {
			u.handleError(ctx, err)
			u.wait(ctx, u.retryDelay)
			continue
		}

		// Clear stalled state on success
		if u.stalledSince != nil {
			stalledFor := time.Since(*u.stalledSince)
			u.logger.Info("upload queue recovered", "stalledFor", stalledFor.Round(time.Second))
			u.emit(ctx, RecoveredEvent{StalledFor: stalledFor})
			u.stalledSince = nil
			u.stalledEntry = nil
		}

		if processed > 0 {
			u.emit(ctx, SyncingEvent{ProcessedCount: processed})
			continue
		}

		u.wait(ctx, u.pollInterval)
	}
}

// uploadAll uploads all pending CRUD entries following the PowerSync protocol:
// 1. Read entries from queue
// 2. Upload each to backend
// 3. Fetch write checkpoint
// 4. Complete batch atomically
func (u *Uploader) uploadAll(ctx context.Context) (int, error) {
	entries, err := u.queue.GetAllEntries(ctx)
	if err != nil {
		return 0, fmt.Errorf("get entries: %w", err)
	}
	if len(entries) == 0 {
		return 0, nil
	}

	// Upload each entry to backend
	emit := u.emitter(ctx)
	for _, entry := range entries {
		if err := u.uploadEntry(ctx, entry, emit); err != nil {
			return 0, err
		}
	}

	// Get write checkpoint from PowerSync server
	clientID, err := powersync.GetClientID(ctx, u.db.DB())
	if err != nil {
		return 0, fmt.Errorf("get client id: %w", err)
	}

	checkpoint, err := u.client.GetWriteCheckpoint(ctx, clientID)
	if err != nil {
		return 0, fmt.Errorf("get write checkpoint: %w", err)
	}

	// Complete batch atomically
	lastID := entries[len(entries)-1].ID
	if err := powersync.CompleteBatch(ctx, u.db.DB(), lastID, checkpoint); err != nil {
		return 0, fmt.Errorf("complete batch: %w", err)
	}

	u.logger.Debug("completed batch", "count", len(entries), "checkpoint", checkpoint)
	return len(entries), nil
}

// uploadEntry uploads a single entry with retries.
func (u *Uploader) uploadEntry(ctx context.Context, entry *powersync.CrudEntry, emit Emitter) error {
	handler, ok := u.handlers[entry.Table]
	if !ok {
		u.logger.Warn("no handler for table, skipping", "table", entry.Table, "rowId", entry.RowID)
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= u.maxRetries; attempt++ {
		if attempt > 0 {
			u.logger.Debug("retrying upload", "table", entry.Table, "rowId", entry.RowID, "attempt", attempt)
			u.wait(ctx, u.retryDelay*time.Duration(attempt))
		}

		err := handler.Handle(ctx, entry, emit)
		if err == nil {
			u.logger.Debug("uploaded entry", "table", entry.Table, "rowId", entry.RowID, "op", entry.Op)
			return nil
		}

		lastErr = err
		u.logger.Warn("upload failed", "table", entry.Table, "rowId", entry.RowID, "attempt", attempt, "error", err)
	}

	u.stalledEntry = entry
	return lastErr
}

func (u *Uploader) handleError(ctx context.Context, err error) {
	if u.stalledSince == nil {
		now := time.Now()
		u.stalledSince = &now
		u.logger.Warn("upload queue stalled", "error", err)
	}

	u.emit(ctx, StalledEvent{
		Error:      err,
		Table:      u.stalledTable(),
		RowID:      u.stalledRowID(),
		StalledFor: time.Since(*u.stalledSince),
	})
}

func (u *Uploader) stalledTable() string {
	if u.stalledEntry != nil {
		return u.stalledEntry.Table
	}
	return ""
}

func (u *Uploader) stalledRowID() string {
	if u.stalledEntry != nil {
		return u.stalledEntry.RowID
	}
	return ""
}

func (u *Uploader) emit(ctx context.Context, event Event) {
	select {
	case u.events <- event:
	case <-ctx.Done():
	default:
	}
}

func (u *Uploader) emitter(ctx context.Context) Emitter {
	return func(event Event) { u.emit(ctx, event) }
}

func (u *Uploader) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
