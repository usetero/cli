package page

import "image/color"

// Metadata represents a piece of information a page wants to display
// in the sidebar or header. Higher priority items are shown first
// and are more likely to appear in compact/header mode.
type Metadata struct {
	// Label is the display name (e.g., "Waste", "Services", "Session")
	Label string

	// Value is the display value (e.g., "34%", "142", "2m ago")
	Value string

	// Priority determines display order (lower = higher priority)
	// Priority 1 items are always shown, even in compact header
	Priority int

	// Color is an optional color for the value (e.g., red for high waste)
	// Use nil for default color
	Color color.Color

	// Icon is an optional icon/symbol to display before the label
	Icon string
}
