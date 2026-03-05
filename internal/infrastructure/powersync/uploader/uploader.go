package uploader

import (
	"context"
	"fmt"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

// TokenSource provides auth tokens for upload operations.
type TokenSource interface {
	GetAccessToken(ctx context.Context) (psclient.AccessToken, error)
}

// Uploader uploads local PowerSync mutations in batches.
type Uploader struct {
	store    *psdb.Store
	client   psclient.Client
	tokens   TokenSource
	log      logging.Scope
	policy   RunPolicy
	events   chan Event
	handlers map[psdb.TableName]MutationHandler
	notifier SyncNotifier

	stalledSince *time.Time
	stalledEntry *psdb.Mutation
}

func New(store *psdb.Store, client psclient.Client, tokens TokenSource, log logging.Scope, opts ...Option) (*Uploader, error) {
	if store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if tokens == nil {
		return nil, fmt.Errorf("token source is required")
	}

	u := &Uploader{
		store:    store,
		client:   client,
		tokens:   tokens,
		log:      log.Child("powersync_uploader"),
		policy:   DefaultRunPolicy(),
		events:   make(chan Event, 16),
		handlers: make(map[psdb.TableName]MutationHandler),
	}
	for _, opt := range opts {
		opt(u)
	}
	if err := u.policy.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *Uploader) Events() <-chan Event { return u.events }

func (u *Uploader) Run(ctx context.Context) error {
	defer close(u.events)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		processed, err := u.uploadNextBatch(ctx)
		if err != nil {
			u.handleError(ctx, err)
			u.wait(ctx, u.policy.RetryDelay)
			continue
		}

		if u.stalledSince != nil {
			stalledFor := time.Since(*u.stalledSince)
			u.emit(ctx, RecoveredEvent{StalledFor: stalledFor})
			u.stalledSince = nil
			u.stalledEntry = nil
		}

		if processed > 0 {
			u.emit(ctx, SyncingEvent{ProcessedCount: processed})
			continue
		}
		u.wait(ctx, u.policy.PollInterval)
	}
}

func (u *Uploader) uploadNextBatch(ctx context.Context) (int, error) {
	batch, err := u.store.NextMutationBatch(ctx)
	if err != nil {
		return 0, fmt.Errorf("get next mutation batch: %w", err)
	}
	if len(batch) == 0 {
		return 0, nil
	}

	token, err := u.tokens.GetAccessToken(ctx)
	if err != nil {
		return 0, fmt.Errorf("get access token: %w", err)
	}
	u.client.SetToken(token)

	for i := range batch {
		if err := u.handleMutationWithRetry(ctx, batch[i]); err != nil {
			u.stalledEntry = &batch[i]
			return 0, err
		}
	}

	clientID, err := u.store.ClientID(ctx)
	if err != nil {
		return 0, fmt.Errorf("get client id: %w", err)
	}
	checkpoint, err := u.client.GetWriteCheckpoint(ctx, psclient.ClientID(clientID))
	if err != nil {
		return 0, fmt.Errorf("get write checkpoint: %w", err)
	}
	checkpointInt, err := checkpoint.ParseInt()
	if err != nil {
		return 0, fmt.Errorf("parse write checkpoint: %w", err)
	}

	lastID := batch[len(batch)-1].ID
	if err := u.store.CompleteUploadedBatch(ctx, lastID, psdb.OpID(checkpointInt)); err != nil {
		return 0, fmt.Errorf("complete uploaded batch: %w", err)
	}

	if u.notifier != nil {
		if err := u.notifier.NotifyUploadCompleted(ctx); err != nil {
			u.log.Warn("sync notifier failed", "error", err)
		}
	}

	return len(batch), nil
}

func (u *Uploader) handleMutationWithRetry(ctx context.Context, mutation psdb.Mutation) error {
	handler, ok := u.handlers[mutation.Table]
	if !ok {
		u.log.Debug("no upload handler for table; skipping", "table", mutation.Table)
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= u.policy.MaxRetries; attempt++ {
		if attempt > 0 {
			u.wait(ctx, u.policy.RetryDelay*time.Duration(attempt))
		}
		if err := handler.Handle(ctx, mutation); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (u *Uploader) handleError(ctx context.Context, err error) {
	if u.stalledSince == nil {
		now := time.Now()
		u.stalledSince = &now
	}
	event := StalledEvent{Error: err}
	if u.stalledSince != nil {
		event.StalledFor = time.Since(*u.stalledSince)
	}
	if u.stalledEntry != nil {
		event.Table = u.stalledEntry.Table
		event.RowID = u.stalledEntry.RowID
	}
	u.emit(ctx, event)
}

func (u *Uploader) emit(ctx context.Context, event Event) {
	select {
	case u.events <- event:
	case <-ctx.Done():
	default:
	}
}

func (u *Uploader) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
