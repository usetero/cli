package screens

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding"
	"github.com/usetero/cli/internal/interfaces/tui/screens/understanding"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
)

// Model owns top-level body routing.
type Model struct {
	core.Router

	onboarding    *onboarding.Model
	understanding *understanding.Model
}

var _ core.Model = (*Model)(nil)
var _ core.BusyProvider = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)
var _ core.HelpProvider = (*Model)(nil)

func New(scope logging.Scope, identityService *identity.Service, workflow *runtimeonboarding.Workflow, appTheme theme.Theme) *Model {
	onboardingModel := onboarding.New(scope.Child("onboarding"), identityService, workflow, appTheme)
	m := &Model{
		Router:        core.Router{},
		onboarding:    onboardingModel,
		understanding: understanding.New(appTheme, nil),
	}
	m.showOnboarding()
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case events.OnboardingCompletedMsg:
		m.showUnderstanding()
		return m, nil
	}
	return m, m.Router.Update(msg)
}

func (m *Model) showOnboarding() {
	m.Router.SetActive(m.onboarding)
}

func (m *Model) showUnderstanding() {
	m.Router.SetActive(m.understanding)
}
