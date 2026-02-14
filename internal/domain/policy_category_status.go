package domain

import "github.com/usetero/cli/internal/format"

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

	// Volume discovery coverage.
	EventsWithVolumes int64 // Log events in this category that have volume data
	TotalEvents       int64 // Total log events in this category
}

// DisplayName returns a human-readable name for the category.
// Converts slugs like "bot_traffic" to "Bot Traffic".
func (c PolicyCategoryStatus) DisplayName() string {
	return format.TitleCase(c.Category)
}
