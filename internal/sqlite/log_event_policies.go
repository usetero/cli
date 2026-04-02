package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
	Approve(ctx context.Context, id, userID string) error
	Dismiss(ctx context.Context, id, userID string) error
}

// logEventPoliciesImpl implements LogEventPolicies.
type logEventPoliciesImpl struct {
	read  *gen.Queries
	write *gen.Queries
}

// Count returns the total number of log event policies.
func (l *logEventPoliciesImpl) Count(ctx context.Context) (int64, error) {
	count, err := l.read.CountLogEventPolicies(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count log event policies")
	}
	return count, nil
}

func (l *logEventPoliciesImpl) Approve(ctx context.Context, id, userID string) error {
	// The current control plane no longer exposes mutable approval columns on
	// synced log_event_policies rows. Keep the legacy tool path non-fatal until
	// findings-based actions replace it.
	_ = ctx
	_ = id
	_ = userID
	return nil
}

func (l *logEventPoliciesImpl) Dismiss(ctx context.Context, id, userID string) error {
	_ = ctx
	_ = id
	_ = userID
	return nil
}
