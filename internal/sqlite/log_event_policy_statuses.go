package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicyStatuses provides type-safe access to pre-computed policy status data.
type LogEventPolicyStatuses interface {
	ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error)
}

type logEventPolicyStatusesImpl struct {
	queries *gen.Queries
}

func (l *logEventPolicyStatusesImpl) ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error) {
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
