package datadog

import (
	"context"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// apiKeyValidatedMsg is sent when API key validation completes.
type apiKeyValidatedMsg struct {
	valid    bool
	errorMsg string
	err      error
}

// APIKeyModel handles Datadog API key entry.
type APIKeyModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	site     domain.DatadogSite

	input      *input.Model
	validating bool
	err        error
	width      int
	height     int
}

// NewAPIKey creates a new API key entry step.
func NewAPIKey(
	ctx context.Context,
	theme *styles.Theme,
	account domain.Account,
	site domain.DatadogSite,
	services api.APIServices,
	scope log.Scope,
) *APIKeyModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if site == "" {
		panic("site is empty")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Datadog API Key")
	inp.SetCharLimit(64)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &APIKeyModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
		site:     site,
		input:    inp,
	}
}

// Init focuses the input.
func (m *APIKeyModel) Init() tea.Cmd {
	return m.input.Focus()
}

// Update handles messages.
func (m *APIKeyModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case apiKeyValidatedMsg:
		m.validating = false
		if msg.err != nil {
			m.scope.Error("api key validation failed", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to validate API key", msg.err, false)
		}
		if !msg.valid {
			m.scope.Info("api key invalid", "reason", msg.errorMsg)
			m.err = &validationError{msg.errorMsg}
			return nil
		}
		m.scope.Info("api key validated")
		apiKey := m.input.Value()
		return func() tea.Msg {
			return msgs.DatadogAPIKeyEntered{APIKey: apiKey}
		}

	case tea.KeyPressMsg:
		if m.validating {
			return nil
		}
		switch msg.String() {
		case "enter":
			apiKey := m.input.Value()
			if apiKey == "" {
				return nil
			}
			m.validating = true
			m.err = nil
			m.scope.Info("validating api key")
			return m.validateAPIKey(apiKey)
		}
	}

	return m.input.Update(msg)
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }

func (m *APIKeyModel) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		valid, errorMsg, err := m.services.DatadogAccounts.ValidateAPIKey(m.ctx, apiKey, m.site.String())
		if err != nil {
			return apiKeyValidatedMsg{err: err}
		}
		return apiKeyValidatedMsg{valid: valid, errorMsg: errorMsg}
	}
}

// View renders the API key entry UI.
func (m *APIKeyModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Enter your Datadog API Key")
	subtitle := s.Help.Render("You can find this in Datadog under Organization Settings → API Keys")

	var status string
	if m.validating {
		status = s.Help.Render("Validating...")
	} else if m.err != nil {
		status = s.Error.Render(m.err.Error())
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *APIKeyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}
