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
			row.ServiceName, row.Health,
			row.LogEventCount, row.LogEventAnalyzedCount,
			row.PolicyPendingCount, row.PolicyApprovedCount, row.PolicyDismissedCount,
			row.PolicyPendingCriticalCount, row.PolicyPendingHighCount, row.PolicyPendingMediumCount, row.PolicyPendingLowCount,
			row.ServiceVolumePerHour,
			row.ServiceDebugVolumePerHour, row.ServiceInfoVolumePerHour, row.ServiceWarnVolumePerHour,
			row.ServiceErrorVolumePerHour, row.ServiceOtherVolumePerHour,
			row.ServiceCostPerHourVolumeUsd,
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
			row.ServiceName, row.Health,
			row.LogEventCount, row.LogEventAnalyzedCount,
			row.PolicyPendingCount, row.PolicyApprovedCount, row.PolicyDismissedCount,
			row.PolicyPendingCriticalCount, row.PolicyPendingHighCount, row.PolicyPendingMediumCount, row.PolicyPendingLowCount,
			row.ServiceVolumePerHour,
			row.ServiceDebugVolumePerHour, row.ServiceInfoVolumePerHour, row.ServiceWarnVolumePerHour,
			row.ServiceErrorVolumePerHour, row.ServiceOtherVolumePerHour,
			row.ServiceCostPerHourVolumeUsd,
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
	health string,
	eventCount, analyzedCount int64,
	policyPending, policyApproved, policyDismissed int64,
	pendingCritical, pendingHigh, pendingMedium, pendingLow int64,
	svcVolume *float64,
	svcDebugVolume, svcInfoVolume, svcWarnVolume, svcErrorVolume, svcOtherVolume *float64,
	svcCostVolume *float64,
	volume, bytes, costUSD, costBytes, costVolume *float64,
	estVolume, estBytes, estCostUSD, estCostBytes, estCostVolume *float64,
	obsVolBefore *float64,
	obsVolAfter interface{},
	obsBytesBefore *float64,
	obsBytesAfter interface{},
	obsCostBefore, obsCostBeforeBytes, obsCostBeforeVolume *float64,
	obsCostAfter, obsCostAfterBytes, obsCostAfterVolume interface{},
) domain.ServiceStatus {
	name := ""
	if serviceName != nil {
		name = *serviceName
	}
	obsVolAfterF := anyToFloatPtr(obsVolAfter)
	obsBytesAfterF := anyToFloatPtr(obsBytesAfter)
	obsCostAfterF := anyToFloatPtr(obsCostAfter)
	obsCostAfterBytesF := anyToFloatPtr(obsCostAfterBytes)
	obsCostAfterVolumeF := anyToFloatPtr(obsCostAfterVolume)

	return domain.ServiceStatus{
		Name:   name,
		Health: domain.ServiceHealth(health),

		LogEventCount:         eventCount,
		LogEventAnalyzedCount: analyzedCount,

		PolicyPendingCount:         policyPending,
		PolicyApprovedCount:        policyApproved,
		PolicyDismissedCount:       policyDismissed,
		PolicyPendingCriticalCount: pendingCritical,
		PolicyPendingHighCount:     pendingHigh,
		PolicyPendingMediumCount:   pendingMedium,
		PolicyPendingLowCount:      pendingLow,

		ServiceVolumePerHour:        svcVolume,
		ServiceDebugVolumePerHour:   svcDebugVolume,
		ServiceInfoVolumePerHour:    svcInfoVolume,
		ServiceWarnVolumePerHour:    svcWarnVolume,
		ServiceErrorVolumePerHour:   svcErrorVolume,
		ServiceOtherVolumePerHour:   svcOtherVolume,
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
		ObservedVolumePerHourAfter:         obsVolAfterF,
		ObservedBytesPerHourBefore:         obsBytesBefore,
		ObservedBytesPerHourAfter:          obsBytesAfterF,
		ObservedCostPerHourBeforeUSD:       obsCostBefore,
		ObservedCostPerHourBeforeBytesUSD:  obsCostBeforeBytes,
		ObservedCostPerHourBeforeVolumeUSD: obsCostBeforeVolume,
		ObservedCostPerHourAfterUSD:        obsCostAfterF,
		ObservedCostPerHourAfterBytesUSD:   obsCostAfterBytesF,
		ObservedCostPerHourAfterVolumeUSD:  obsCostAfterVolumeF,
	}
}

func anyToFloatPtr(v interface{}) *float64 {
	f, ok := toFloat64(v)
	if !ok {
		return nil
	}
	return &f
}
