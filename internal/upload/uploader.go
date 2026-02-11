// Package upload handles uploading local changes to the backend.
package upload

import (
	"context"
	"fmt"
	"time"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	psapi "github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/db"
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
type Uploader interface {
	Run(ctx context.Context) error
	Events() <-chan Event
}

// uploader is the concrete implementation of Uploader.
type uploader struct {
	db             sqlite.DB
	queue          *db.CrudQueue
	client         psapi.Client
	tokenRefresher TokenRefresher
	handlers       map[sqlite.Table]Handler
	scope          log.Scope

	// Configuration
	pollInterval time.Duration
	retryDelay   time.Duration
	maxRetries   int

	// Event channel for status updates
	events chan Event

	// State tracking
	stalledSince *time.Time
	stalledEntry *db.CrudEntry
}

// New creates a new uploader.
func New(
	database sqlite.DB,
	client psapi.Client,
	tokenRefresher TokenRefresher,
	conversations api.Conversations,
	messages api.Messages,
	services api.Services,
	policies api.LogEventPolicies,
	scope log.Scope,
) Uploader {
	scope = scope.Child("upload")
	return &uploader{
		db:             database,
		queue:          db.NewCrudQueue(database),
		client:         client,
		tokenRefresher: tokenRefresher,
		handlers: map[sqlite.Table]Handler{
			sqlite.TableConversations:    newConversationHandler(conversations, scope),
			sqlite.TableMessages:         newMessageHandler(messages, database, scope),
			sqlite.TableServices:         newServiceHandler(services, scope),
			sqlite.TableLogEventPolicies: newPolicyHandler(policies, scope),
		},
		scope:        scope,
		pollInterval: defaultPollInterval,
		retryDelay:   defaultRetryDelay,
		maxRetries:   defaultMaxRetries,
		events:       make(chan Event, 10),
	}
}

// Events returns the channel for receiving upload status events.
func (u *uploader) Events() <-chan Event {
	return u.events
}

// Run starts the upload loop. It blocks until the context is cancelled.
func (u *uploader) Run(ctx context.Context) error {
	u.scope.Info("upload loop started")
	defer func() {
		u.scope.Info("upload loop stopped")
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
			u.scope.Warn("failed to get access token", "error", err)
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
			u.scope.Info("upload queue recovered", "stalledFor", stalledFor.Round(time.Second))
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
func (u *uploader) uploadAll(ctx context.Context) (int, error) {
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
	clientID, err := db.GetClientID(ctx, u.db)
	if err != nil {
		return 0, fmt.Errorf("get client id: %w", err)
	}

	checkpoint, err := u.client.GetWriteCheckpoint(ctx, clientID)
	if err != nil {
		return 0, fmt.Errorf("get write checkpoint: %w", err)
	}

	// Complete batch atomically
	lastID := entries[len(entries)-1].ID
	if err := db.CompleteBatch(ctx, u.db, lastID, checkpoint); err != nil {
		return 0, fmt.Errorf("complete batch: %w", err)
	}

	u.scope.Debug("completed batch", "count", len(entries), "checkpoint", checkpoint)
	return len(entries), nil
}

// uploadEntry uploads a single entry with retries.
func (u *uploader) uploadEntry(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	handler, ok := u.handlers[entry.Table]
	if !ok {
		u.scope.Warn("no handler for table, skipping", "table", entry.Table, "rowId", entry.RowID)
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= u.maxRetries; attempt++ {
		if attempt > 0 {
			u.scope.Debug("retrying upload", "table", entry.Table, "rowId", entry.RowID, "attempt", attempt)
			u.wait(ctx, u.retryDelay*time.Duration(attempt))
		}

		err := handler.Handle(ctx, entry, emit)
		if err == nil {
			u.scope.Debug("uploaded entry", "table", entry.Table, "rowId", entry.RowID, "op", entry.Op)
			return nil
		}

		lastErr = err
		u.scope.Warn("upload failed", "table", entry.Table, "rowId", entry.RowID, "attempt", attempt, "error", err)
	}

	u.stalledEntry = entry
	return lastErr
}

func (u *uploader) handleError(ctx context.Context, err error) {
	if u.stalledSince == nil {
		now := time.Now()
		u.stalledSince = &now
		u.scope.Warn("upload queue stalled", "error", err)
	}

	u.emit(ctx, StalledEvent{
		Error:      err,
		Table:      u.stalledTable(),
		RowID:      u.stalledRowID(),
		StalledFor: time.Since(*u.stalledSince),
	})
}

func (u *uploader) stalledTable() sqlite.Table {
	if u.stalledEntry != nil {
		return u.stalledEntry.Table
	}
	return ""
}

func (u *uploader) stalledRowID() string {
	if u.stalledEntry != nil {
		return u.stalledEntry.RowID
	}
	return ""
}

func (u *uploader) emit(ctx context.Context, event Event) {
	select {
	case u.events <- event:
	case <-ctx.Done():
	default:
	}
}

func (u *uploader) emitter(ctx context.Context) Emitter {
	return func(event Event) { u.emit(ctx, event) }
}

func (u *uploader) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
