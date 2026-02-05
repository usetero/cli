// Package datadog provides Datadog configuration steps.
package datadog

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/api"
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
	theme    *styles.Theme
	services api.APIServices
	logger   log.Logger
	account  domain.Account
	err      error
}

// NewCheck creates a new datadog check step.
func NewCheck(
	ctx context.Context,
	theme *styles.Theme,
	account domain.Account,
	services api.APIServices,
	logger log.Logger,
) *CheckModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}
	return &CheckModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		logger:   logger,
		account:  account,
	}
}

// Init starts checking for Datadog configuration.
func (m *CheckModel) Init() tea.Cmd {
	m.logger.Info("checking datadog configuration", "accountID", m.account.ID)
	return m.checkDatadog()
}

func (m *CheckModel) checkDatadog() tea.Cmd {
	return func() tea.Msg {
		hasDatadog, err := m.services.DatadogAccounts.HasAccount(m.ctx, m.account.ID.String())
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
			m.logger.Error("datadog check failed", "error", msg.err)
			m.err = msg.err
			return nil
		}
		if msg.hasDatadog {
			m.logger.Info("datadog configured")
			return func() tea.Msg { return msgs.DatadogReady{} }
		}
		m.logger.Info("datadog setup required")
		return func() tea.Msg { return msgs.DatadogNeeded{} }

	case tea.KeyPressMsg:
		if msg.String() == "r" && m.err != nil {
			m.logger.Debug("retrying datadog check")
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
		return s.Error.Render("Failed to check Datadog configuration. Press 'r' to retry.")
	}
	return s.Title.Render("Checking Datadog configuration...")
}

// SetSize updates dimensions.
func (m *CheckModel) SetSize(width, height int) {}
