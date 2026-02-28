package onboarding

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
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

// StatusProvider is an optional capability for structured stage status text.
type StatusProvider interface {
	Status() appmsg.StepStatus
}
