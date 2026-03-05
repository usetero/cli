package sqlite

import (
	"context"
	"time"
)

const (
	DefaultTimeout = 5 * time.Second
)

// WithTimeout applies a default timeout to the given context.
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return context.WithTimeout(ctx, timeout)
}
