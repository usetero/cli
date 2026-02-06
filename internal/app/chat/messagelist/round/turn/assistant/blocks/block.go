package blocks

import tea "charm.land/bubbletea/v2"

// Block is the interface for all content blocks in an assistant message.
// Blocks are fixed-height components - their height is determined by content.
// Parent sets width via SetWidth, then asks Height() to know space needed.
type Block interface {
	// Index returns the block's position in the content array.
	Index() int

	// Update handles messages.
	Update(tea.Msg) tea.Cmd

	// View renders the block at the width set by SetWidth.
	View() string

	// SetWidth sets the available width for rendering.
	SetWidth(int)

	// Height returns the number of lines this block renders.
	Height() int
}
