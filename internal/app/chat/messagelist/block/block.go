// Package block defines the shared interface for all visual blocks
// in the message list viewport.
package block

import tea "charm.land/bubbletea/v2"

// AssistantPadding is the left padding width for assistant content blocks.
// Shared between the messagelist viewport (which wraps blocks with padding)
// and the assistant model (which sets block widths minus padding).
const AssistantPadding = 2

// Kind identifies the type of a block for layout and styling decisions.
type Kind int

const (
	KindUser Kind = iota
	KindAssistantText
	KindThinking
	KindTool
	KindThinkingAnimation
)

// Block is the interface for all visual atoms in the message list.
// Each block is an independently focusable, renderable unit.
// Width is set via SetWidth before View is called.
type Block interface {
	// View renders the block at the width set by SetWidth.
	View() string

	// Height returns the number of lines this block renders.
	Height() int

	// Update handles messages.
	Update(tea.Msg) tea.Cmd

	// SetWidth sets the available width for rendering.
	SetWidth(int)

	// SetFocused sets whether this block is the focused block in the viewport.
	SetFocused(bool)

	// Focused returns whether this block is currently focused.
	Focused() bool

	// Kind returns the block type.
	Kind() Kind
}

// Toggleable is an optional interface for blocks that can be toggled
// (e.g. expand/collapse thinking blocks, show/hide tool body).
type Toggleable interface {
	Toggle()
}
