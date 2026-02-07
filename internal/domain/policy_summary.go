package domain

// PolicyLogStatus is the lifecycle status for a policy.
type PolicyLogStatus string

const (
	PolicyLogStatusPending   PolicyLogStatus = "PENDING"
	PolicyLogStatusApproved  PolicyLogStatus = "APPROVED"
	PolicyLogStatusDismissed PolicyLogStatus = "DISMISSED"
)

func (s PolicyLogStatus) String() string { return string(s) }

// PolicySummary is the aggregated policy impact across all Datadog accounts.
type PolicySummary struct {
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
	TotalCostPerHour       *float64 // nil when pricing unavailable
	TotalVolumePerHour     float64
	TotalBytesPerHour      float64
}
