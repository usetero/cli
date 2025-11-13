package auth

import (
	"context"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
<<<<<<< HEAD
	"github.com/charmbracelet/lipgloss/v2"
=======
>>>>>>> 17e8dd9 (chore: initial commit)
	authservice "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
	"github.com/usetero/cli/pkg/client"
)

// TokenValidator validates stored auth tokens
type TokenValidator interface {
	IsAuthenticated() bool
	GetAccessToken(ctx context.Context) (string, error)
	ClearTokens() error
}

// CheckAuthStep checks if the user has a valid auth token
type CheckAuthStep struct {
	// Services
	tokenValidator TokenValidator
	authService    *authservice.Service

	// Pass-through to next step
	preferencesService *preferences.Service
	apiEndpoint        string
	logger             log.Logger
	globalBindings     []key.Binding

	// UI state
	checking     bool
	checked      bool
	hasValidAuth bool
	accessToken  string
	err          error
	width        int
}

// NewCheckAuthStep creates a new auth check step
func NewCheckAuthStep(tokenValidator TokenValidator, authService *authservice.Service, preferencesService *preferences.Service, apiEndpoint string, logger log.Logger, globalBindings []key.Binding) step.Step {
	if tokenValidator == nil {
		panic("tokenValidator cannot be nil")
	}
	if authService == nil {
		panic("authService cannot be nil")
	}
	if preferencesService == nil {
		panic("preferencesService cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CheckAuthStep{
		tokenValidator:     tokenValidator,
		authService:        authService,
		preferencesService: preferencesService,
		apiEndpoint:        apiEndpoint,
		logger:             logger,
		globalBindings:     globalBindings,
		width:              80,
	}
}

// checkAuthMsg is sent when check completes
type checkAuthMsg struct {
	hasValidAuth bool
	accessToken  string
	err          error
}

// Init starts checking for valid auth
func (s *CheckAuthStep) Init() tea.Cmd {
	s.checking = true
	return s.checkAuth()
}

// checkAuth checks if there's a valid access token
func (s *CheckAuthStep) checkAuth() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		s.logger.Info("checking authentication")

		if !s.tokenValidator.IsAuthenticated() {
			return checkAuthMsg{hasValidAuth: false}
		}

		// Get the access token
		accessToken, err := s.tokenValidator.GetAccessToken(ctx)
		if err != nil {
			s.logger.Warn("failed to get access token, clearing tokens", "error", err)
			// Clear invalid tokens
			_ = s.tokenValidator.ClearTokens()
			return checkAuthMsg{hasValidAuth: false}
		}

		// TODO: Optionally validate token with API
		// For now, just check if it exists
		return checkAuthMsg{hasValidAuth: true, accessToken: accessToken}
	}
}

// Update handles messages
func (s *CheckAuthStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case checkAuthMsg:
		s.checking = false
		if msg.err != nil {
			s.logger.Error("failed to check auth", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		// Clear any previous error
		s.err = nil
		s.checked = true
		s.hasValidAuth = msg.hasValidAuth
		s.accessToken = msg.accessToken

		if s.hasValidAuth {
			s.logger.Info("valid authentication found")
		} else {
			s.logger.Info("no valid authentication found")
		}
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Retry checking authentication if there was an error
			if s.err != nil {
				s.logger.Info("retrying authentication check")
				s.err = nil
				s.checking = true
				return s, s.checkAuth()
			}
		}
	}

	return s, nil
}

// View renders the check UI
func (s *CheckAuthStep) View() string {
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	if s.checking {
		return titleStyle.Render("Checking authentication...")
	}

	if s.hasValidAuth {
		successStyle := lipgloss.NewStyle().
			Foreground(theme.Success).
			Bold(true)
		return successStyle.Render("✓ Already authenticated")
	}

	return titleStyle.Render("No authentication found")
=======
	common := styles.Common()

	if s.checking {
		return common.Title.Render("Checking authentication...")
	}

	if s.hasValidAuth {
		return common.Success.Render("✓ Already authenticated")
	}

	return common.Title.Render("No authentication found")
>>>>>>> 17e8dd9 (chore: initial commit)
}

// SetSize sets the width available for rendering
func (s *CheckAuthStep) SetSize(width, height int) {
	s.width = width
}

// IsComplete returns true when check is done successfully
func (s *CheckAuthStep) IsComplete() bool {
	return s.checked && s.err == nil
}

// NeedsAuth returns true if user needs to authenticate
func (s *CheckAuthStep) NeedsAuth() bool {
	return s.checked && !s.hasValidAuth
}

// IsBusy returns true while checking
func (s *CheckAuthStep) IsBusy() bool {
	return s.checking
}

// HasError returns true if there was an error checking authentication
func (s *CheckAuthStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *CheckAuthStep) Error() error {
	return s.err
}

// Next returns the next step after checking auth
func (s *CheckAuthStep) Next() step.Step {
	if s.NeedsAuth() {
		// No valid auth - go to auth step
		return NewAuthenticateStep(s.logger, s.authService, s.preferencesService, s.apiEndpoint, s.globalBindings)
	}

	// Has valid auth - create authenticated client and go to role selection
	apiClient := client.New(s.apiEndpoint, s.accessToken)
	return role.NewSelectStep(apiClient, s.preferencesService, s.logger, s.globalBindings)
}

// Help returns the key bindings for this step
func (s *CheckAuthStep) Help() help.KeyMap {
	// Show retry option if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	// No user interaction during normal checking
	return keymap.Simple{Keys: []key.Binding{}}
}
