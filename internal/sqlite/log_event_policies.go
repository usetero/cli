package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
	ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error)
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

// ListTopPendingPoliciesByCategory returns the top pending policies for a category, ordered by cost then volume.
func (l *logEventPoliciesImpl) ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error) {
	rows, err := l.queries.ListTopPendingPoliciesByCategory(ctx, gen.ListTopPendingPoliciesByCategoryParams{
		Category: &category,
		Limit:    limit,
	})
	if err != nil {
		return nil, WrapSQLiteError(err, "list top pending policies by category")
	}

	result := make([]domain.WastePolicy, len(rows))
	for i, row := range rows {
		result[i] = domain.WastePolicy{
			LogEventName:               row.LogEventName,
			ServiceName:                row.ServiceName,
			VolumePerHour:              row.VolumePerHour,
			BytesPerHour:               row.BytesPerHour,
			EstimatedCostPerHour:       row.EstimatedCostPerHour,
			EstimatedCostPerHourBytes:  row.EstimatedCostPerHourBytes,
			EstimatedCostPerHourVolume: row.EstimatedCostPerHourVolume,
			EstimatedBytesPerHour:      row.EstimatedBytesPerHour,
			EstimatedVolumePerHour:     row.EstimatedVolumePerHour,
		}
	}
	return result, nil
}
