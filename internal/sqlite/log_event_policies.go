package sqlite

import (
	"context"
	"encoding/json"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicies provides type-safe access to log event policies.
type LogEventPolicies interface {
	Count(ctx context.Context) (int64, error)
	ListCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error)
	ListTopPendingPoliciesByCategory(ctx context.Context, category string, limit int64) ([]domain.WastePolicy, error)
	ListPendingPIIPolicies(ctx context.Context) ([]domain.PIIPolicy, error)
	CountFixedPIIPolicies(ctx context.Context) (int64, error)
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

// ListCategoryStatuses returns policy counts and impact grouped by category.
func (l *logEventPoliciesImpl) ListCategoryStatuses(ctx context.Context) ([]domain.PolicyCategoryStatus, error) {
	rows, err := l.queries.ListPolicyCategoryStatuses(ctx)
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
			EstimatedVolumePerHour: derefFloat(row.EstimatedVolumePerHour),
			EstimatedBytesPerHour:  derefFloat(row.EstimatedBytesPerHour),
			EstimatedCostPerHour:   derefFloat(row.EstimatedCostPerHour),
			Benefit:                row.Benefits,
			ObservedVolumeBefore:   derefFloat(row.ObservedVolumeBefore),
			ObservedVolumeAfter:    derefFloat(row.ObservedVolumeAfter),
			ObservedBytesBefore:    derefFloat(row.ObservedBytesBefore),
			ObservedBytesAfter:     derefFloat(row.ObservedBytesAfter),
			ObservedCostBefore:     derefFloat(row.ObservedCostBefore),
			ObservedCostAfter:      derefFloat(row.ObservedCostAfter),
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
			EstimatedCostPerHour:  row.EstimatedCostPerHour,
			EstimatedBytesPerHour: row.EstimatedBytesPerHour,
		}
	}
	return result, nil
}

// ListPendingPIIPolicies returns pending PII leakage policies sorted by severity then volume.
func (l *logEventPoliciesImpl) ListPendingPIIPolicies(ctx context.Context) ([]domain.PIIPolicy, error) {
	rows, err := l.queries.ListPendingPIIPolicies(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list pending pii policies")
	}

	result := make([]domain.PIIPolicy, 0, len(rows))
	for _, row := range rows {
		p := domain.PIIPolicy{
			LogEventName:  row.LogEventName,
			ServiceName:   row.ServiceName,
			VolumePerHour: row.VolumePerHour,
			MaxSeverity:   domain.PIISeverity(row.MaxSeverity),
		}

		// Parse the analysis JSON to extract PII field paths.
		if row.Analysis != "" {
			var envelope domain.PIIAnalysisEnvelope
			if err := json.Unmarshal([]byte(row.Analysis), &envelope); err == nil && envelope.PIILeakage != nil {
				p.Fields = envelope.PIILeakage.Fields
			}
		}

		result = append(result, p)
	}
	return result, nil
}

// CountFixedPIIPolicies returns the number of approved PII policies.
func (l *logEventPoliciesImpl) CountFixedPIIPolicies(ctx context.Context) (int64, error) {
	count, err := l.queries.CountFixedPIIPolicies(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count fixed pii policies")
	}
	return count, nil
}

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
