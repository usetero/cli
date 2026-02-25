package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventStatuses provides access to per-log-event status data.
type LogEventStatuses interface {
	ListByService(ctx context.Context, serviceName string, limit int64) ([]domain.LogEventStatus, error)
}

type logEventStatusesImpl struct {
	queries *gen.Queries
}

// ListByService returns log event statuses for a service, ordered by cost.
func (l *logEventStatusesImpl) ListByService(ctx context.Context, serviceName string, limit int64) ([]domain.LogEventStatus, error) {
	rows, err := l.queries.ListLogEventStatusesByService(ctx, gen.ListLogEventStatusesByServiceParams{
		Name:  &serviceName,
		Limit: limit,
	})
	if err != nil {
		return nil, WrapSQLiteError(err, "list log event statuses by service")
	}

	result := make([]domain.LogEventStatus, len(rows))
	for i, row := range rows {
		result[i] = domain.LogEventStatus{
			Name:                row.LogEventName,
			VolumePerHour:       row.VolumePerHour,
			BytesPerHour:        row.BytesPerHour,
			CostPerHourUSD:      row.CostPerHourUsd,
			PendingPolicyCount:         row.PendingPolicyCount,
			ApprovedPolicyCount:        row.ApprovedPolicyCount,
			PolicyPendingCriticalCount: row.PolicyPendingCriticalCount,
			PolicyPendingHighCount:     row.PolicyPendingHighCount,
			PolicyPendingMediumCount:   row.PolicyPendingMediumCount,
			PolicyPendingLowCount:      row.PolicyPendingLowCount,
		}
	}
	return result, nil
}
