package domain

// PolicyCategoryStatus is the per-category policy breakdown.
type PolicyCategoryStatus struct {
	Category               string
	PendingCount           int64
	ApprovedCount          int64
	DismissedCount         int64
	EstimatedVolumePerHour float64
	EstimatedBytesPerHour  float64
	RiskLevel              string // predominant risk level
	Benefit                string // predominant benefit type
}
