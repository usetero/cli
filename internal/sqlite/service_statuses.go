package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// ServiceStatuses provides access to per-service status data.
type ServiceStatuses interface {
	ListAllServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error)
	ListEnabledServiceStatuses(ctx context.Context, limit int64) ([]domain.ServiceStatus, error)
}

type serviceStatusesImpl struct {
	queries *gen.Queries
}

// ListAllServiceStatuses returns all services sorted by severity.
func (s *serviceStatusesImpl) ListAllServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error) {
	rows, err := s.queries.ListAllServiceStatuses(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list all service statuses")
	}

	result := make([]domain.ServiceStatus, len(rows))
	for i, row := range rows {
		result[i] = mapServiceStatus(
			row.ServiceName, row.Health, row.Error, row.ErrorAt, row.Warning, row.WarningAt,
			row.LogEventCount, row.LogEventAnalyzedCount, row.LogEventQuarantinedCount,
			row.PolicyPendingCount, row.PolicyApprovedCount, row.PolicyDismissedCount,
			row.ServiceVolumePerHour, row.ServiceCostPerHourVolumeUsd,
			row.LogEventVolumePerHour, row.LogEventBytesPerHour,
			row.LogEventCostPerHourUsd, row.LogEventCostPerHourBytesUsd, row.LogEventCostPerHourVolumeUsd,
			row.EstimatedVolumeReductionPerHour, row.EstimatedBytesReductionPerHour,
			row.EstimatedCostReductionPerHourUsd, row.EstimatedCostReductionPerHourBytesUsd, row.EstimatedCostReductionPerHourVolumeUsd,
			row.ObservedVolumePerHourBefore, row.ObservedVolumePerHourAfter,
			row.ObservedBytesPerHourBefore, row.ObservedBytesPerHourAfter,
			row.ObservedCostPerHourBeforeUsd, row.ObservedCostPerHourBeforeBytesUsd, row.ObservedCostPerHourBeforeVolumeUsd,
			row.ObservedCostPerHourAfterUsd, row.ObservedCostPerHourAfterBytesUsd, row.ObservedCostPerHourAfterVolumeUsd,
		)
	}
	return result, nil
}

// ListEnabledServiceStatuses returns enabled services (not DISABLED/INACTIVE),
// sorted by severity, limited to the given count.
func (s *serviceStatusesImpl) ListEnabledServiceStatuses(ctx context.Context, limit int64) ([]domain.ServiceStatus, error) {
	rows, err := s.queries.ListEnabledServiceStatuses(ctx, limit)
	if err != nil {
		return nil, WrapSQLiteError(err, "list enabled service statuses")
	}

	result := make([]domain.ServiceStatus, len(rows))
	for i, row := range rows {
		result[i] = mapServiceStatus(
			row.ServiceName, row.Health, row.Error, row.ErrorAt, row.Warning, row.WarningAt,
			row.LogEventCount, row.LogEventAnalyzedCount, row.LogEventQuarantinedCount,
			row.PolicyPendingCount, row.PolicyApprovedCount, row.PolicyDismissedCount,
			row.ServiceVolumePerHour, row.ServiceCostPerHourVolumeUsd,
			row.LogEventVolumePerHour, row.LogEventBytesPerHour,
			row.LogEventCostPerHourUsd, row.LogEventCostPerHourBytesUsd, row.LogEventCostPerHourVolumeUsd,
			row.EstimatedVolumeReductionPerHour, row.EstimatedBytesReductionPerHour,
			row.EstimatedCostReductionPerHourUsd, row.EstimatedCostReductionPerHourBytesUsd, row.EstimatedCostReductionPerHourVolumeUsd,
			row.ObservedVolumePerHourBefore, row.ObservedVolumePerHourAfter,
			row.ObservedBytesPerHourBefore, row.ObservedBytesPerHourAfter,
			row.ObservedCostPerHourBeforeUsd, row.ObservedCostPerHourBeforeBytesUsd, row.ObservedCostPerHourBeforeVolumeUsd,
			row.ObservedCostPerHourAfterUsd, row.ObservedCostPerHourAfterBytesUsd, row.ObservedCostPerHourAfterVolumeUsd,
		)
	}
	return result, nil
}

//nolint:unparam // positional args mirror generated row fields
func mapServiceStatus(
	serviceName *string,
	health, errStr, errorAt, warning, warningAt string,
	eventCount, analyzedCount, quarantinedCount int64,
	policyPending, policyApproved, policyDismissed int64,
	svcVolume, svcCostVolume *float64,
	volume, bytes, costUSD, costBytes, costVolume *float64,
	estVolume, estBytes, estCostUSD, estCostBytes, estCostVolume *float64,
	obsVolBefore, obsVolAfter, obsBytesBefore, obsBytesAfter *float64,
	obsCostBefore, obsCostBeforeBytes, obsCostBeforeVolume *float64,
	obsCostAfter, obsCostAfterBytes, obsCostAfterVolume *float64,
) domain.ServiceStatus {
	name := ""
	if serviceName != nil {
		name = *serviceName
	}
	return domain.ServiceStatus{
		Name:      name,
		Health:    domain.ServiceHealth(health),
		Error:     errStr,
		ErrorAt:   errorAt,
		Warning:   warning,
		WarningAt: warningAt,

		LogEventCount:            eventCount,
		LogEventAnalyzedCount:    analyzedCount,
		LogEventQuarantinedCount: quarantinedCount,

		PolicyPendingCount:   policyPending,
		PolicyApprovedCount:  policyApproved,
		PolicyDismissedCount: policyDismissed,

		ServiceVolumePerHour:        svcVolume,
		ServiceCostPerHourVolumeUSD: svcCostVolume,

		LogEventVolumePerHour:        volume,
		LogEventBytesPerHour:         bytes,
		LogEventCostPerHourUSD:       costUSD,
		LogEventCostPerHourBytesUSD:  costBytes,
		LogEventCostPerHourVolumeUSD: costVolume,

		EstimatedVolumeReductionPerHour:     estVolume,
		EstimatedBytesReductionPerHour:      estBytes,
		EstimatedCostReductionPerHourUSD:    estCostUSD,
		EstimatedCostReductionPerHourBytes:  estCostBytes,
		EstimatedCostReductionPerHourVolume: estCostVolume,

		ObservedVolumePerHourBefore:        obsVolBefore,
		ObservedVolumePerHourAfter:         obsVolAfter,
		ObservedBytesPerHourBefore:         obsBytesBefore,
		ObservedBytesPerHourAfter:          obsBytesAfter,
		ObservedCostPerHourBeforeUSD:       obsCostBefore,
		ObservedCostPerHourBeforeBytesUSD:  obsCostBeforeBytes,
		ObservedCostPerHourBeforeVolumeUSD: obsCostBeforeVolume,
		ObservedCostPerHourAfterUSD:        obsCostAfter,
		ObservedCostPerHourAfterBytesUSD:   obsCostAfterBytes,
		ObservedCostPerHourAfterVolumeUSD:  obsCostAfterVolume,
	}
}
