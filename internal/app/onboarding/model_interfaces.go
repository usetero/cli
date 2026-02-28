package onboarding

import (
	appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Step is the onboarding step contract.
type Step interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
	SetSize(width, height int)
	ShortHelp() []key.Binding
	Hidden() bool
	Status() appmsg.StepStatus
}
