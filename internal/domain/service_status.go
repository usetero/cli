package domain

// ServiceHealth is the health status for a service or account.
type ServiceHealth string

const (
	ServiceHealthDisabled ServiceHealth = "DISABLED"
	ServiceHealthInactive ServiceHealth = "INACTIVE"
	ServiceHealthError    ServiceHealth = "ERROR"
	ServiceHealthOK       ServiceHealth = "OK"
)

func (s ServiceHealth) String() string { return string(s) }

// ServiceStatus mirrors service_statuses_cache. All columns included;
// callers pick what they need.
type ServiceStatus struct {
	ID     string
	Name   string
	Health ServiceHealth

	// Event counts.
	LogEventCount         int64
	LogEventAnalyzedCount int64

	// Policy counts.
	PolicyPendingCount         int64
	PolicyApprovedCount        int64
	PolicyDismissedCount       int64
	PolicyPendingCriticalCount int64
	PolicyPendingHighCount     int64
	PolicyPendingMediumCount   int64
	PolicyPendingLowCount      int64

	// Service-level throughput (ground truth from service_log_volumes). Nil when unmeasured.
	ServiceVolumePerHour        *float64
	ServiceDebugVolumePerHour   *float64
	ServiceInfoVolumePerHour    *float64
	ServiceWarnVolumePerHour    *float64
	ServiceErrorVolumePerHour   *float64
	ServiceOtherVolumePerHour   *float64
	ServiceCostPerHourVolumeUSD *float64

	// Log event throughput (discovered events subset). Nil when unmeasured.
	LogEventVolumePerHour        *float64
	LogEventBytesPerHour         *float64
	LogEventCostPerHourUSD       *float64
	LogEventCostPerHourBytesUSD  *float64
	LogEventCostPerHourVolumeUSD *float64

	// Estimated savings from pending policies. Nil when unmeasured.
	EstimatedVolumeReductionPerHour     *float64
	EstimatedBytesReductionPerHour      *float64
	EstimatedCostReductionPerHourUSD    *float64
	EstimatedCostReductionPerHourBytes  *float64
	EstimatedCostReductionPerHourVolume *float64

	// Observed impact from approved policies (before/after). Nil when unmeasured.
	ObservedVolumePerHourBefore        *float64
	ObservedVolumePerHourAfter         *float64
	ObservedBytesPerHourBefore         *float64
	ObservedBytesPerHourAfter          *float64
	ObservedCostPerHourBeforeUSD       *float64
	ObservedCostPerHourBeforeBytesUSD  *float64
	ObservedCostPerHourBeforeVolumeUSD *float64
	ObservedCostPerHourAfterUSD        *float64
	ObservedCostPerHourAfterBytesUSD   *float64
	ObservedCostPerHourAfterVolumeUSD  *float64
}
