package domain

import (
	"github.com/usetero/cli/internal/format"
)

// CategoryType constants. Drives which tab owns a category.
const (
	CategoryTypeCompliance = "compliance" // Legal/security risk
	CategoryTypeWaste      = "waste"      // Cost reduction
)

// PolicyCategoryStatus is the per-category policy breakdown.
// Float pointers are nil when no data exists (e.g. no pending/approved policies contribute).
type PolicyCategoryStatus struct {
	Category       string
	PendingCount   int64
	ApprovedCount  int64
	DismissedCount int64

	// Estimated impact from pending policies.
	EstimatedVolumePerHour     *float64
	EstimatedBytesPerHour      *float64
	EstimatedCostPerHour       *float64
	EstimatedCostPerHourBytes  *float64
	EstimatedCostPerHourVolume *float64

	// Observed impact from approved policies (before/after).
	ObservedVolumeBefore     *float64
	ObservedVolumeAfter      *float64
	ObservedBytesBefore      *float64
	ObservedBytesAfter       *float64
	ObservedCostBefore       *float64
	ObservedCostBeforeBytes  *float64
	ObservedCostBeforeVolume *float64
	ObservedCostAfter        *float64
	ObservedCostAfterBytes   *float64
	ObservedCostAfterVolume  *float64

	// Volume discovery coverage.
	EventsWithVolumes int64 // Log events in this category that have volume data
	TotalEvents       int64 // Total log events in this category

	// How policies in this category reduce cost.
	ImpactType string // "volume" (drops events) or "attribute" (strips fields)
}

// DisplayName returns a human-readable name for the category.
// Converts slugs like "bot_traffic" to "Bot Traffic".
func (c PolicyCategoryStatus) DisplayName() string {
	return format.TitleCase(c.Category)
}

// ReducesVolume reports whether this category's policies drop entire events.
func (c PolicyCategoryStatus) ReducesVolume() bool {
	return c.EstimatedVolumePerHour != nil && *c.EstimatedVolumePerHour > 0
}

// ImpactLabel returns a human-readable description of how this category reduces cost.
func (c PolicyCategoryStatus) ImpactLabel() string {
	switch c.ImpactType {
	case "volume":
		return "drops events"
	case "attribute":
		return "strips fields"
	default:
		return ""
	}
}
