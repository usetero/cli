package uploader

import (
	"context"
	"fmt"
	"sync"
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
	processor *batchProcessor
	tracker   stallTracker
	events    chan Event

	mu         sync.Mutex
	runStarted bool
}

func New(store *psdb.Store, client psclient.Client, tokens TokenSource, log logging.Scope, opts ...Option) *Uploader {
	if store == nil {
		panic("powersync uploader requires store")
	}
	if client == nil {
		panic("powersync uploader requires client")
	}
	if tokens == nil {
		panic("powersync uploader requires token source")
	}

	u := &Uploader{
		processor: newBatchProcessor(store, client, tokens, log.Child("powersync_uploader")),
		events:    make(chan Event, 16),
	}
	for _, opt := range opts {
		opt(u)
	}
	if err := u.processor.policy.Validate(); err != nil {
		panic(fmt.Sprintf("powersync uploader requires valid policy: %v", err))
	}
	return u
}

func (u *Uploader) Events() <-chan Event { return u.events }

func (u *Uploader) Run(ctx context.Context) error {
	if err := u.markRunStarted(); err != nil {
		return err
	}
	defer close(u.events)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := u.processor.UploadNextBatch(ctx)
		if err != nil {
			u.emit(ctx, u.tracker.Stalled(result.stalledEntry, err))
			u.wait(ctx, u.processor.policy.RetryDelay)
			continue
		}

		if recovered := u.tracker.Recovered(); recovered != nil {
			u.emit(ctx, *recovered)
		}

		if result.processedCount > 0 {
			u.emit(ctx, SyncingEvent{ProcessedCount: result.processedCount})
			continue
		}
		u.wait(ctx, u.processor.policy.PollInterval)
	}
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

func (u *Uploader) markRunStarted() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.runStarted {
		return ErrAlreadyRun
	}
	u.runStarted = true
	return nil
}
