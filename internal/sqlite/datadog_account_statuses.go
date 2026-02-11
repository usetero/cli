package sqlite

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// DatadogAccountStatuses provides access to Datadog account status data.
type DatadogAccountStatuses interface {
	GetSummary(ctx context.Context) (domain.AccountSummary, error)
}

// datadogAccountStatusesImpl implements DatadogAccountStatuses.
type datadogAccountStatusesImpl struct {
	queries *gen.Queries
}

// GetSummary returns aggregated status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetSummary(ctx context.Context) (domain.AccountSummary, error) {
	row, err := d.queries.GetAccountSummary(ctx)
	if err != nil {
		return domain.AccountSummary{}, WrapSQLiteError(err, "get account summary")
	}

	return domain.AccountSummary{
		ReadyForUse: row.ReadyForUse != 0,

		// Services
		ServiceCount:     row.ServiceCount,
		ActiveServices:   row.ActiveServices,
		ReadyServices:    row.ReadyServices,
		BrokenServices:   row.BrokenServices,
		StaleServices:    row.StaleServices,
		DiscoveringCount: row.DiscoveringCount,
		AnalyzingCount:   row.AnalyzingCount,

		// Events
		EventCount:       row.EventCount,
		AnalyzedCount:    row.AnalyzedCount,
		CleanCount:       row.CleanCount,
		PendingCount:     row.PendingCount,
		ResolvedCount:    row.ResolvedCount,
		QuarantinedCount: row.QuarantinedCount,
		PercentComplete:  row.PercentComplete,
		WorstStatus:      domain.ServiceLogStatus(row.WorstStatus),
		LogError:         fmt.Sprint(row.LogError),

		// Policies
		PolicyCount:          row.PolicyCount,
		PendingPolicyCount:   row.PendingPolicyCount,
		ApprovedPolicyCount:  row.ApprovedPolicyCount,
		DismissedPolicyCount: row.DismissedPolicyCount,

		// Estimated savings
		EstimatedCostPerHour:       row.EstimatedCostPerHour,
		EstimatedCostPerHourBytes:  row.EstimatedCostPerHourBytes,
		EstimatedCostPerHourVolume: row.EstimatedCostPerHourVolume,
		EstimatedVolumePerHour:     row.EstimatedVolumePerHour,
		EstimatedBytesPerHour:      row.EstimatedBytesPerHour,

		// Observed impact
		ObservedCostBefore:   row.ObservedCostBefore,
		ObservedCostAfter:    row.ObservedCostAfter,
		ObservedVolumeBefore: row.ObservedVolumeBefore,
		ObservedVolumeAfter:  row.ObservedVolumeAfter,

		// Totals
		TotalCostPerHour:   row.TotalCostPerHour,
		TotalVolumePerHour: row.TotalVolumePerHour,
		TotalBytesPerHour:  row.TotalBytesPerHour,
	}, nil
}
