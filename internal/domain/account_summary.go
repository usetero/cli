package domain

// PolicyLogStatus is the lifecycle status for a policy.
type PolicyLogStatus string

const (
	PolicyLogStatusPending   PolicyLogStatus = "PENDING"
	PolicyLogStatusApproved  PolicyLogStatus = "APPROVED"
	PolicyLogStatusDismissed PolicyLogStatus = "DISMISSED"
)

func (s PolicyLogStatus) String() string { return string(s) }

// AccountSummary is the aggregated status across all Datadog accounts.
// Sourced from a single query against datadog_account_statuses_cache.
type AccountSummary struct {
	ReadyForUse bool

	// Catalog: services
	ServiceCount     int64
	ActiveServices   int64
	ReadyServices    int64
	BrokenServices   int64
	StaleServices    int64
	DiscoveringCount int64
	AnalyzingCount   int64

	// Catalog: events
	EventCount       int64
	AnalyzedCount    int64
	CleanCount       int64
	PendingCount     int64
	ResolvedCount    int64
	QuarantinedCount int64
	PercentComplete  float64
	WorstStatus      ServiceLogStatus
	LogError         string

	// Policies
	PolicyCount          int64
	PendingPolicyCount   int64
	ApprovedPolicyCount  int64
	DismissedPolicyCount int64

	// Cost & volume: estimated savings from pending policies
	EstimatedCostPerHour       *float64 // nil when pricing unavailable
	EstimatedCostPerHourBytes  *float64
	EstimatedCostPerHourVolume *float64
	EstimatedVolumePerHour     float64
	EstimatedBytesPerHour      float64

	// Cost & volume: observed impact from approved policies
	ObservedCostBefore   *float64 // nil when pricing unavailable
	ObservedCostAfter    *float64
	ObservedVolumeBefore float64
	ObservedVolumeAfter  float64

	// Cost & volume: totals
	TotalCostPerHour   *float64 // nil when pricing unavailable
	TotalVolumePerHour float64
	TotalBytesPerHour  float64
}
