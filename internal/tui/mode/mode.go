package mode

import tea "github.com/charmbracelet/bubbletea/v2"

// Mode represents a distinct mode of the TUI (onboarding, app, etc.)
// Each mode is self-contained and manages its own state and UI.
// TUI orchestrates transitions between modes when they complete.
type Mode interface {
	// Init initializes the mode
	Init() tea.Cmd

	// Update handles messages and returns a command
	Update(tea.Msg) tea.Cmd

	// View returns the string representation of the mode's UI
	View() string

	// SetSize sets the dimensions available for rendering
	SetSize(width, height int)

	// IsComplete returns true when the mode has finished its work
	IsComplete() bool

	// IsBusy returns true if the mode is performing background work
	IsBusy() bool

	// HasError returns true if the mode is in an error state
	HasError() bool

	// Error returns the current error, or nil if no error
	Error() error
}
