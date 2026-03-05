package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEvents provides type-safe access to log events.
type LogEvents interface {
	Count(ctx context.Context) (int64, error)
}

// logEventsImpl implements LogEvents.
type logEventsImpl struct {
	queries *gen.Queries
}

// Count returns the total number of log events.
func (l *logEventsImpl) Count(ctx context.Context) (int64, error) {
	count, err := l.queries.CountLogEvents(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count log events")
	}
	return count, nil
}
