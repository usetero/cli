package datadogapikey

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	datadogappkey "github.com/usetero/cli/internal/tui/onboarding/datadog/appkey"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// validateAPIKeyMsg is sent when Datadog API key validation completes.
type validateAPIKeyMsg struct {
	apiKey string
	valid  bool
	errMsg string
}

// Model handles collecting the user's Datadog API key.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	role    string
	org     domain.Organization
	account domain.Account
	site    string // Selected Datadog site (US1, EU1, etc.)

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	input         input.Model
	spinner       spinner.Model
	showingInput  bool // false = interstitial, true = input screen
	validating    bool
	validated     bool
	validatedKey  string // Validated API key stored in-memory
	validationErr error
	copiedURL     bool
	width         int
	height        int
}

// New creates a new API key collection model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	account domain.Account,
	site string,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	colors := theme.Colors

	inp := input.New(theme, logger).
		SetPlaceholder("Enter your Datadog API key...").
		SetWidth(50).
		SetEchoMode(textinput.EchoPassword).
		SetEchoCharacter('•')

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return Model{
		ctx:          ctx,
		theme:        theme,
		role:         role,
		org:          org,
		account:      account,
		site:         site,
		services:     services,
		prefs:        prefs,
		logger:       logger,
		input:        inp,
		spinner:      sp,
		showingInput: false, // Start with interstitial
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
	var cmd tea.Cmd

	// Always update input for cursor blinking if we're showing it
	if m.showingInput {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.validating {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case validateAPIKeyMsg:
		m.validating = false

		if !msg.valid {
			m.logger.Info("datadog api key invalid", "error", msg.errMsg)
			m.validationErr = fmt.Errorf("%s", msg.errMsg)
			return m, nil
		}

		m.validationErr = nil
		m.validatedKey = msg.apiKey
		m.validated = true
		m.logger.Info("datadog api key validated")
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "c":
			url := ddvendor.GetAPIKeyURL(m.site)
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
			url := ddvendor.GetAPIKeyURL(m.site)
			err := browser.OpenURL(url)
			if err != nil {
				m.logger.Error("failed to open browser", "error", err, "url", url)
			} else {
				m.logger.Debug("opened browser for API key creation", "url", url)
			}
			return m, nil

		case "enter":
			// Interstitial screen: open browser and transition to input
			if !m.showingInput {
				url := ddvendor.GetAPIKeyURL(m.site)
				err := browser.OpenURL(url)
				if err != nil {
					m.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					m.logger.Debug("opened browser for API key creation", "url", url)
				}
				m.showingInput = true
				return m, m.input.Focus()
			}

			// Input screen: submit API key
			if m.validating || m.validated {
				return m, nil
			}

			apiKey := m.input.Value()
			if apiKey != "" {
				m.validating = true
				return m, tea.Batch(m.spinner.Tick, m.validateAPIKey(apiKey))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// validateAPIKey validates the API key via the control plane.
func (m Model) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		m.logger.Debug("validating datadog api key", "site", m.site)

		valid, errorMsg, err := m.services.DatadogAccounts.ValidateAPIKey(m.ctx, apiKey, m.site)
		if err != nil {
			m.logger.Error("failed to validate api key", "error", err)
			return validateAPIKeyMsg{
				apiKey: apiKey,
				valid:  false,
				errMsg: "Failed to connect to control plane",
			}
		}

		return validateAPIKeyMsg{
			apiKey: apiKey,
			valid:  valid,
			errMsg: errorMsg,
		}
	}
}

// View renders the API key step UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	if m.validated {
		title := themeStyles.Success.Render("Datadog API key verified!")
		help := themeStyles.Help.Render("Press Enter to continue")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", help)
	}

	if m.validating {
		titleStyle := lipgloss.NewStyle().Foreground(colors.Accent).Bold(false)
		title := titleStyle.Render("Verifying your Datadog API key...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", m.spinner.View()+" Connecting to control plane...")
	}

	title := themeStyles.Title.Render("Connect to Datadog")
	linkStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Underline(true)
	docsLink := themeStyles.Help.Render("Need help? ") + linkStyle.Render("docs.usetero.com/integrations/datadog")

	// Interstitial screen
	if !m.showingInput {
		prompt := themeStyles.Body.Render("First, create an API key.")
		action := themeStyles.Action.Render("Press Enter to open Datadog.")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", prompt, "", action, "", docsLink)
	}

	// Input screen
	prompt := themeStyles.Body.Render("Paste your API key")
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

// IsBusy returns true while validating.
func (m Model) IsBusy() bool {
	return m.validating
}

// HasError returns true if validation failed.
func (m Model) HasError() bool {
	return m.validationErr != nil
}

// Error returns the validation error.
func (m Model) Error() error {
	return m.validationErr
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
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
	if m.validationErr != nil {
		return nil, m.validationErr
	}
	if !m.validated || m.validatedKey == "" {
		return nil, step.ErrNotReady
	}

	return datadogappkey.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		m.account,
		m.site,
		m.validatedKey,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
