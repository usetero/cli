package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicyCategoryStatuses provides access to pre-computed per-category policy rollups.
type LogEventPolicyCategoryStatuses interface {
	ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
	ListQualityCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
	ListComplianceCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
	CountObservedByComplianceCategory(ctx context.Context) (map[domain.PolicyCategory]int64, error)
}

// logEventPolicyCategoryStatusesImpl implements LogEventPolicyCategoryStatuses.
type logEventPolicyCategoryStatusesImpl struct {
	queries *gen.Queries
}

// ListWasteCategoryStatuses returns pre-computed category rollups for the waste tab.
func (l *logEventPolicyCategoryStatusesImpl) ListWasteCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	return l.listByType(ctx, domain.CategoryTypeWaste)
}

// ListQualityCategoryStatuses returns pre-computed category rollups for the quality tab.
func (l *logEventPolicyCategoryStatusesImpl) ListQualityCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	return l.listByType(ctx, domain.CategoryTypeQuality)
}

// ListComplianceCategoryStatuses returns pre-computed category rollups for the compliance tab.
func (l *logEventPolicyCategoryStatusesImpl) ListComplianceCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	return l.listByType(ctx, domain.CategoryTypeCompliance)
}

// CountObservedByComplianceCategory returns the number of leaking (observed) pending policies per compliance category.
func (l *logEventPolicyCategoryStatusesImpl) CountObservedByComplianceCategory(ctx context.Context) (map[domain.PolicyCategory]int64, error) {
	rows, err := l.queries.CountObservedPoliciesByComplianceCategory(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "count observed policies by compliance category")
	}

	result := make(map[domain.PolicyCategory]int64, len(rows))
	for _, row := range rows {
		if row.Category != nil {
			result[domain.PolicyCategory(*row.Category)] = row.ObservedCount
		}
	}
	return result, nil
}

func (l *logEventPolicyCategoryStatusesImpl) listByType(ctx context.Context, categoryType domain.CategoryType) ([]domain.PolicyCategoryStatus, error) {
	ct := string(categoryType)
	rows, err := l.queries.ListCategoryStatusesByCostAndType(ctx, &ct)
	if err != nil {
		return nil, WrapSQLiteError(err, "list policy category statuses by type")
	}

	result := make([]domain.PolicyCategoryStatus, len(rows))
	for i, row := range rows {
		result[i] = domain.PolicyCategoryStatus{
			Category:               domain.PolicyCategory(row.Category),
			DisplayName:            row.DisplayName,
			Principle:              row.Principle,
			PendingCount:           row.PendingCount,
			ApprovedCount:          row.ApprovedCount,
			DismissedCount:         row.DismissedCount,
			PolicyPendingCriticalCount: row.PolicyPendingCriticalCount,
			PolicyPendingHighCount:     row.PolicyPendingHighCount,
			PolicyPendingMediumCount:   row.PolicyPendingMediumCount,
			PolicyPendingLowCount:      row.PolicyPendingLowCount,
			EstimatedVolumePerHour: row.EstimatedVolumeReductionPerHour,
			EstimatedBytesPerHour:  row.EstimatedBytesReductionPerHour,
			EstimatedCostPerHour:   row.EstimatedCostReductionPerHourUsd,
			EventsWithVolumes:      row.EventsWithVolumes,
			TotalEvents:            row.TotalEventCount,
			Action:                 domain.PolicyAction(row.PolicyAction),
		}
	}
	return result, nil
}
