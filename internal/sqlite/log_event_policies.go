package sqlite

import (
	"context"
	"time"

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
	now := time.Now().UTC().Format(time.RFC3339)
	err := l.write.ApproveLogEventPolicy(ctx, gen.ApproveLogEventPolicyParams{
		ID:         &id,
		ApprovedAt: &now,
		ApprovedBy: &userID,
	})
	if err != nil {
		return WrapSQLiteError(err, "approve log event policy")
	}
	return nil
}

func (l *logEventPoliciesImpl) Dismiss(ctx context.Context, id, userID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	err := l.write.DismissLogEventPolicy(ctx, gen.DismissLogEventPolicyParams{
		ID:          &id,
		DismissedAt: &now,
		DismissedBy: &userID,
	})
	if err != nil {
		return WrapSQLiteError(err, "dismiss log event policy")
	}
	return nil
}
