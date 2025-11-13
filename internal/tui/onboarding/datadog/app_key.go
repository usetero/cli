package datadog

import (
	"context"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/services"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

// DatadogAccountCreator creates Datadog accounts
type DatadogAccountCreator interface {
	CreateAccount(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error)
}

// createAccountMsg is sent when Datadog account creation completes
type createAccountMsg struct {
	account *api.DatadogAccount
	err     error
}

// AppKeyStep handles collecting the user's Datadog application key.
type AppKeyStep struct {
	// Accumulated state from previous steps
	role      string
	orgID     string
	accountID string
	site      string // Selected Datadog site
	apiKey    string // Validated API key from previous step

	// Services (defined by consumer interfaces)
	accountCreator DatadogAccountCreator

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

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
func NewAppKeyStep(role string, orgID string, accountID string, site string, apiKey string, accountCreator DatadogAccountCreator, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if accountCreator == nil {
		panic("accountCreator cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	inp := input.New(logger)
	inp.SetPlaceholder("Enter your Datadog application key...")
	inp.SetWidth(50)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &AppKeyStep{
		role:           role,
		orgID:          orgID,
		accountID:      accountID,
		site:           site,
		apiKey:         apiKey,
		accountCreator: accountCreator,
		apiClient:      apiClient,
		logger:         logger,
		input:          inp,
		showingInput:   false, // Start with interstitial
		width:          80,
		globalBindings: globalBindings,
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
				err := browser.OpenURL(url)
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
				err := browser.OpenURL(url)
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
		ctx := context.Background()
		s.logger.Debug("creating datadog account", log.String("accountID", s.accountID), log.String("site", s.site))

		account, err := s.accountCreator.CreateAccount(
			ctx,
			s.accountID,
			"Datadog", // Default name
			s.site,
			s.apiKey,
			appKey,
		)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: account}
	}
}

// View renders the app key input UI
func (s *AppKeyStep) View() string {
	common := styles.Common()

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			common.Body.Render("Creating Datadog account..."),
		)
	}

	stepTitle := common.Title.Render("Step 3 of 3: Create a service account")
	url := ddvendor.GetAppKeyURL(s.site)

	// Interstitial screen
	if !s.showingInput {
		subtitle := common.Body.Render("Next, create a service account called \"Tero\" and copy its Application key:")
		instruction := common.Action.Render("Press Enter to open in browser, or press 'c' to copy the URL")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			stepTitle,
			"",
			subtitle,
			"",
			common.URL.Render("  "+url),
			"",
			instruction,
		)
	}

	// Input screen
	var statusLine string
	if s.copiedURL {
		statusLine = common.Success.Render("✓ URL copied to clipboard")
	}

	subtitle := common.Body.Render("Create a service account called \"Tero\", then paste its Application key here:")

	parts := []string{
		stepTitle,
		"",
		subtitle,
		"",
		common.URL.Render("  " + url),
	}

	if statusLine != "" {
		parts = append(parts, "", statusLine)
	}

	parts = append(parts, "", s.input.View())

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize sets the width available for rendering
func (s *AppKeyStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsComplete returns true if Datadog account has been created successfully
func (s *AppKeyStep) IsComplete() bool {
	return s.created && s.createdAccount != nil && s.err == nil
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
func (s *AppKeyStep) Next() step.Step {
	// Create service service for next step
	serviceService := api.NewServiceService(s.apiClient, s.logger)

	// Datadog account created - move to service discovery
	datadogAccountID := s.createdAccount.ID
	return services.NewDiscoveryStep(s.role, s.orgID, s.accountID, &datadogAccountID, serviceService, s.apiClient, s.logger, s.globalBindings)
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
