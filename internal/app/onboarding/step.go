package onboarding

import tea "charm.land/bubbletea/v2"

// Step represents a single step in the onboarding flow.
type Step interface {
	// Init returns the initial command for the step.
	Init() tea.Cmd

	// Update handles a message and returns a command.
	// Steps emit completion messages (from msgs package) when done.
	Update(msg tea.Msg) tea.Cmd

	// View renders the step.
	View() string

	// SetSize updates the step's dimensions.
	SetSize(width, height int)
}
