package accounts

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// accountCreatedMsg is sent when account creation completes.
type accountCreatedMsg struct {
	account domain.Account
	err     error
}

// CreateModel handles account creation.
type CreateModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger
	org      domain.Organization

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewCreate creates a new account creation step.
func NewCreate(
	ctx context.Context,
	theme *styles.Theme,
	org domain.Organization,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) *CreateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Account name (e.g., Production, Staging)")
	inp.SetCharLimit(100)

	return &CreateModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		logger:   logger,
		org:      org,
		input:    inp,
	}
}

// Init focuses the input.
func (m *CreateModel) Init() tea.Cmd {
	return m.input.Focus()
}

// Update handles messages.
func (m *CreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.logger.Error("failed to create account", "error", msg.err)
			m.err = msg.err
			return nil
		}
		_ = m.prefs.SetDefaultAccountID(msg.account.ID)
		m.services.SetAccountID(msg.account.ID)
		org := m.org
		acc := msg.account
		m.logger.Info("account created", "id", acc.ID, "name", acc.Name)
		return func() tea.Msg { return msgs.AccountCreated{Org: org, Account: acc} }

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		switch msg.String() {
		case "enter":
			name := m.input.Value()
			if name == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.logger.Info("creating account", "name", name)
			return m.createAccount(name)
		}
	}

	return m.input.Update(msg)
}

func (m *CreateModel) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		account, err := m.services.Accounts.Create(m.ctx, id, m.org.ID, name)
		if err != nil {
			return accountCreatedMsg{err: err}
		}
		return accountCreatedMsg{account: *account}
	}
}

// View renders the account creation UI.
func (m *CreateModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Create a Datadog account")
	subtitle := s.Help.Render("Accounts connect to a Datadog organization")

	var status string
	if m.creating {
		status = s.Help.Render("Creating...")
	} else if m.err != nil {
		status = s.Error.Render("Failed to create account. Try again.")
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *CreateModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}
