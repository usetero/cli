package datadog

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// createAccountMsg is sent when Datadog account creation completes
type createAccountMsg struct {
	account *api.DatadogAccount
	err     error
}

// AppKeyStep handles collecting the user's Datadog application key.
type AppKeyStep struct {
	// Context for API calls
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role    string
	org     api.Organization
	account api.Account
	site    string // Selected Datadog site
	apiKey  string // Validated API key from previous step

	// Services
	datadogAccounts api.DatadogAccounts
	workspaces      api.Workspaces
	preferences     preferences.Preferences
	apiClient       api.Client
	logger          log.Logger

	// UI state
	input          *input.Component
	showingInput   bool // false = interstitial, true = input screen
	creating       bool
	created        bool
	createdAccount *api.DatadogAccount
	err            error
	copiedURL      bool // true if URL was just copied
	width          int
	globalBindings []key.Binding
}

// NewAppKeyStep creates a new Datadog app key collection step
func NewAppKeyStep(ctx context.Context, theme *styles.Theme, role string, org api.Organization, account api.Account, site string, apiKey string, datadogAccounts api.DatadogAccounts, workspaces api.Workspaces, prefs preferences.Preferences, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if datadogAccounts == nil {
		panic("datadogAccounts cannot be nil")
	}
	if workspaces == nil {
		panic("workspaces cannot be nil")
	}
	if prefs == nil {
		panic("preferences cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	inp := input.New(theme, logger)
	inp.SetPlaceholder("Enter your Datadog application key...")
	inp.SetWidth(50)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &AppKeyStep{
		ctx:             ctx,
		theme:           theme,
		role:            role,
		org:             org,
		account:         account,
		site:            site,
		apiKey:          apiKey,
		datadogAccounts: datadogAccounts,
		workspaces:      workspaces,
		preferences:     prefs,
		apiClient:       apiClient,
		logger:          logger,
		input:           inp,
		showingInput:    false, // Start with interstitial
		width:           80,
		globalBindings:  globalBindings,
	}
}

// Init initializes the app key step
func (s *AppKeyStep) Init() tea.Cmd {
	// Don't focus input yet - we're on the interstitial screen
	return nil
}

// Update handles messages
func (s *AppKeyStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	// Always update input for cursor blinking if we're showing it
	if s.showingInput {
		cmd := s.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case createAccountMsg:
		s.creating = false
		if msg.err != nil {
			s.logger.Error("failed to create datadog account", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		s.logger.Info("datadog account created", log.String("id", msg.account.ID), log.String("site", msg.account.Site))

		// Clear any previous error and mark success
		s.err = nil
		s.created = true
		s.createdAccount = msg.account
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			// Copy URL to clipboard
			url := ddvendor.GetAppKeyURL(s.site)
			err := clipboard.WriteAll(url)
			if err != nil {
				s.logger.Error("failed to copy to clipboard", "error", err)
			} else {
				s.logger.Debug("copied URL to clipboard", "url", url)
				s.copiedURL = true
				// If we're on the interstitial, transition to input screen after copy
				if !s.showingInput {
					s.showingInput = true
					return s, s.input.Focus()
				}
			}
			return s, nil

		case "o":
			if !s.creating && !s.created {
				// Open Datadog Application key creation page
				url := ddvendor.GetAppKeyURL(s.site)
				err := openBrowser(url)
				if err != nil {
					s.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					s.logger.Debug("opened browser for app key creation", "url", url)
				}
				return s, nil
			}

		case "enter":
			// Interstitial screen: open browser and transition to input
			if !s.showingInput {
				url := ddvendor.GetAppKeyURL(s.site)
				err := openBrowser(url)
				if err != nil {
					s.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					s.logger.Debug("opened browser for app key creation", "url", url)
				}
				s.showingInput = true
				return s, s.input.Focus()
			}

			// Input screen: retry on error
			if s.err != nil {
				appKey := s.input.Value()
				if appKey != "" {
					s.logger.Info("retrying datadog account creation")
					s.err = nil
					s.creating = true
					return s, s.createAccount(appKey)
				}
				return s, nil
			}

			// Input screen: submit if not already creating
			if !s.creating && !s.created {
				appKey := s.input.Value()
				if appKey != "" {
					s.creating = true
					return s, s.createAccount(appKey)
				}
			}
		}
	}

	return s, tea.Batch(cmds...)
}

// createAccount creates a Datadog account with both API and App keys
func (s *AppKeyStep) createAccount(appKey string) tea.Cmd {
	return func() tea.Msg {
		s.logger.Debug("creating datadog account", log.String("accountID", s.account.ID), log.String("site", s.site))

		datadogAccount, err := s.datadogAccounts.CreateAccount(
			s.ctx,
			uuid.New(),
			s.account.ID,
			"Datadog", // Default name
			s.site,
			s.apiKey,
			appKey,
		)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: datadogAccount}
	}
}

// View renders the app key input UI
func (s *AppKeyStep) View() string {
	themeStyles := s.theme.Styles
	colors := s.theme.Colors

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			themeStyles.Body.Render("Creating Datadog account..."),
		)
	}

	title := themeStyles.Title.Render("Connect to Datadog")

	linkStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Underline(true)
	docsLink := themeStyles.Help.Render("Need help? ") + linkStyle.Render("docs.usetero.com/integrations/datadog")

	// Interstitial screen
	if !s.showingInput {
		prompt := themeStyles.Body.Render("Now, create a Service Account application key.")
		explanation := themeStyles.Subtitle.Render("This lets Tero read your telemetry data and discover waste.")
		action := themeStyles.Action.Render("Press Enter to open Datadog.")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			prompt,
			explanation,
			"",
			action,
			"",
			docsLink,
		)
	}

	// Input screen
	prompt := themeStyles.Body.Render("Paste the application key")

	parts := []string{
		title,
		"",
		prompt,
		"",
		s.input.View(),
		"",
		docsLink,
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize sets the width available for rendering
func (s *AppKeyStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsBusy returns true while creating the account
func (s *AppKeyStep) IsBusy() bool {
	return s.creating
}

// HasError returns true if there was an error creating the Datadog account
func (s *AppKeyStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *AppKeyStep) Error() error {
	return s.err
}

// Next returns the next step after account creation
func (s *AppKeyStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.created || s.createdAccount == nil {
		return nil, step.ErrNotReady
	}

	// Create datadog account service for status polling
	datadogAccountService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// Datadog account created - move to unified discovery step
	datadogAccountID := s.createdAccount.ID
	return NewDiscoveryStep(s.ctx, s.theme, s.role, s.org, s.account, &datadogAccountID, datadogAccountService, s.workspaces, s.preferences, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *AppKeyStep) Help() help.KeyMap {
	// Show retry hint if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	// Interstitial screen
	if !s.showingInput {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "open"),
				),
				key.NewBinding(
					key.WithKeys("c"),
					key.WithHelp("c", "copy URL"),
				),
			},
		}
	}

	// Input screen
	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "open Datadog"),
			),
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "copy URL"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
		},
	}
}

// Close releases any resources held by the step.
func (s *AppKeyStep) Close() error {
	return nil
}
