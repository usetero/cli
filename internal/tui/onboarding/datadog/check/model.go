package datadogcheck

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	datadogdiscovery "github.com/usetero/cli/internal/tui/onboarding/datadog/discovery"
	datadogselectregion "github.com/usetero/cli/internal/tui/onboarding/datadog/selectregion"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// checkDatadogMsg is sent when check completes.
type checkDatadogMsg struct {
	hasDatadog     bool
	datadogAccount *api.DatadogAccount
	err            error
}

// Model checks if the account has a Datadog integration configured.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	role    string
	org     domain.Organization
	account domain.Account

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	checking       bool
	checked        bool
	hasDatadog     bool
	datadogAccount *api.DatadogAccount
	err            error
	width          int
	height         int
}

// New creates a new datadog check model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	account domain.Account,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	return Model{
		ctx:      ctx,
		theme:    theme,
		role:     role,
		org:      org,
		account:  account,
		services: services,
		prefs:    prefs,
		logger:   logger,
		width:    80,
	}
}

// Init starts checking for Datadog account.
func (m Model) Init() tea.Cmd {
	m.checking = true
	return m.checkDatadogAccount()
}

// checkDatadogAccount checks if account has Datadog configured.
func (m Model) checkDatadogAccount() tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("checking datadog account", "accountID", m.account.ID)

		hasDatadog, err := m.services.DatadogAccounts.HasAccount(m.ctx, m.account.ID.String())
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		if !hasDatadog {
			return checkDatadogMsg{hasDatadog: false}
		}

		datadogAccount, err := m.services.DatadogAccounts.GetAccount(m.ctx, m.account.ID.String())
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		return checkDatadogMsg{hasDatadog: true, datadogAccount: datadogAccount}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case checkDatadogMsg:
		m.checking = false
		if msg.err != nil {
			m.logger.Error("failed to check datadog account", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

		m.err = nil
		if msg.hasDatadog {
			m.logger.Info("datadog account found")
			m.datadogAccount = msg.datadogAccount
		} else {
			m.logger.Info("no datadog account found")
		}
		m.checked = true
		m.hasDatadog = msg.hasDatadog
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.err != nil {
				m.logger.Info("retrying datadog account check")
				m.err = nil
				m.checking = true
				return m, m.checkDatadogAccount()
			}
		}
	}

	return m, nil
}

// View renders the check UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.checking {
		return themeStyles.Title.Render("Checking Datadog account...")
	}

	if m.hasDatadog {
		return themeStyles.Success.Render("Datadog account found")
	}

	return themeStyles.Title.Render("No Datadog account found")
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns true while checking.
func (m Model) IsBusy() bool {
	return m.checking
}

// HasError returns true if there was an error.
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
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}
	return keymap.Simple{Keys: []key.Binding{}}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.checked {
		return nil, step.ErrNotReady
	}

	if !m.hasDatadog {
		// No Datadog account - go to select region step
		return datadogselectregion.New(
			m.ctx,
			m.theme,
			m.role,
			m.org,
			m.account,
			m.services,
			m.prefs,
			m.logger,
		), nil
	}

	// Datadog account exists - go to discovery step
	return datadogdiscovery.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		m.account,
		&m.datadogAccount.ID,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
