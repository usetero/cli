package auth

import (
	"context"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	authservice "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/pkg/client"
)

// authState tracks the current state of the authentication flow
type authState int

const (
	stateInitializing authState = iota
	stateReady                  // Ready for user to open browser or copy URL
	stateComplete
)

// AuthenticateStep handles device code flow authentication.
type AuthenticateStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme
	theme *styles.Theme

	// Services
	authService    authservice.Auth
	preferences    preferences.Preferences
	apiEndpoint    string
	logger         log.Logger
	globalBindings []key.Binding

	// UI state
	width             int
	state             authState
	deviceAuth        *authservice.DeviceAuth
	authResult        *authservice.Result
	err               error
	polling           bool
	openFailed        bool
	copiedToClipboard bool
	spinner           spinner.Model
}

// deviceAuthMsg is sent when device authorization is initiated
type deviceAuthMsg struct {
	deviceAuth *authservice.DeviceAuth
	err        error
}

// authCompleteMsg is sent when authentication completes
type authCompleteMsg struct {
	result *authservice.Result
	err    error
}

// NewAuthenticateStep creates a new authentication step
func NewAuthenticateStep(ctx context.Context, theme *styles.Theme, logger log.Logger, authService authservice.Auth, prefs preferences.Preferences, apiEndpoint string, globalBindings []key.Binding) step.Step {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if authService == nil {
		panic("authService cannot be nil")
	}
	if prefs == nil {
		panic("prefs cannot be nil")
	}

	colors := theme.Colors

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return &AuthenticateStep{
		ctx:            ctx,
		theme:          theme,
		authService:    authService,
		preferences:    prefs,
		apiEndpoint:    apiEndpoint,
		logger:         logger,
		globalBindings: globalBindings,
		state:          stateInitializing,
		spinner:        sp,
	}
}

// Init initializes the auth step by starting device authorization
func (s *AuthenticateStep) Init() tea.Cmd {
	// Check if already authenticated
	if s.authService.IsAuthenticated() {
		s.logger.Info("already authenticated")
		s.state = stateComplete
		return nil
	}

	// Start device authorization and spinner
	return tea.Batch(
		s.spinner.Tick,
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
			defer cancel()

			deviceAuth, err := s.authService.StartDeviceAuth(ctx)
			return deviceAuthMsg{deviceAuth: deviceAuth, err: err}
		},
	)
}

