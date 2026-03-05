package uploader

import (
	"context"
	"fmt"
	"time"

	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

type RunPolicy struct {
	PollInterval time.Duration
	RetryDelay   time.Duration
	MaxRetries   int
}

func DefaultRunPolicy() RunPolicy {
	return RunPolicy{
		PollInterval: 100 * time.Millisecond,
		RetryDelay:   1 * time.Second,
		MaxRetries:   3,
	}
}

func (p RunPolicy) Validate() error {
	if p.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be > 0")
	}
	if p.RetryDelay <= 0 {
		return fmt.Errorf("retry delay must be > 0")
	}
	if p.MaxRetries < 0 {
		return fmt.Errorf("max retries must be >= 0")
	}
	return nil
}

type MutationHandler interface {
	Handle(ctx context.Context, mutation psdb.Mutation) error
}

type SyncNotifier interface {
	NotifyUploadCompleted(ctx context.Context) error
}

type Option func(*Uploader)

func WithPolicy(policy RunPolicy) Option {
	return func(u *Uploader) {
		u.policy = policy
	}
}

func WithHandler(table psdb.TableName, handler MutationHandler) Option {
	return func(u *Uploader) {
		if handler == nil {
			delete(u.handlers, table)
			return
		}
		u.handlers[table] = handler
	}
}

func WithSyncNotifier(notifier SyncNotifier) Option {
	return func(u *Uploader) {
		u.notifier = notifier
	}
}
