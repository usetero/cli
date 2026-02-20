package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
}

// logEventPoliciesImpl implements LogEventPolicies.
type logEventPoliciesImpl struct {
	queries *gen.Queries
}

// Count returns the total number of log event policies.
func (l *logEventPoliciesImpl) Count(ctx context.Context) (int64, error) {
	count, err := l.queries.CountLogEventPolicies(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count log event policies")
	}
	return count, nil
}