// Update handles messages
func (s *AuthenticateStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case deviceAuthMsg:
		if msg.err != nil {
			s.logger.Error("failed to start device authorization", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		s.deviceAuth = msg.deviceAuth
		s.state = stateReady
		s.err = nil // Clear any previous error
		s.logger.Info("device authorization started")
		s.logger.Debug("device auth details", "user_code", s.deviceAuth.UserCode, "expires_in", s.deviceAuth.ExpiresIn)

		// Auto-open browser and start polling
		err := browser.OpenURL(s.deviceAuth.VerificationURIComplete)
		if err != nil {
			s.logger.Warn("failed to auto-open browser", "error", err)
			s.openFailed = true
			// Don't start polling yet - let user manually open or copy URL
			return s, nil
		}

		s.logger.Debug("auto-opened browser for auth", "url", s.deviceAuth.VerificationURIComplete)
		s.polling = true
		return s, s.pollForAuth()

	case authCompleteMsg:
		s.polling = false

		if msg.err != nil {
			s.logger.Error("authentication failed", "error", msg.err)
			s.err = msg.err
			// Stay in stateReady so user can retry
			return s, nil
		}

		s.authResult = msg.result
		s.state = stateComplete
		s.logger.Info("authentication complete", "user_email", s.authResult.User.Email)
		return s, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tea.KeyMsg:
		if s.state == stateReady {
			switch msg.String() {
			case "enter":
				// Try to open browser (allow anytime, user might want different browser)
				err := browser.OpenURL(s.deviceAuth.VerificationURIComplete)
				if err != nil {
					s.logger.Warn("failed to open browser", "error", err)
					s.openFailed = true
					return s, nil
				}

				s.logger.Debug("opened browser for auth", "url", s.deviceAuth.VerificationURIComplete)

				// Successfully opened, start polling if not already
				s.openFailed = false
				s.err = nil // Clear any error when user retries
				if !s.polling {
					s.polling = true
					return s, s.pollForAuth()
				}
				return s, nil

			case "c":
				// Copy URL to clipboard (allow anytime)
				err := clipboard.WriteAll(s.deviceAuth.VerificationURIComplete)
				if err != nil {
					s.logger.Error("failed to copy to clipboard", "error", err)
					return s, nil
				}
				s.copiedToClipboard = true
				s.logger.Debug("auth URL copied to clipboard")

				// Start polling if not already
				s.err = nil // Clear any error when user retries
				if !s.polling {
					s.polling = true
					return s, s.pollForAuth()
				}
				return s, nil

			case "r":
				// Restart - get new device code (only show for recoverable errors)
				if s.err != nil && isRecoverableError(s.err) {
					s.polling = false
					s.err = nil
					s.openFailed = false
					s.copiedToClipboard = false
					return s, s.Init()
				}
			}
		}
	}

	return s, nil
}

// pollForAuth starts the background polling process
func (s *AuthenticateStep) pollForAuth() tea.Cmd {
	return func() tea.Msg {
		interval := time.Duration(s.deviceAuth.Interval) * time.Second

		result, err := s.authService.WaitForAuth(s.ctx, s.deviceAuth.DeviceCode, interval)
		return authCompleteMsg{result: result, err: err}
	}
}

// View renders the auth step
func (s *AuthenticateStep) View() string {
	styles := s.theme.Styles
	colors := s.theme.Colors

	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch s.state {
	case stateInitializing:
		return styles.Title.Render("Initializing authentication...")

	case stateReady:
		if s.deviceAuth == nil {
			return styles.Title.Render("Loading...")
		}

		var parts []string

		// Title
		parts = append(parts, styles.Title.Render("Authenticate with Tero"), "")

		// Code - prominently displayed so user can verify it matches the browser
		parts = append(parts,
			styles.Body.Render("Your code: ")+styles.Title.Render(s.deviceAuth.UserCode),
			"",
		)

		// URL
		parts = append(parts,
			styles.Body.Render("Visit this URL to sign in:"),
			styles.URL.Render(s.deviceAuth.VerificationURIComplete),
			"",
		)

		// Instruction
		parts = append(parts,
			styles.Body.Render("Confirm the code matches, then click \"Confirm\" in your browser."),
			"",
		)

		// Action hint or status
		if s.polling {
			parts = append(parts, s.spinner.View()+" "+mutedStyle.Render("Waiting for authentication..."))
		} else if s.openFailed {
			parts = append(parts, styles.Error.Render("Couldn't open browser. Press 'c' to copy URL"))
		} else if s.copiedToClipboard {
			parts = append(parts, styles.Success.Render("URL copied to clipboard"))
		} else {
			parts = append(parts, styles.Action.Render("Press Enter to open in browser, or press 'c' to copy the URL"))
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

	case stateComplete:
		return styles.Title.Render("Authentication successful!")

	default:
		return ""
	}
}

// isRecoverableError checks if the error allows the user to retry
func isRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "device code expired - press 'r' to restart" ||
		msg == "user denied authorization"
}

// SetSize sets the width available for rendering
func (s *AuthenticateStep) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns true while waiting for authentication
func (s *AuthenticateStep) IsBusy() bool {
	return s.state == stateInitializing || s.polling
}

// HasError returns true if there was an error during authentication
func (s *AuthenticateStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *AuthenticateStep) Error() error {
	return s.err
}

// Help returns key bindings for the auth step
func (s *AuthenticateStep) Help() help.KeyMap {
	if s.state == stateReady {
		keys := []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "open in browser"),
			),
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "copy URL"),
			),
		}

		// Add restart option if there's a recoverable error
		if s.err != nil && isRecoverableError(s.err) {
			keys = append(keys, key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "restart authentication"),
			))
		}

		return keymap.Simple{Keys: keys}
	}

	return keymap.Simple{
		Keys: []key.Binding{},
	}
}

// Next returns the next step after auth completes (role selection)
// Creates an authenticated API client and passes it to the role step
func (s *AuthenticateStep) Next() (step.Step, error) {
	if s.state != stateComplete {
		return nil, step.ErrNotReady
	}

	// Create authenticated API client with the access token from auth result
	refreshFunc := func() (string, error) {
		return s.authService.GetAccessToken(s.ctx)
	}
	apiClient := client.New(s.apiEndpoint, s.authResult.AccessToken, refreshFunc)

	// Create workspace service for onboarding flow
	workspaceService := api.NewWorkspaceService(apiClient, s.logger)

	// Pass authenticated client, preferences, auth, and other dependencies to next step
	return role.NewSelectStep(s.ctx, s.theme, workspaceService, apiClient, s.preferences, s.authService, s.logger, s.globalBindings), nil
}

// Close releases any resources held by the step.
func (s *AuthenticateStep) Close() error {
	return nil
}
