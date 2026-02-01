package auth

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	authservice "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/pkg/client"
)

// CheckAuthStep checks if the user has a valid auth token
type CheckAuthStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme
	theme *styles.Theme

	// Services
	auth        authservice.Auth
	preferences preferences.Preferences
	apiEndpoint string
	logger      log.Logger

	globalBindings []key.Binding

	// UI state
	checking     bool
	checked      bool
	hasValidAuth bool
	accessToken  string
	err          error
	width        int
}

// NewCheckAuthStep creates a new auth check step
func NewCheckAuthStep(ctx context.Context, theme *styles.Theme, authService authservice.Auth, prefs preferences.Preferences, apiEndpoint string, logger log.Logger, globalBindings []key.Binding) step.Step {
	if authService == nil {
		panic("authService cannot be nil")
	}
	if prefs == nil {
		panic("prefs cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CheckAuthStep{
		ctx:            ctx,
		theme:          theme,
		auth:           authService,
		preferences:    prefs,
		apiEndpoint:    apiEndpoint,
		logger:         logger,
		globalBindings: globalBindings,
		width:          80,
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
		s.logger.Info("checking authentication")

		if !s.auth.IsAuthenticated() {
			return checkAuthMsg{hasValidAuth: false}
		}

		// Get the access token
		accessToken, err := s.auth.GetAccessToken(s.ctx)
		if err != nil {
			s.logger.Warn("failed to get access token, clearing tokens", "error", err)
			// Clear invalid tokens
			_ = s.auth.ClearTokens()
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
	styles := s.theme.Styles

	if s.checking {
		return styles.Title.Render("Checking authentication...")
	}

	if s.hasValidAuth {
		return styles.Success.Render("Already authenticated")
	}

	return styles.Title.Render("No authentication found")
}

// SetSize sets the width available for rendering
func (s *CheckAuthStep) SetSize(width, height int) {
	s.width = width
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
func (s *CheckAuthStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.checked {
		return nil, step.ErrNotReady
	}

	if !s.hasValidAuth {
		// No valid auth - go to auth step
		return NewAuthenticateStep(s.ctx, s.theme, s.logger, s.auth, s.preferences, s.apiEndpoint, s.globalBindings), nil
	}

	// Has valid auth - create authenticated client and go to role selection
	refreshFunc := func() (string, error) {
		return s.auth.GetAccessToken(s.ctx)
	}
	apiClient := client.New(s.apiEndpoint, s.accessToken, refreshFunc)
	return role.NewSelectStep(s.ctx, s.theme, apiClient, s.preferences, s.auth, s.logger, s.globalBindings), nil
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
