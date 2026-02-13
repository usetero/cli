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
	ListPIIPolicies(ctx context.Context) ([]domain.PIIPolicy, error)
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
			RiskLevel:              domain.RiskLevel(row.RiskLevel),
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

// ListPIIPolicies returns individual PII leakage policies with service/event context.
func (l *logEventPoliciesImpl) ListPIIPolicies(ctx context.Context) ([]domain.PIIPolicy, error) {
	rows, err := l.queries.ListPIIPolicies(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list pii policies")
	}

	result := make([]domain.PIIPolicy, 0, len(rows))
	for _, row := range rows {
		p := domain.PIIPolicy{
			LogEventName: row.LogEventName,
			ServiceName:  row.ServiceName,
			RiskLevel:    domain.RiskLevel(row.RiskLevel),
			Status:       domain.PolicyLogStatus(row.Status),
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

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
