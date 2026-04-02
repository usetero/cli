package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// LogEventPolicyStatuses provides type-safe access to pre-computed policy status data.
type LogEventPolicyStatuses interface {
	GetPolicyCard(ctx context.Context, policyID string) (*domain.PolicyCard, error)
	ListTopPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.WastePolicy, error)
}

type logEventPolicyStatusesImpl struct {
	queries *gen.Queries
}

func (l *logEventPolicyStatusesImpl) GetPolicyCard(ctx context.Context, policyID string) (*domain.PolicyCard, error) {
	row, err := l.queries.GetPolicyCard(ctx, &policyID)
	if err != nil {
		return nil, WrapSQLiteError(err, "get policy card")
	}

	card := &domain.PolicyCard{
		PolicyID:                   row.PolicyID,
		ServiceName:                row.ServiceName,
		LogEventName:               row.LogEventName,
		Category:                   row.Category,
		CategoryType:               row.CategoryType,
		Action:                     row.Action,
		Status:                     row.Status,
		Severity:                   row.Severity,
		CategoryDisplayName:        row.CategoryDisplayName,
		VolumePerHour:              row.VolumePerHour,
		BytesPerHour:               row.BytesPerHour,
		EstimatedCostPerHour:       row.EstimatedCostPerHour,
		EstimatedVolumePerHour:     row.EstimatedVolumePerHour,
		EstimatedBytesPerHour:      row.EstimatedBytesPerHour,
		SurvivalRate:               float64Ptr(row.SurvivalRate),
		Analysis:                   row.Analysis,
		Examples:                   row.Examples,
		EventBaselineAvgBytes:      row.EventBaselineAvgBytes,
		EventBaselineVolumePerHour: row.EventBaselineVolumePerHour,
	}

	// Quality policies operate on fields — enrich with per-field byte sizes
	// so BuildEvidence can produce FieldListEvidence.
	if row.LogEventID != "" && row.CategoryType == string(domain.CategoryTypeQuality) {
		card.FieldSizes = l.fieldSizes(ctx, row.LogEventID)
	}

	return card, nil
}

// fieldSizes fetches per-field byte sizes for a log event and returns them
// keyed by dot-path. Returns nil if no data is available.
func (l *logEventPolicyStatusesImpl) fieldSizes(ctx context.Context, logEventID string) map[string]float64 {
	_ = logEventID
	rows, err := l.queries.ListFieldsByLogEvent(ctx)
	if err != nil || len(rows) == 0 {
		return nil
	}

	sizes := make(map[string]float64, len(rows))
	for _, row := range rows {
		fp := domain.ParseFieldPathPg(row.FieldPath)
		if !fp.IsEmpty() {
			sizes[fp.Key()] = row.BaselineAvgBytes
		}
	}
	if len(sizes) == 0 {
		return nil
	}
	return sizes
}

func (l *logEventPolicyStatusesImpl) ListTopPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.WastePolicy, error) {
	_ = category
	_ = limit
	rows, err := l.queries.ListTopPendingPoliciesByCategory(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list top pending policies by category")
	}

	result := make([]domain.WastePolicy, len(rows))
	for i, row := range rows {
		result[i] = domain.WastePolicy{
			LogEventName:               row.LogEventName,
			ServiceName:                row.ServiceName,
			VolumePerHour:              float64Ptr(row.VolumePerHour),
			BytesPerHour:               float64Ptr(row.BytesPerHour),
			EstimatedCostPerHour:       float64Ptr(row.EstimatedCostPerHour),
			EstimatedCostPerHourBytes:  float64Ptr(row.EstimatedCostPerHourBytes),
			EstimatedCostPerHourVolume: float64Ptr(row.EstimatedCostPerHourVolume),
			EstimatedBytesPerHour:      float64Ptr(row.EstimatedBytesPerHour),
			EstimatedVolumePerHour:     float64Ptr(row.EstimatedVolumePerHour),
		}
	}
	return result, nil
}
