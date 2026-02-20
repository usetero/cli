package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicyCategoryStatuses provides access to pre-computed per-category policy rollups.
type LogEventPolicyCategoryStatuses interface {
	ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
}

// logEventPolicyCategoryStatusesImpl implements LogEventPolicyCategoryStatuses.
type logEventPolicyCategoryStatusesImpl struct {
	queries *gen.Queries
}

// ListWasteCategoryStatuses returns pre-computed category rollups for the waste tab.
func (l *logEventPolicyCategoryStatusesImpl) ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	rows, err := l.queries.ListWastePolicyCategoryStatuses(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list waste policy category statuses")
	}

	result := make([]domain.PolicyCategoryStatus, len(rows))
	for i, row := range rows {
		result[i] = domain.PolicyCategoryStatus{
			Category:               row.Category,
			PendingCount:           row.PendingCount,
			ApprovedCount:          row.ApprovedCount,
			DismissedCount:         row.DismissedCount,
			EstimatedVolumePerHour: row.EstimatedVolumeReductionPerHour,
			EstimatedBytesPerHour:  row.EstimatedBytesReductionPerHour,
			EstimatedCostPerHour:   row.EstimatedCostReductionPerHourUsd,
			EventsWithVolumes:      row.EventsWithVolumes,
			TotalEvents:            row.TotalEventCount,
			ImpactType:             row.ImpactType,
		}
	}
	return result, nil
}
