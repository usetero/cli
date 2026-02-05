// Package auth provides authentication steps for onboarding.
package auth

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// checkResultMsg is the internal message for auth check completion.
type checkResultMsg struct {
	hasValidAuth bool
	err          error
}

// CheckModel checks if the user has a valid auth token.
type CheckModel struct {
	ctx    context.Context
	theme  *styles.Theme
	auth   auth.Auth
	logger log.Logger
	err    error
}

// NewCheck creates a new auth check step.
func NewCheck(ctx context.Context, theme *styles.Theme, authService auth.Auth, logger log.Logger) *CheckModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}
	return &CheckModel{
		ctx:    ctx,
		theme:  theme,
		auth:   authService,
		logger: logger,
	}
}

// Init starts the auth check.
func (m *CheckModel) Init() tea.Cmd {
	m.logger.Info("checking authentication")
	return m.checkAuth()
}

func (m *CheckModel) checkAuth() tea.Cmd {
	return func() tea.Msg {
		if !m.auth.IsAuthenticated() {
			return checkResultMsg{hasValidAuth: false}
		}

		_, err := m.auth.GetAccessToken(m.ctx)
		if err != nil {
			_ = m.auth.ClearTokens()
			return checkResultMsg{hasValidAuth: false}
		}

		return checkResultMsg{hasValidAuth: true}
	}
}

// Update handles messages.
func (m *CheckModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case checkResultMsg:
		if msg.err != nil {
			m.logger.Error("auth check failed", "error", msg.err)
			m.err = msg.err
			return nil
		}
		if msg.hasValidAuth {
			m.logger.Info("auth valid")
		} else {
			m.logger.Info("auth required")
		}
		return func() tea.Msg {
			return msgs.AuthChecked{NeedsAuth: !msg.hasValidAuth}
		}
	}
	return nil
}

// View renders the check UI.
func (m *CheckModel) View() string {
	if m.err != nil {
		return m.theme.Styles.Error.Render("Error checking authentication")
	}
	return m.theme.Styles.Title.Render("Checking authentication...")
}

// SetSize updates dimensions.
func (m *CheckModel) SetSize(width, height int) {}
