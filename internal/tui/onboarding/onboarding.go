package onboarding

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/layouts"
	authcheck "github.com/usetero/cli/internal/tui/onboarding/auth"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/onboarding/sync"
)

// Onboarding orchestrates the onboarding flow.
// It manages the step-by-step progression through authentication,
// role selection, organization setup, account setup, and datadog integration.
// When complete, it exposes the final state (org, account) for app creation.
type Onboarding struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Flow and layout management
	flow   *step.Flow
	layout layouts.Layout
	theme  *styles.Theme

	// State
	ready          bool
	org            api.Organization // Set when onboarding completes
	account        api.Account      // Set when onboarding completes
	globalBindings []key.Binding
	logger         log.Logger
}

// New creates a new onboarding model starting with auth
func New(
	ctx context.Context,
	theme *styles.Theme,
	logger log.Logger,
	authService auth.Auth,
	prefs preferences.Preferences,
	apiEndpoint string,
	globalBindings []key.Binding,
) *Onboarding {
	// Start onboarding flow with auth check step
	// Check step validates existing auth, or proceeds to auth step if needed
	flow := step.NewFlow(
		authcheck.NewCheckAuthStep(ctx, theme, authService, prefs, apiEndpoint, logger, globalBindings),
		logger,
	)

	return &Onboarding{
		ctx:            ctx,
		flow:           flow,
		layout:         layouts.NewHeader(theme, logger),
		theme:          theme,
		ready:          false,
		logger:         logger,
		globalBindings: globalBindings,
	}
}

// Init initializes the onboarding flow
func (m *Onboarding) Init() tea.Cmd {
	return m.flow.Init()
}

// Update handles messages and delegates to the flow
func (m *Onboarding) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Cascade to flow
	flowCmd := m.flow.Update(msg)
	cmds = append(cmds, flowCmd)

	// Combine flow bindings + global bindings for layout
	var bindings []key.Binding
	bindings = append(bindings, m.flow.Help().ShortHelp()...)
	bindings = append(bindings, m.globalBindings...)
	m.layout.SetKeyBindings(bindings)

	// Pass error state to layout (always set, even if nil to clear previous errors)
	m.layout.SetError(m.flow.Error())

	// Cascade to layout
	layoutCmd := m.layout.Update(msg)
	cmds = append(cmds, layoutCmd)

	// Check if flow completed and extract org/account from final step
	if m.flow.IsComplete() && m.org.ID == "" {
		// Flow completed - extract final state from the last step
		if lastStep, ok := m.flow.LastStep().(*sync.Step); ok {
			m.org = lastStep.Organization()
			m.account = lastStep.Account()
			m.logger.Info("onboarding completed", "orgID", m.org.ID, "accountID", m.account.ID)
		}
	}

	return tea.Batch(cmds...)
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

// Organization returns the organization from completed onboarding
// Only valid after IsComplete() returns true
func (m *Onboarding) Organization() api.Organization {
	return m.org
}

// Account returns the account from completed onboarding
// Only valid after IsComplete() returns true
func (m *Onboarding) Account() api.Account {
	return m.account
}

// Close releases any resources held by onboarding.
func (m *Onboarding) Close() error {
	return m.flow.Close()
}
