package messages

import tea "charm.land/bubbletea/v2"

// Item is a single renderable item in the message list.
// User messages, assistant messages, and tool blocks all implement this.
type Item interface {
	// ID returns a unique identifier for this item.
	ID() string

	// Update handles Bubble Tea messages when this item is focused.
	Update(msg tea.Msg) (Item, tea.Cmd)

	// Render returns the rendered content at the given width.
	// The returned string may be multiple lines.
	Render(width int) string

	// Height returns the number of lines this item renders at the given width.
	Height(width int) int

	// SetFocused updates the focus state. Focused items may render differently
	// (e.g., highlighted border) and receive key events.
	SetFocused(focused bool) Item
}

// Expandable is implemented by items that can expand/collapse (e.g., tool output).
type Expandable interface {
	Item
	ToggleExpanded() Item
	IsExpanded() bool
}

// Copyable is implemented by items that support copying content to clipboard.
type Copyable interface {
	Item
	CopyableContent() string
}

// Highlightable is implemented by items that support text selection.
// Items that implement this interface can have a portion of their content
// highlighted (e.g., for mouse text selection).
type Highlightable interface {
	Item
	// SetHighlight sets the highlight range within the item.
	// Use -1 values to clear the highlight.
	SetHighlight(startLine, startCol, endLine, endCol int)
	// Highlight returns the current highlight range.
	Highlight() (startLine, startCol, endLine, endCol int)
	// ClearHighlight removes any highlight.
	ClearHighlight()
	// IsHighlighted returns true if the item has a highlight set.
	IsHighlighted() bool
	// HighlightedContent returns the plain text content of the highlighted region.
	HighlightedContent() string
}
