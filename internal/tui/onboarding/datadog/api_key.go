package datadog

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
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// validateAPIKeyMsg is sent when Datadog API key validation completes
type validateAPIKeyMsg struct {
	apiKey string
	valid  bool
	errMsg string
}

// APIKeyStep handles collecting the user's Datadog API key.
type APIKeyStep struct {
	// Context for API calls
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role    string
	org     api.Organization
	account api.Account
	site    string // Selected Datadog site (US1, EU1, etc.)

	// Services
	datadogAccounts api.DatadogAccounts
	apiClient       api.Client
	logger          log.Logger

	// UI state
	input          *input.Component
	spinner        spinner.Model
	showingInput   bool // false = interstitial, true = input screen
	validating     bool
	validated      bool
	validatedKey   string // Validated API key stored in-memory
	validationErr  error  // Validation error if API key is invalid
	copiedURL      bool   // true if URL was just copied
	width          int
	globalBindings []key.Binding
}

// NewAPIKeyStep creates a new Datadog API key collection step
func NewAPIKeyStep(ctx context.Context, theme *styles.Theme, role string, org api.Organization, account api.Account, site string, datadogAccounts api.DatadogAccounts, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if datadogAccounts == nil {
		panic("datadogAccounts cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	colors := theme.Colors

	inp := input.New(theme, logger)
	inp.SetPlaceholder("Enter your Datadog API key...")
	inp.SetWidth(50)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return &APIKeyStep{
		ctx:             ctx,
		theme:           theme,
		role:            role,
		org:             org,
		account:         account,
		site:            site,
		datadogAccounts: datadogAccounts,
		apiClient:       apiClient,
		logger:          logger,
		input:           inp,
		spinner:         sp,
		showingInput:    false, // Start with interstitial
		width:           80,
		globalBindings:  globalBindings,
	}
}

// Init initializes the Datadog API key step
func (s *APIKeyStep) Init() tea.Cmd {
	// Don't focus input yet - we're on the interstitial screen
	return nil
}

// Update handles messages for the Datadog API key step
func (s *APIKeyStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Always update input for cursor blinking if we're showing it
	if s.showingInput {
		cmd = s.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle spinner ticks
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if s.validating {
			s.spinner, cmd = s.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case validateAPIKeyMsg:
		// Exit busy state
		s.validating = false

		// Handle validation errors
		if !msg.valid {
			s.logger.Info("datadog api key invalid", log.String("error", msg.errMsg))
			s.validationErr = fmt.Errorf("%s", msg.errMsg)
			return s, nil
		}

		// Success! Clear any previous error and store the validated key
		s.validationErr = nil
		s.validatedKey = msg.apiKey
		s.validated = true
		s.logger.Info("datadog api key validated")
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			// Copy URL to clipboard
			url := ddvendor.GetAPIKeyURL(s.site)
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
			// Open Datadog API key creation page
			url := ddvendor.GetAPIKeyURL(s.site)
			err := openBrowser(url)
			if err != nil {
				s.logger.Error("failed to open browser", "error", err, "url", url)
			} else {
				s.logger.Debug("opened browser for API key creation", "url", url)
			}
			return s, nil

		case "enter":
			// Interstitial screen: open browser and transition to input
			if !s.showingInput {
				url := ddvendor.GetAPIKeyURL(s.site)
				err := openBrowser(url)
				if err != nil {
					s.logger.Error("failed to open browser", "error", err, "url", url)
				} else {
					s.logger.Debug("opened browser for API key creation", "url", url)
				}
				s.showingInput = true
				return s, s.input.Focus()
			}

			// Input screen: submit API key
			if s.validating || s.validated {
				return s, nil
			}

			apiKey := s.input.Value()
			if apiKey != "" {
				s.validating = true
				return s, tea.Batch(s.spinner.Tick, s.validateAPIKey(apiKey))
			}
		}
	}

	return s, tea.Batch(cmds...)
}

// validateAPIKey validates the API key via the control plane
func (s *APIKeyStep) validateAPIKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		s.logger.Debug("validating datadog api key", log.String("site", s.site))

		valid, errorMsg, err := s.datadogAccounts.ValidateAPIKey(s.ctx, apiKey, s.site)
		if err != nil {
			s.logger.Error("failed to validate api key", "error", err)
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

// View renders the Datadog API key step UI
func (s *APIKeyStep) View() string {
	themeStyles := s.theme.Styles
	colors := s.theme.Colors

	// Show success state
	if s.validated {
		title := themeStyles.Success.Render("✓ Datadog API key verified!")
		help := themeStyles.Help.Render("Press Enter to continue")
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			help,
		)
	}

	// Show validating state
	if s.validating {
		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(false)

		title := titleStyle.Render("Verifying your Datadog API key...")
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.spinner.View()+" Connecting to control plane...",
		)
	}

	title := themeStyles.Title.Render("Connect to Datadog")

	linkStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Underline(true)
	docsLink := themeStyles.Help.Render("Need help? ") + linkStyle.Render("docs.usetero.com/integrations/datadog")

	// Interstitial screen
	if !s.showingInput {
		prompt := themeStyles.Body.Render("First, create an API key.")
		action := themeStyles.Action.Render("Press Enter to open Datadog.")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			prompt,
			"",
			action,
			"",
			docsLink,
		)
	}

	// Input screen
	prompt := themeStyles.Body.Render("Paste your API key")

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
func (s *APIKeyStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsBusy returns true while validating
func (s *APIKeyStep) IsBusy() bool {
	return s.validating
}

// HasError returns true if validation failed
func (s *APIKeyStep) HasError() bool {
	return s.validationErr != nil
}

// Error returns the validation error, or nil if no error
func (s *APIKeyStep) Error() error {
	return s.validationErr
}

// Next returns the next step after API key validation
func (s *APIKeyStep) Next() (step.Step, error) {
	if s.validationErr != nil {
		return nil, s.validationErr
	}
	if !s.validated || s.validatedKey == "" {
		return nil, step.ErrNotReady
	}

	// Create Datadog service for next step
	datadogService := api.NewDatadogAccountService(s.apiClient, s.logger)

	return NewAppKeyStep(s.ctx, s.theme, s.role, s.org, s.account, s.site, s.validatedKey, datadogService, s.apiClient, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *APIKeyStep) Help() help.KeyMap {
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
