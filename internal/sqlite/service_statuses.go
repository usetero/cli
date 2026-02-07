package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// ServiceStatuses provides access to per-service status data.
type ServiceStatuses interface {
	ListServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error)
}

type serviceStatusesImpl struct {
	queries *gen.Queries
}

// ListServiceStatuses returns all services with their catalog status, sorted by severity.
func (s *serviceStatusesImpl) ListServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error) {
	rows, err := s.queries.ListServiceStatuses(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list service statuses")
	}

	result := make([]domain.ServiceStatus, len(rows))
	for i, row := range rows {
		name := ""
		if row.ServiceName != nil {
			name = *row.ServiceName
		}
		result[i] = domain.ServiceStatus{
			Name:            name,
			Status:          domain.ServiceLogStatus(row.LogStatus),
			Error:           row.LogError,
			PercentComplete: row.LogPercentComplete,
			EventCount:      row.LogEventCount,
			AnalyzedCount:   row.LogAnalyzedCount,
			VolumePerHour:   row.LogVolumePerHour,
			BytesPerHour:    row.LogBytesPerHour,
			CostPerHourUSD:  row.LogCostPerHourUsd,
		}
	}
	return result, nil
}
