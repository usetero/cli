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
		Health: domain.ServiceHealth(fmt.Sprint(row.Health)),

		// Services
		ServiceCount:     row.ServiceCount,
		ActiveServices:   row.ActiveServices,
		OkServices:       row.OkServices,
		DisabledServices: row.DisabledServices,
		InactiveServices: row.InactiveServices,

		// Events
		EventCount:    row.EventCount,
		AnalyzedCount: row.AnalyzedCount,

		// Policies
		PendingPolicyCount:         row.PendingPolicyCount,
		ApprovedPolicyCount:        row.ApprovedPolicyCount,
		DismissedPolicyCount:       row.DismissedPolicyCount,
		PolicyPendingCriticalCount: row.PolicyPendingCriticalCount,
		PolicyPendingHighCount:     row.PolicyPendingHighCount,
		PolicyPendingMediumCount:   row.PolicyPendingMediumCount,
		PolicyPendingLowCount:      row.PolicyPendingLowCount,

		// Estimated savings
		EstimatedCostPerHour:       row.EstimatedCostPerHour,
		EstimatedCostPerHourBytes:  row.EstimatedCostPerHourBytes,
		EstimatedCostPerHourVolume: row.EstimatedCostPerHourVolume,
		EstimatedVolumePerHour:     row.EstimatedVolumePerHour,
		EstimatedBytesPerHour:      row.EstimatedBytesPerHour,

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
