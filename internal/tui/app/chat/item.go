package chat

import tea "charm.land/bubbletea/v2"

// Item is an element in the chat message list.
// Items manage their own state, animation, and rendering.
type Item interface {
	// ID returns a unique identifier for this item.
	ID() string

	// Init initializes the item (e.g., starts animations).
	Init() tea.Cmd

	// Update handles messages and returns commands.
	Update(msg tea.Msg) tea.Cmd

	// View renders the item to a string.
	View() string

	// SetWidth sets the available width for rendering.
	SetWidth(width int)

	// Spinning returns true if the item has an active animation.
	// Used to efficiently route tick messages only to items that need them.
	Spinning() bool
}
