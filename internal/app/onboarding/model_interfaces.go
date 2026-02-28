package onboarding

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// StepCore is the local contract for onboarding step components.
type StepCore interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
	SetSize(width, height int)
}

// HelpProvider is an optional capability for short help bindings.
type HelpProvider interface {
	ShortHelp() []key.Binding
}
