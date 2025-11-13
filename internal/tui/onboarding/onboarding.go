package onboarding

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/tui/layouts"
	authcheck "github.com/usetero/cli/internal/tui/onboarding/auth"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// PreferencesReader reads onboarding completion state from preferences
type PreferencesReader interface {
	GetDefaultOrgID() string
	GetDefaultAccountID() string
}

// Onboarding orchestrates the onboarding flow.
// It manages the step-by-step progression through authentication,
// role selection, organization setup, account setup, and datadog integration.
// When complete, it exposes the final state (orgID, accountID) for app creation.
type Onboarding struct {
	// Services
	preferencesService PreferencesReader

	// Flow and layout management
	flow   *step.Flow
	layout layouts.Layout

	// State
	ready          bool
	orgID          string // Set when onboarding completes
	accountID      string // Set when onboarding completes
	globalBindings []key.Binding
	logger         log.Logger
}

// New creates a new onboarding model starting with auth
func New(
	logger log.Logger,
	authService *auth.Service,
	preferencesService *preferences.Service,
	apiEndpoint string,
	globalBindings []key.Binding,
) *Onboarding {
	// Start onboarding flow with auth check step
	// Check step validates existing auth, or proceeds to auth step if needed
	flow := step.NewFlow(
		authcheck.NewCheckAuthStep(authService, authService, preferencesService, apiEndpoint, logger, globalBindings),
	)

	return &Onboarding{
		flow:               flow,
		layout:             layouts.NewHeader(logger),
		ready:              false,
		logger:             logger,
		preferencesService: preferencesService,
		globalBindings:     globalBindings,
	}
}

// Init initializes the onboarding flow
func (m *Onboarding) Init() tea.Cmd {
	return m.flow.Init()
}

// Update handles messages and delegates to the flow
func (m *Onboarding) Update(msg tea.Msg) tea.Cmd {
	// Cascade to flow first so error state is up to date
	flowCmd := m.flow.Update(msg)

	// Combine flow bindings + global bindings for layout
	var bindings []key.Binding
	bindings = append(bindings, m.flow.Help().ShortHelp()...)
	bindings = append(bindings, m.globalBindings...)
	m.layout.SetKeyBindings(bindings)

	// Pass error state to layout (always set, even if nil to clear previous errors)
	m.layout.SetError(m.flow.Error())

	// Cascade to layout
	layoutCmd := m.layout.Update(msg)

	// Check if flow completed and extract org/account IDs
	if m.flow.IsComplete() && m.orgID == "" {
		// Flow completed - extract final state from preferences
		// The onboarding flow sets these in preferences as it progresses
		m.orgID = m.preferencesService.GetDefaultOrgID()
		m.accountID = m.preferencesService.GetDefaultAccountID()
		m.logger.Info("onboarding completed", "orgID", m.orgID, "accountID", m.accountID)
	}

	return tea.Batch(layoutCmd, flowCmd)
}

// View renders the onboarding header + current step content
func (m *Onboarding) View() string {
	if !m.ready {
		return ""
	}

	// Ask layout for available content dimensions
	contentWidth, contentHeight := m.layout.ContentSize()

	// Get current step content
	stepContent := m.flow.View()

	// Bottom-align the step content in available space
	content := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		AlignVertical(lipgloss.Bottom).
		Render(stepContent)

	// Layout handles header + content + footer composition
	return m.layout.Render(content)
}

// SetSize sets dimensions and propagates to layout and flow
func (m *Onboarding) SetSize(width, height int) {
	m.layout.SetSize(width, height)
	m.flow.SetSize(width, height)
	m.ready = true
}

// IsComplete returns true when onboarding flow is complete
func (m *Onboarding) IsComplete() bool {
	return m.flow.IsComplete()
}

// IsBusy delegates to the current step in the flow
func (m *Onboarding) IsBusy() bool {
	return m.flow.IsBusy()
}

// HasError returns true if the current step has an error
func (m *Onboarding) HasError() bool {
	return m.flow.HasError()
}

// Error returns the current step's error, or nil if no error
func (m *Onboarding) Error() error {
	return m.flow.Error()
}

// OrganizationID returns the organization ID from completed onboarding
// Only valid after IsComplete() returns true
func (m *Onboarding) OrganizationID() string {
	return m.orgID
}

// AccountID returns the account ID from completed onboarding
// Only valid after IsComplete() returns true
func (m *Onboarding) AccountID() string {
	return m.accountID
}
