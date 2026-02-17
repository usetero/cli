package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
	ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
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

// ListWasteCategoryStatuses returns policy counts and impact grouped by category,
// excluding compliance categories (filtered at the SQL level by category_type).
func (l *logEventPoliciesImpl) ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	rows, err := l.queries.ListWasteCategoryStatuses(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list policy category statuses")
	}

	result := make([]domain.PolicyCategoryStatus, len(rows))
	for i, row := range rows {
		result[i] = domain.PolicyCategoryStatus{
			Category:               row.Category,
			PendingCount:           row.PendingCount,
			ApprovedCount:          row.ApprovedCount,
			DismissedCount:         row.DismissedCount,
			EstimatedVolumePerHour: row.EstimatedVolumePerHour,
			EstimatedBytesPerHour:  row.EstimatedBytesPerHour,
			EstimatedCostPerHour:   row.EstimatedCostPerHour,
			ObservedVolumeBefore:   row.ObservedVolumeBefore,
			ObservedVolumeAfter:    row.ObservedVolumeAfter,
			ObservedBytesBefore:    row.ObservedBytesBefore,
			ObservedBytesAfter:     row.ObservedBytesAfter,
			ObservedCostBefore:     row.ObservedCostBefore,
			ObservedCostAfter:      row.ObservedCostAfter,
			EventsWithVolumes:      row.EventsWithVolumes,
			TotalEvents:            row.TotalEvents,
		}
	}
	return result, nil
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
			LogEventName:          row.LogEventName,
			ServiceName:           row.ServiceName,
			VolumePerHour:         row.VolumePerHour,
			BytesPerHour:          row.BytesPerHour,
			EstimatedCostPerHour:  row.EstimatedCostPerHour,
			EstimatedBytesPerHour: row.EstimatedBytesPerHour,
		}
	}
	return result, nil
}
