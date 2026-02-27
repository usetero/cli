// Package datadog provides Datadog configuration steps.
package datadog

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// checkResultMsg is sent when datadog check completes.
type checkResultMsg struct {
	hasDatadog bool
	err        error
}

// CheckModel checks if Datadog is configured for the account.
type CheckModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	err      error
}

// NewCheck creates a new datadog check step.
func NewCheck(
	ctx context.Context,
	theme styles.Theme,
	account domain.Account,
	services api.APIServices,
	scope log.Scope,
) *CheckModel {
	if ctx == nil {
		panic("ctx is nil")
	}

	return &CheckModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
	}
}

// Init starts checking for Datadog configuration.
func (m *CheckModel) Init() tea.Cmd {
	m.scope.Info("checking datadog configuration", "accountID", m.account.ID)
	return m.checkDatadog()
}

func (m *CheckModel) checkDatadog() tea.Cmd {
	return func() tea.Msg {
		hasDatadog, err := m.services.DatadogAccounts.HasAccount(m.ctx, m.account.ID)
		if err != nil {
			return checkResultMsg{err: err}
		}
		return checkResultMsg{hasDatadog: hasDatadog}
	}
}

// Update handles messages.
func (m *CheckModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case checkResultMsg:
		if msg.err != nil {
			m.scope.Error("datadog check failed", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to check Datadog status", msg.err, false)
		}
		if msg.hasDatadog {
			m.scope.Info("datadog configured")
			return func() tea.Msg { return msgs.DatadogReady{} }
		}
		m.scope.Info("datadog setup required")
		return func() tea.Msg { return msgs.DatadogNeeded{} }

	case tea.KeyPressMsg:
		if msg.String() == "r" && m.err != nil {
			m.scope.Debug("retrying datadog check")
			m.err = nil
			return m.checkDatadog()
		}
	}
	return nil
}

// View renders the check UI.
func (m *CheckModel) View() string {
	s := m.theme.Styles

	if m.err != nil {
		return s.Error.Render(errorfmt.UserFacing(m.err, "Failed to check Datadog configuration. Press 'r' to retry."))
	}
	return s.Title.Render("Checking Datadog configuration...")
}

// SetSize updates dimensions.
func (m *CheckModel) SetSize(width, height int) {}

// ShortHelp returns the key bindings for the short help view.
func (m *CheckModel) ShortHelp() []key.Binding {
	if m.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	return nil
}

func (m *CheckModel) Hidden() bool {
	return m.err == nil
}

func (m *CheckModel) StatusText() string {
	return "Checking Datadog configuration..."
}
