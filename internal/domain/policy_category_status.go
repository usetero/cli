package domain

// RiskLevel represents the risk level of a policy.
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

func (r RiskLevel) String() string { return string(r) }

// PolicyCategoryStatus is the per-category policy breakdown.
type PolicyCategoryStatus struct {
	Category               string
	PendingCount           int64
	ApprovedCount          int64
	DismissedCount         int64
	EstimatedVolumePerHour float64
	EstimatedBytesPerHour  float64
	RiskLevel              RiskLevel
	Benefit                string
}
