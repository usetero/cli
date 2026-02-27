// Package onboarding provides the onboarding flow for new users.
package onboarding

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	iauth "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// Model is the onboarding orchestrator.
type Model struct {
	// Dependencies available from start
	ctx       context.Context
	theme     styles.Theme
	services  api.APIServices
	userPrefs preferences.UserPreferences
	orgPrefs  preferences.OrgPreferences
	auth      iauth.Auth
	syncer    powersync.Syncer
	scope     log.Scope

	// Accumulated state from step completions
	state engineState

	// Current step
	gate   Gate
	step   Step
	width  int
	height int
}

// New creates a new onboarding model.
func New(
	ctx context.Context,
	theme styles.Theme,
	services api.APIServices,
	userPrefs preferences.UserPreferences,
	orgPrefs preferences.OrgPreferences,
	authService iauth.Auth,
	syncer powersync.Syncer,
	scope log.Scope,
) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	if userPrefs == nil {
		panic("userPrefs is nil")
	}
	if orgPrefs == nil {
		panic("orgPrefs is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}

	scope = scope.Child("onboarding")

	return &Model{
		ctx:       ctx,
		theme:     theme,
		services:  services,
		userPrefs: userPrefs,
		orgPrefs:  orgPrefs,
		auth:      authService,
		syncer:    syncer,
		scope:     scope,
	}
}

// SetOrgPreferences replaces the org preferences used by subsequent onboarding steps.
// Called by the app when switching to a different organization.
func (m *Model) SetOrgPreferences(prefs preferences.OrgPreferences) {
	m.orgPrefs = prefs
}

// Init starts the onboarding flow with preflight resolution.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("onboarding started")
	return m.goToGate(GatePreflight)
}

// StartFromOrgSelect starts the onboarding flow at the organization selection
// step, skipping role selection. Used when switching orgs or
// accounts — the caller clears the relevant preferences so onboarding prompts
// for selection instead of auto-selecting.
func (m *Model) StartFromOrgSelect() tea.Cmd {
	m.scope.Info("onboarding started from org select")
	return m.goToGate(GateOrgSelect)
}

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step != nil {
			m.step.SetSize(m.width, m.height)
		}
		return nil
	}

	if cmd := m.handleTransition(msg); cmd != nil {
		return cmd
	}

	// Delegate to current step
	if m.step != nil {
		return m.step.Update(msg)
	}
	return nil
}

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	view := m.step.View()
	if v, ok := m.step.(VisibilityProvider); ok && v.Hidden() {
		view = m.hiddenStepView(v.StatusText())
	}

	// Bottom-align step content in available space
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(view)
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.step != nil {
		m.step.SetSize(width, height)
	}
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.step != nil {
		return m.step.ShortHelp()
	}
	return nil
}

func (m *Model) hiddenStepView(status string) string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")
	body := s.Body.Render(status)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}
