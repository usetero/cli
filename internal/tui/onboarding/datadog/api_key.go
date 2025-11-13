package datadog

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	ddvendor "github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

// validateAPIKeyMsg is sent when Datadog API key validation completes
type validateAPIKeyMsg struct {
	apiKey string
	valid  bool
	errMsg string
}

// DatadogAPIKeyValidator validates Datadog API keys
type DatadogAPIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string, site string) (bool, string, error)
}

// APIKeyStep handles collecting the user's Datadog API key.
type APIKeyStep struct {
	// Accumulated state from previous steps
	role      string
	orgID     string
	accountID string
	site      string // Selected Datadog site (US1, EU1, etc.)

	// Services (defined by consumer interfaces)
	apiKeyValidator DatadogAPIKeyValidator

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

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
func NewAPIKeyStep(role string, orgID string, accountID string, site string, apiKeyValidator DatadogAPIKeyValidator, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if apiKeyValidator == nil {
		panic("apiKeyValidator cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	theme := styles.CurrentTheme()

	inp := input.New(logger)
	inp.SetPlaceholder("Enter your Datadog API key...")
	inp.SetWidth(50)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &APIKeyStep{
		role:            role,
		orgID:           orgID,
		accountID:       accountID,
		site:            site,
		apiKeyValidator: apiKeyValidator,
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
			err := browser.OpenURL(url)
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
				err := browser.OpenURL(url)
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
		ctx := context.Background()
		s.logger.Debug("validating datadog api key", log.String("site", s.site))

		valid, errorMsg, err := s.apiKeyValidator.ValidateAPIKey(ctx, apiKey, s.site)
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
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	helpStyle := lipgloss.NewStyle().
		Foreground(theme.TextMuted)

	successStyle := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	// Show success state
	if s.validated {
		title := successStyle.Render("✓ Datadog API key verified!")
		help := helpStyle.Render("Press Enter to continue")
=======
	common := styles.Common()
	theme := styles.CurrentTheme()

	// Show success state
	if s.validated {
		title := common.Success.Render("✓ Datadog API key verified!")
		help := common.Help.Render("Press Enter to continue")
>>>>>>> 17e8dd9 (chore: initial commit)
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
			Foreground(theme.Primary).
			Bold(false)

		title := titleStyle.Render("Verifying your Datadog API key...")
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.spinner.View()+" Connecting to control plane...",
		)
	}

<<<<<<< HEAD
	stepTitleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)
	stepTitle := stepTitleStyle.Render("Step 2 of 3: Get your API key")

	urlStyle := lipgloss.NewStyle().
		Foreground(theme.TextSubtle)
=======
	stepTitle := common.Title.Render("Step 2 of 3: Get your API key")
>>>>>>> 17e8dd9 (chore: initial commit)
	url := ddvendor.GetAPIKeyURL(s.site)

	// Interstitial screen
	if !s.showingInput {
<<<<<<< HEAD
		subtitleStyle := lipgloss.NewStyle().
			Foreground(theme.Text)
		subtitle := subtitleStyle.Render("Datadog uses two keys for access. First, your API key:")

		instructionStyle := lipgloss.NewStyle().
			Foreground(theme.Primary)
		instruction := instructionStyle.Render("Press Enter to open in browser, or press 'c' to copy the URL")
=======
		subtitle := common.Body.Render("Datadog uses two keys for access. First, your API key:")
		instruction := common.Action.Render("Press Enter to open in browser, or press 'c' to copy the URL")
>>>>>>> 17e8dd9 (chore: initial commit)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			stepTitle,
			"",
			subtitle,
			"",
<<<<<<< HEAD
			urlStyle.Render("  "+url),
=======
			common.URL.Render("  "+url),
>>>>>>> 17e8dd9 (chore: initial commit)
			"",
			instruction,
		)
	}

	// Input screen
<<<<<<< HEAD
	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	var statusLine string
	if s.copiedURL {
		copyStyle := lipgloss.NewStyle().
			Foreground(theme.Success)
		statusLine = copyStyle.Render("✓ URL copied to clipboard")
	}

	subtitle := subtitleStyle.Render("Create an API key in Datadog, then paste it here:")
=======
	var statusLine string
	if s.copiedURL {
		statusLine = common.Success.Render("✓ URL copied to clipboard")
	}

	subtitle := common.Body.Render("Create an API key in Datadog, then paste it here:")
>>>>>>> 17e8dd9 (chore: initial commit)

	parts := []string{
		stepTitle,
		"",
		subtitle,
		"",
<<<<<<< HEAD
		urlStyle.Render("  " + url),
=======
		common.URL.Render("  " + url),
>>>>>>> 17e8dd9 (chore: initial commit)
	}

	if statusLine != "" {
		parts = append(parts, "", statusLine)
	}

	parts = append(parts, "", s.input.View())

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize sets the width available for rendering
func (s *APIKeyStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsComplete returns true if a valid Datadog API key has been collected and validated
func (s *APIKeyStep) IsComplete() bool {
	return s.validated && s.validatedKey != ""
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
func (s *APIKeyStep) Next() step.Step {
	// Create Datadog service for next step
	datadogService := api.NewDatadogAccountService(s.apiClient, s.logger)

	return NewAppKeyStep(s.role, s.orgID, s.accountID, s.site, s.validatedKey, datadogService, s.apiClient, s.logger, s.globalBindings)
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
