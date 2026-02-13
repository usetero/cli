package domain

// categoryDisplayNames maps category slugs to human-readable names.
var categoryDisplayNames = map[string]string{
	"instrumentation_bloat":       "Instrumentation Bloat",
	"duplicate_fields":            "Duplicate Fields",
	"accidental_debug_statements": "Debug Statements",
	"noise":                       "Noise",
	"pii_leakage":                 "PII Leakage",
}

// PolicyCategoryStatus is the per-category policy breakdown.
type PolicyCategoryStatus struct {
	Category               string
	PendingCount           int64
	ApprovedCount          int64
	DismissedCount         int64
	EstimatedVolumePerHour float64
	EstimatedBytesPerHour  float64
	EstimatedCostPerHour   float64
	Benefit                string

	// Observed impact from approved policies (before/after).
	ObservedVolumeBefore float64
	ObservedVolumeAfter  float64
	ObservedBytesBefore  float64
	ObservedBytesAfter   float64
	ObservedCostBefore   float64
	ObservedCostAfter    float64
}

// DisplayName returns a human-readable name for the category.
func (c PolicyCategoryStatus) DisplayName() string {
	if name, ok := categoryDisplayNames[c.Category]; ok {
		return name
	}
	return c.Category
}
