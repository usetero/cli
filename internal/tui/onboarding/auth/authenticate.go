package auth

import (
	"context"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/pkg/browser"
	authservice "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
	"github.com/usetero/cli/pkg/client"
)

// Authenticator defines the interface for authentication operations.
// Consumer-driven interface - this step only needs these methods.
type Authenticator interface {
	IsAuthenticated() bool
	StartDeviceAuth(ctx context.Context) (*authservice.DeviceAuth, error)
	WaitForAuth(ctx context.Context, deviceCode string, interval time.Duration) (*authservice.Result, error)
}

// authState tracks the current state of the authentication flow
type authState int

const (
	stateInitializing authState = iota
	stateReady                  // Ready for user to open browser or copy URL
	stateComplete
)

// AuthenticateStep handles device code flow authentication.
type AuthenticateStep struct {
	// Services (defined by consumer interfaces)
	authenticator Authenticator

	// Pass-through to next step
	preferencesService *preferences.Service
	apiEndpoint        string
	logger             log.Logger
	globalBindings     []key.Binding

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
func NewAuthenticateStep(logger log.Logger, authenticator Authenticator, preferencesService *preferences.Service, apiEndpoint string, globalBindings []key.Binding) step.Step {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if authenticator == nil {
		panic("authenticator cannot be nil")
	}
	if preferencesService == nil {
		panic("preferencesService cannot be nil")
	}

	theme := styles.CurrentTheme()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &AuthenticateStep{
		authenticator:      authenticator,
		preferencesService: preferencesService,
		apiEndpoint:        apiEndpoint,
		logger:             logger,
		globalBindings:     globalBindings,
		state:              stateInitializing,
		spinner:            sp,
	}
}

// Init initializes the auth step by starting device authorization
func (s *AuthenticateStep) Init() tea.Cmd {
	// Check if already authenticated
	if s.authenticator.IsAuthenticated() {
		s.logger.Info("already authenticated")
		s.state = stateComplete
		return nil
	}

	// Start device authorization and spinner
	return tea.Batch(
		s.spinner.Tick,
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			deviceAuth, err := s.authenticator.StartDeviceAuth(ctx)
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
		return s, nil

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
		ctx := context.Background()
		interval := time.Duration(s.deviceAuth.Interval) * time.Second

		result, err := s.authenticator.WaitForAuth(ctx, s.deviceAuth.DeviceCode, interval)
		return authCompleteMsg{result: result, err: err}
	}
}

// View renders the auth step
func (s *AuthenticateStep) View() string {
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	urlStyle := lipgloss.NewStyle().
		Foreground(theme.TextSubtle)

	actionStyle := lipgloss.NewStyle().
		Foreground(theme.Primary)

	mutedStyle := lipgloss.NewStyle().
		Foreground(theme.TextMuted)

	errorStyle := lipgloss.NewStyle().Foreground(theme.Error)
	successStyle := lipgloss.NewStyle().Foreground(theme.Success)

	switch s.state {
	case stateInitializing:
		return titleStyle.Render("Initializing authentication...")

	case stateReady:
		if s.deviceAuth == nil {
			return titleStyle.Render("Loading...")
=======
	common := styles.Common()
	theme := styles.CurrentTheme()

	mutedStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)

	switch s.state {
	case stateInitializing:
		return common.Title.Render("Initializing authentication...")

	case stateReady:
		if s.deviceAuth == nil {
			return common.Title.Render("Loading...")
>>>>>>> 17e8dd9 (chore: initial commit)
		}

		var parts []string

		// Title
<<<<<<< HEAD
		parts = append(parts, titleStyle.Render("Authenticate with Tero"), "")

		// URL
		parts = append(parts,
			subtitleStyle.Render("Visit this URL to sign in:"),
			urlStyle.Render(s.deviceAuth.VerificationURIComplete),
=======
		parts = append(parts, common.Title.Render("Authenticate with Tero"), "")

		// URL
		parts = append(parts,
			common.Body.Render("Visit this URL to sign in:"),
			common.URL.Render(s.deviceAuth.VerificationURIComplete),
>>>>>>> 17e8dd9 (chore: initial commit)
			"",
		)

		// Action hint or status
		if s.polling {
			parts = append(parts, s.spinner.View()+" "+mutedStyle.Render("Waiting for authentication..."))
		} else if s.openFailed {
<<<<<<< HEAD
			parts = append(parts, errorStyle.Render("Couldn't open browser. Press 'c' to copy URL"))
		} else if s.copiedToClipboard {
			parts = append(parts, successStyle.Render("✓ URL copied to clipboard"))
		} else {
			parts = append(parts, actionStyle.Render("Press Enter to open in browser, or press 'c' to copy the URL"))
=======
			parts = append(parts, common.Error.Render("Couldn't open browser. Press 'c' to copy URL"))
		} else if s.copiedToClipboard {
			parts = append(parts, common.Success.Render("✓ URL copied to clipboard"))
		} else {
			parts = append(parts, common.Action.Render("Press Enter to open in browser, or press 'c' to copy the URL"))
>>>>>>> 17e8dd9 (chore: initial commit)
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

	case stateComplete:
<<<<<<< HEAD
		return titleStyle.Render("✓ Authentication successful!")
=======
		return common.Title.Render("✓ Authentication successful!")
>>>>>>> 17e8dd9 (chore: initial commit)

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
	return msg == "device code expired - please restart authentication" ||
		msg == "user denied authorization"
}

// SetSize sets the width available for rendering
func (s *AuthenticateStep) SetSize(width, height int) {
	s.width = width
}

// IsComplete returns true if authentication is complete
func (s *AuthenticateStep) IsComplete() bool {
	return s.state == stateComplete
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
func (s *AuthenticateStep) Next() step.Step {
	// Create authenticated API client with the access token from auth result
	apiClient := client.New(s.apiEndpoint, s.authResult.AccessToken)

	// Pass authenticated client, preferences service, and other dependencies to next step
	return role.NewSelectStep(apiClient, s.preferencesService, s.logger, s.globalBindings)
}
