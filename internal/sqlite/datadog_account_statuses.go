package sqlite

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// CatalogStatus is the aggregated catalog health across all Datadog accounts.
type CatalogStatus struct {
	ReadyForUse      bool
	ServiceCount     int64
	ActiveServices   int64
	EventCount       int64
	AnalyzedCount    int64
	AnalyzingCount   int64
	DiscoveringCount int64
	BrokenServices   int64
	PercentComplete  float64
	WorstStatus      string
	LogError         string
}

// PolicyStatus is the aggregated policy/savings status across all Datadog accounts.
type PolicyStatus struct {
	ReadyForUse            bool
	PendingPolicyCount     int64
	PolicyCount            int64
	ApprovedPolicyCount    int64
	EstimatedCostPerHour   *float64 // nil when pricing unavailable
	EstimatedVolumePerHour float64
	EstimatedBytesPerHour  float64
	ObservedCostBefore     *float64 // nil when pricing unavailable
	ObservedCostAfter      *float64 // nil when pricing unavailable
	ObservedVolumeBefore   float64
	ObservedVolumeAfter    float64
}

// DatadogAccountStatuses provides access to Datadog account status data.
type DatadogAccountStatuses interface {
	GetCatalogStatus(ctx context.Context) (CatalogStatus, error)
	GetPolicyStatus(ctx context.Context) (PolicyStatus, error)
}

// datadogAccountStatusesImpl implements DatadogAccountStatuses.
type datadogAccountStatusesImpl struct {
	queries *gen.Queries
}

// GetCatalogStatus returns aggregated catalog status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetCatalogStatus(ctx context.Context) (CatalogStatus, error) {
	row, err := d.queries.GetCatalogStatus(ctx)
	if err != nil {
		return CatalogStatus{}, WrapSQLiteError(err, "get catalog status")
	}

	return CatalogStatus{
		ReadyForUse:      row.ReadyForUse != 0,
		ServiceCount:     row.ServiceCount,
		ActiveServices:   row.ActiveServices,
		EventCount:       row.EventCount,
		AnalyzedCount:    row.AnalyzedCount,
		AnalyzingCount:   row.AnalyzingCount,
		DiscoveringCount: row.DiscoveringCount,
		BrokenServices:   row.BrokenServices,
		PercentComplete:  row.PercentComplete,
		WorstStatus:      row.WorstStatus,
		LogError:         fmt.Sprint(row.LogError),
	}, nil
}

// GetPolicyStatus returns aggregated policy/savings status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetPolicyStatus(ctx context.Context) (PolicyStatus, error) {
	row, err := d.queries.GetPolicyStatus(ctx)
	if err != nil {
		return PolicyStatus{}, WrapSQLiteError(err, "get policy status")
	}

	return PolicyStatus{
		ReadyForUse:            row.ReadyForUse != 0,
		PendingPolicyCount:     row.PendingPolicyCount,
		PolicyCount:            row.PolicyCount,
		ApprovedPolicyCount:    row.ApprovedPolicyCount,
		EstimatedCostPerHour:   row.EstimatedCostPerHour,
		EstimatedVolumePerHour: row.EstimatedVolumePerHour,
		EstimatedBytesPerHour:  row.EstimatedBytesPerHour,
		ObservedCostBefore:     row.ObservedCostBefore,
		ObservedCostAfter:      row.ObservedCostAfter,
		ObservedVolumeBefore:   row.ObservedVolumeBefore,
		ObservedVolumeAfter:    row.ObservedVolumeAfter,
	}, nil
}
