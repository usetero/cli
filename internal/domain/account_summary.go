package domain

// PolicyLogStatus is the lifecycle status for a policy.
type PolicyLogStatus string

const (
	PolicyLogStatusPending   PolicyLogStatus = "PENDING"
	PolicyLogStatusApproved  PolicyLogStatus = "APPROVED"
	PolicyLogStatusDismissed PolicyLogStatus = "DISMISSED"
)

func (s PolicyLogStatus) String() string { return string(s) }

// AccountSummary mirrors datadog_account_statuses_cache aggregated across
// all Datadog accounts. All columns included; callers pick what they need.
type AccountSummary struct {
	ReadyForUse bool

	// Health.
	Health    ServiceHealth
	Error     string
	ErrorAt   string
	Warning   string
	WarningAt string

	// Service counts.
	ServiceCount     int64
	ActiveServices   int64
	OkServices       int64
	ErrorServices    int64
	StaleServices    int64
	DisabledServices int64
	InactiveServices int64

	// Event counts.
	EventCount       int64
	AnalyzedCount    int64
	QuarantinedCount int64

	// Policy counts.
	PendingPolicyCount   int64
	ApprovedPolicyCount  int64
	DismissedPolicyCount int64

	// Service-level throughput (ground truth).
	TotalServiceVolumePerHour *float64
	TotalServiceCostPerHour   *float64

	// Log event throughput (discovered events).
	TotalCostPerHour       *float64 // nil when pricing unavailable
	TotalCostPerHourBytes  *float64
	TotalCostPerHourVolume *float64
	TotalVolumePerHour     float64
	TotalBytesPerHour      float64

	// Estimated savings from pending policies.
	EstimatedCostPerHour       *float64 // nil when pricing unavailable
	EstimatedCostPerHourBytes  *float64
	EstimatedCostPerHourVolume *float64
	EstimatedVolumePerHour     float64
	EstimatedBytesPerHour      float64

	// Observed impact from approved policies (before/after).
	ObservedCostBefore   *float64 // nil when pricing unavailable
	ObservedCostAfter    *float64
	ObservedVolumeBefore float64
	ObservedVolumeAfter  float64
	ObservedBytesBefore  float64
	ObservedBytesAfter   float64
}
