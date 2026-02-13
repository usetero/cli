package sqlite

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
	ListCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
	ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error)
	ListPendingByCategory(ctx context.Context, category string) ([]domain.PolicyDetail, error)
	Approve(ctx context.Context, id, userID string) error
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

// ListCategoryStatuses returns policy counts and impact grouped by category.
func (l *logEventPoliciesImpl) ListCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	rows, err := l.read.ListPolicyCategoryStatuses(ctx)
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
			Benefit:                row.Benefits,
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

// ListPendingByCategory returns individual pending policies for a category.
func (l *logEventPoliciesImpl) ListPendingByCategory(ctx context.Context, category string) ([]domain.PolicyDetail, error) {
	rows, err := l.read.ListPendingPoliciesByCategory(ctx, &category)
	if err != nil {
		return nil, WrapSQLiteError(err, "list pending policies by category")
	}

	result := make([]domain.PolicyDetail, len(rows))
	for i, row := range rows {
		result[i] = domain.PolicyDetail{
			PolicyID:               derefString(row.PolicyID),
			LogEventName:           row.LogEventName,
			RiskLevel:              domain.RiskLevel(derefString(row.RiskLevel)),
			Benefits:               derefString(row.Benefits),
			EstimatedCostPerHour:   derefFloat(row.EstimatedCostReductionPerHourUsd),
			EstimatedVolumePerHour: derefFloat(row.EstimatedVolumeReductionPerHour),
		}
	}
	return result, nil
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

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
