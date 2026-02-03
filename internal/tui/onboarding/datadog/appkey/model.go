package datadogappkey

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/google/uuid"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	datadogdiscovery "github.com/usetero/cli/internal/tui/onboarding/datadog/discovery"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// createAccountMsg is sent when Datadog account creation completes.
type createAccountMsg struct {
	account *api.DatadogAccount
	err     error
}

// Model handles collecting the user's Datadog application key.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	role    string
	org     domain.Organization
	account domain.Account
	site    string // Selected Datadog site
	apiKey  string // Validated API key from previous step

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	input          input.Model
	showingInput   bool // false = interstitial, true = input screen
	creating       bool
	created        bool
	createdAccount *api.DatadogAccount
	err            error
	copiedURL      bool
	width          int
	height         int
}

// New creates a new app key collection model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	account domain.Account,
	site string,
	apiKey string,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	inp := input.New(theme, logger).
		SetPlaceholder("Enter your Datadog application key...").
		SetWidth(50).
		SetEchoMode(textinput.EchoPassword).
		SetEchoCharacter('•')

	return Model{
		ctx:          ctx,
		theme:        theme,
		role:         role,
		org:          org,
		account:      account,
		site:         site,
		apiKey:       apiKey,
		services:     services,
		prefs:        prefs,
		logger:       logger,
		input:        inp,
		showingInput: false,
		width:        80,
	}
}

// Init initializes the step.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	if m.showingInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case createAccountMsg:
		m.creating = false
		if msg.err != nil {
			m.logger.Error("failed to create datadog account", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

		m.logger.Info("datadog account created", "id", msg.account.ID, "site", msg.account.Site)
		m.err = nil
		m.created = true
		m.createdAccount = msg.account
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "c":
			url := ddvendor.GetAppKeyURL(m.site)
			err := clipboard.WriteAll(url)
			if err != nil {
				m.logger.Error("failed to copy to clipboard", "error", err)
			} else {
				m.logger.Debug("copied URL to clipboard", "url", url)
				m.copiedURL = true
				if !m.showingInput {
					m.showingInput = true
					return m, m.input.Focus()
				}
			}
			return m, nil

		case "o":
			if !m.creating && !m.created {
				url := ddvendor.GetAppKeyURL(m.site)
				err := browser.OpenURL(url)
				if err != nil {
					m.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					m.logger.Debug("opened browser for app key creation", "url", url)
				}
				return m, nil
			}

		case "enter":
			// Interstitial screen: open browser and transition to input
			if !m.showingInput {
				url := ddvendor.GetAppKeyURL(m.site)
				err := browser.OpenURL(url)
				if err != nil {
					m.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					m.logger.Debug("opened browser for app key creation", "url", url)
				}
				m.showingInput = true
				return m, m.input.Focus()
			}

			// Input screen: retry on error
			if m.err != nil {
				appKey := m.input.Value()
				if appKey != "" {
					m.logger.Info("retrying datadog account creation")
					m.err = nil
					m.creating = true
					return m, m.createAccount(appKey)
				}
				return m, nil
			}

			// Input screen: submit
			if !m.creating && !m.created {
				appKey := m.input.Value()
				if appKey != "" {
					m.creating = true
					return m, m.createAccount(appKey)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// createAccount creates a Datadog account with both API and App keys.
func (m Model) createAccount(appKey string) tea.Cmd {
	return func() tea.Msg {
		m.logger.Debug("creating datadog account", "accountID", m.account.ID, "site", m.site)

		datadogAccount, err := m.services.DatadogAccounts.CreateAccount(
			m.ctx,
			uuid.New(),
			m.account.ID.String(),
			"Datadog",
			m.site,
			m.apiKey,
			appKey,
		)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: datadogAccount}
	}
}

// View renders the app key input UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	if m.creating {
		return lipgloss.JoinVertical(lipgloss.Left, "", themeStyles.Body.Render("Creating Datadog account..."))
	}

	title := themeStyles.Title.Render("Connect to Datadog")
	linkStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Underline(true)
	docsLink := themeStyles.Help.Render("Need help? ") + linkStyle.Render("docs.usetero.com/integrations/datadog")

	// Interstitial screen
	if !m.showingInput {
		prompt := themeStyles.Body.Render("Now, create a Service Account application key.")
		explanation := themeStyles.Help.Render("This lets Tero read your telemetry data and discover waste.")
		action := themeStyles.Action.Render("Press Enter to open Datadog.")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", prompt, explanation, "", action, "", docsLink)
	}

	// Input screen
	prompt := themeStyles.Body.Render("Paste the application key")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", prompt, "", m.input.View(), "", docsLink)
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

// IsBusy returns true while creating the account.
func (m Model) IsBusy() bool {
	return m.creating
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
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "retry")),
			},
		}
	}

	if !m.showingInput {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
				key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy URL")),
			},
		}
	}

	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open Datadog")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy URL")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.created || m.createdAccount == nil {
		return nil, step.ErrNotReady
	}

	return datadogdiscovery.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		m.account,
		&m.createdAccount.ID,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
