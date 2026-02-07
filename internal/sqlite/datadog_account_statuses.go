package sqlite

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// DatadogAccountStatuses provides access to Datadog account status data.
type DatadogAccountStatuses interface {
	GetCatalogSummary(ctx context.Context) (domain.CatalogSummary, error)
	GetPolicySummary(ctx context.Context) (domain.PolicySummary, error)
}

// datadogAccountStatusesImpl implements DatadogAccountStatuses.
type datadogAccountStatusesImpl struct {
	queries *gen.Queries
}

// GetCatalogSummary returns aggregated catalog status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetCatalogSummary(ctx context.Context) (domain.CatalogSummary, error) {
	row, err := d.queries.GetCatalogStatus(ctx)
	if err != nil {
		return domain.CatalogSummary{}, WrapSQLiteError(err, "get catalog summary")
	}

	return domain.CatalogSummary{
		ReadyForUse:      row.ReadyForUse != 0,
		ServiceCount:     row.ServiceCount,
		ActiveServices:   row.ActiveServices,
		EventCount:       row.EventCount,
		AnalyzedCount:    row.AnalyzedCount,
		AnalyzingCount:   row.AnalyzingCount,
		DiscoveringCount: row.DiscoveringCount,
		BrokenServices:   row.BrokenServices,
		StaleServices:    row.StaleServices,
		PercentComplete:  row.PercentComplete,
		WorstStatus:      domain.ServiceLogStatus(row.WorstStatus),
		LogError:         fmt.Sprint(row.LogError),
	}, nil
}

// GetPolicySummary returns aggregated policy/savings status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetPolicySummary(ctx context.Context) (domain.PolicySummary, error) {
	row, err := d.queries.GetPolicyStatus(ctx)
	if err != nil {
		return domain.PolicySummary{}, WrapSQLiteError(err, "get policy summary")
	}

	return domain.PolicySummary{
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
		TotalCostPerHour:       row.TotalCostPerHour,
		TotalVolumePerHour:     row.TotalVolumePerHour,
		TotalBytesPerHour:      row.TotalBytesPerHour,
	}, nil
}
