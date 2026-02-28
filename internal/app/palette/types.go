// Package palette provides a command palette overlay for the TUI.
// Triggered by "/" in the input bar, it shows a fuzzy-filtered list of commands.
package palette

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Command is a single palette entry.
// If Children is set, selecting this command drills into the sub-commands
// instead of executing Handler.
type Command struct {
	Name     string         // Display name (e.g. "New Conversation")
	Handler  func() tea.Cmd // What to do when selected (leaf commands)
	Children []Command      // Sub-commands (selecting drills down)
}

// OpenMsg requests the palette to open.
type OpenMsg struct{}

// CloseMsg is sent when the palette closes itself.
type CloseMsg struct{}

// Key bindings.
var (
	selectKey = key.NewBinding(key.WithKeys("enter"))
	closeKey  = key.NewBinding(key.WithKeys("esc"))
	nextKey   = key.NewBinding(key.WithKeys("down", "ctrl+n"))
	prevKey   = key.NewBinding(key.WithKeys("up", "ctrl+p"))
)

const (
	maxVisible   = 8 // max items shown
	headerHeight = 1
	inputHeight  = 1
	paddingH     = 1
	diag         = "╱"
)

// level captures the state of a palette level for back navigation.
type level struct {
	commands []Command
	title    string
}

// match pairs a command with its fuzzy match positions.
type match struct {
	command      Command
	matchIndexes []int
}
