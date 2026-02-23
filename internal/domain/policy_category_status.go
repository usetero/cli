package domain

import (
	"cmp"

	"github.com/usetero/cli/internal/format"
)

// CategoryType drives which tab owns a category.
type CategoryType string

const (
	CategoryTypeCompliance CategoryType = "compliance" // Legal/security risk
	CategoryTypeWaste      CategoryType = "waste"      // Event-level cuts
	CategoryTypeQuality    CategoryType = "quality"    // Field-level improvements
)

// PolicyAction describes what a policy does to log events.
type PolicyAction string

const (
	PolicyActionDrop   PolicyAction = "drop"   // Drops entire events
	PolicyActionSample PolicyAction = "sample" // Keeps a fraction of events
	PolicyActionFilter PolicyAction = "filter" // Drops a subset by field value
	PolicyActionTrim   PolicyAction = "trim"   // Removes or truncates fields
	PolicyActionNone   PolicyAction = "none"   // Informational only
)

// PolicyCategoryStatus is the per-category policy breakdown.
// Float pointers are nil when no data exists (e.g. no pending/approved policies contribute).
type PolicyCategoryStatus struct {
	Category    string
	DisplayName string // Human-readable name from control plane (e.g. "Bot Traffic")
	Principle   string // One-liner explaining what this category catches

	PendingCount   int64
	ApprovedCount  int64
	DismissedCount int64

	// Estimated impact from pending policies.
	EstimatedVolumePerHour *float64
	EstimatedBytesPerHour  *float64
	EstimatedCostPerHour   *float64

	// Volume discovery coverage.
	EventsWithVolumes int64 // Log events in this category that have volume data
	TotalEvents       int64 // Total log events in this category

	// What policies in this category do.
	Action PolicyAction
}

// PolicyCategoryStatusByCostDesc sorts by estimated cost descending (highest impact first),
// with nil costs last. Ties break by pending count descending.
func PolicyCategoryStatusByCostDesc(a, b PolicyCategoryStatus) int {
	ac, bc := a.EstimatedCostPerHour, b.EstimatedCostPerHour
	switch {
	case ac == nil && bc == nil:
		return cmp.Compare(b.PendingCount, a.PendingCount)
	case ac == nil:
		return 1 // a after b
	case bc == nil:
		return -1 // a before b
	default:
		if c := cmp.Compare(*bc, *ac); c != 0 {
			return c
		}
		return cmp.Compare(b.PendingCount, a.PendingCount)
	}
}

// Name returns the human-readable name for the category.
// Uses the display name from the control plane, falling back to title-cased slug.
func (c PolicyCategoryStatus) Name() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	return format.TitleCase(c.Category)
}

// ReducesVolume reports whether this category's policies drop entire events.
func (c PolicyCategoryStatus) ReducesVolume() bool {
	return c.EstimatedVolumePerHour != nil && *c.EstimatedVolumePerHour > 0
}

// ActionLabel returns a human-readable description of what this category's policies do.
func (c PolicyCategoryStatus) ActionLabel() string {
	switch c.Action {
	case PolicyActionDrop:
		return "drops events"
	case PolicyActionSample:
		return "samples events"
	case PolicyActionFilter:
		return "filters events"
	case PolicyActionTrim:
		return "trims fields"
	default:
		return ""
	}
}
