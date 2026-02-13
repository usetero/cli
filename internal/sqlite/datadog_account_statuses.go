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

		// Health
		Health:    domain.ServiceHealth(fmt.Sprint(row.Health)),
		Error:     fmt.Sprint(row.Error),
		ErrorAt:   fmt.Sprint(row.ErrorAt),
		Warning:   fmt.Sprint(row.Warning),
		WarningAt: fmt.Sprint(row.WarningAt),

		// Services
		ServiceCount:     row.ServiceCount,
		ActiveServices:   row.ActiveServices,
		OkServices:       row.OkServices,
		ErrorServices:    row.ErrorServices,
		StaleServices:    row.StaleServices,
		DisabledServices: row.DisabledServices,
		InactiveServices: row.InactiveServices,

		// Events
		EventCount:       row.EventCount,
		AnalyzedCount:    row.AnalyzedCount,
		QuarantinedCount: row.QuarantinedCount,

		// Policies
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
		ObservedBytesBefore:  row.ObservedBytesBefore,
		ObservedBytesAfter:   row.ObservedBytesAfter,

		// Totals
		TotalCostPerHour:       row.TotalCostPerHour,
		TotalCostPerHourBytes:  row.TotalCostPerHourBytes,
		TotalCostPerHourVolume: row.TotalCostPerHourVolume,
		TotalVolumePerHour:     row.TotalVolumePerHour,
		TotalBytesPerHour:      row.TotalBytesPerHour,

		// Service-level throughput
		TotalServiceVolumePerHour: row.TotalServiceVolumePerHour,
		TotalServiceCostPerHour:   row.TotalServiceCostPerHour,
	}, nil
}
