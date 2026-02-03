package accountcreate

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	datadogcheck "github.com/usetero/cli/internal/tui/onboarding/datadog/check"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// createAccountMsg is sent when account creation completes.
type createAccountMsg struct {
	account *domain.Account
	err     error
}

// Model handles creating a new account (Datadog connection).
type Model struct {
	ctx   context.Context
	theme *styles.Theme
	role  string
	org   domain.Organization

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	input          input.Model
	creating       bool
	created        bool
	createdAccount *domain.Account
	err            error
	width          int
	height         int
}

// New creates a new account create model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	inp := input.New(theme, logger).
		SetPlaceholder("Production Datadog").
		SetCharLimit(100)

	return Model{
		ctx:      ctx,
		theme:    theme,
		role:     role,
		org:      org,
		services: services,
		prefs:    prefs,
		logger:   logger,
		input:    inp,
		width:    80,
	}
}

// Init initializes the step.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.err != nil {
				m.err = nil
				m.creating = false
				return m, cmd
			}

			if !m.creating && !m.created {
				name := m.input.Value()
				if name != "" {
					m.creating = true
					return m, tea.Batch(cmd, m.createAccount(name))
				}
			}
		}

	case createAccountMsg:
		m.creating = false
		if msg.err != nil {
			m.logger.Error("failed to create account", "error", msg.err)
			m.err = msg.err
			return m, cmd
		}

		m.logger.Info("account created", "id", msg.account.ID, "name", msg.account.Name)
		m.createdAccount = msg.account

		if err := m.prefs.SetDefaultAccountID(msg.account.ID); err != nil {
			m.logger.Error("failed to save account preference", "error", err)
			m.err = err
			return m, cmd
		}

		m.created = true
		return m, cmd
	}

	return m, cmd
}

// createAccount returns a command that creates the account.
func (m Model) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("creating account", "name", name, "organizationID", m.org.ID)

		account, err := m.services.Accounts.Create(m.ctx, uuid.New(), m.org.ID, name)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: account}
	}
}

// View renders the create account UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.creating {
		return themeStyles.Title.Render("Creating Datadog account...")
	}

	title := themeStyles.Title.Render("Connect a Datadog account")
	prompt := themeStyles.Body.Render("Enter a name for this Datadog connection")
	help := themeStyles.Help.Render("You can connect multiple Datadog accounts later")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		"",
		m.input.View(),
		"",
		help,
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	if width > 10 {
		m.input = m.input.SetWidth(width - 6)
	}
	return m
}

// IsBusy returns true while creating.
func (m Model) IsBusy() bool {
	return m.creating
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m Model) Error() error {
	return m.err
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.created {
		return nil, step.ErrNotReady
	}

	return datadogcheck.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		*m.createdAccount,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
