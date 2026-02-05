package datadog

import (
	"context"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// accountCreatedMsg is sent when account creation completes.
type accountCreatedMsg struct {
	datadogAccountID domain.DatadogAccountID
	err              error
}

// AppKeyModel handles Datadog Application key entry and account creation.
type AppKeyModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	site     domain.DatadogSite
	apiKey   string

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewAppKey creates a new App key entry step.
func NewAppKey(
	ctx context.Context,
	theme *styles.Theme,
	account domain.Account,
	site domain.DatadogSite,
	apiKey string,
	services api.APIServices,
	scope log.Scope,
) *AppKeyModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if site == "" {
		panic("site is empty")
	}
	if apiKey == "" {
		panic("apiKey is empty")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Datadog Application Key")
	inp.SetCharLimit(64)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &AppKeyModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
		site:     site,
		apiKey:   apiKey,
		input:    inp,
	}
}

// Init focuses the input.
func (m *AppKeyModel) Init() tea.Cmd {
	return m.input.Focus()
}

// Update handles messages.
func (m *AppKeyModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.scope.Error("failed to create datadog account", "error", msg.err)
			m.err = msg.err
			return nil
		}
		m.scope.Info("datadog account created", "id", msg.datadogAccountID)
		ddAccountID := msg.datadogAccountID
		return func() tea.Msg {
			return msgs.DatadogAccountCreated{DatadogAccountID: ddAccountID}
		}

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		switch msg.String() {
		case "enter":
			appKey := m.input.Value()
			if appKey == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.scope.Info("creating datadog account")
			return m.createAccount(appKey)
		}
	}

	return m.input.Update(msg)
}

func (m *AppKeyModel) createAccount(appKey string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		ddAccount, err := m.services.DatadogAccounts.CreateAccount(
			m.ctx,
			id,
			m.account.ID.String(),
			m.account.Name,
			m.site.String(),
			m.apiKey,
			appKey,
		)
		if err != nil {
			return accountCreatedMsg{err: err}
		}
		return accountCreatedMsg{datadogAccountID: domain.DatadogAccountID(ddAccount.ID)}
	}
}

// View renders the App key entry UI.
func (m *AppKeyModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Enter your Datadog Application Key")
	subtitle := s.Help.Render("You can find this in Datadog under Organization Settings → Application Keys")

	var status string
	if m.creating {
		status = s.Help.Render("Connecting to Datadog...")
	} else if m.err != nil {
		status = s.Error.Render("Invalid Application Key. Please try again.")
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *AppKeyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}
