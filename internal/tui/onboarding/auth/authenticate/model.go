package authenticate

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
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// state tracks the current state of the authentication flow.
type state int

const (
	stateInitializing state = iota
	stateReady
	stateComplete
)

// deviceAuthMsg is sent when device authorization is initiated.
type deviceAuthMsg struct {
	deviceAuth *auth.DeviceAuth
	err        error
}

// authCompleteMsg is sent when authentication completes.
type authCompleteMsg struct {
	result *auth.Result
	err    error
}

// Model handles device code flow authentication.
type Model struct {
	ctx         context.Context
	theme       *styles.Theme
	auth        auth.Auth
	prefs       preferences.Preferences
	apiEndpoint string
	logger      log.Logger

	state             state
	deviceAuth        *auth.DeviceAuth
	authResult        *auth.Result
	err               error
	polling           bool
	openFailed        bool
	copiedToClipboard bool
	spinner           spinner.Model
	width             int
	height            int
}

// New creates a new authenticate model.
func New(ctx context.Context, theme *styles.Theme, authService auth.Auth, prefs preferences.Preferences, apiEndpoint string, logger log.Logger) Model {
	colors := theme.Colors

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return Model{
		ctx:         ctx,
		theme:       theme,
		auth:        authService,
		prefs:       prefs,
		apiEndpoint: apiEndpoint,
		logger:      logger,
		state:       stateInitializing,
		spinner:     sp,
		width:       80,
	}
}

// Init initializes the auth step by starting device authorization.
func (m Model) Init() tea.Cmd {
	if m.auth.IsAuthenticated() {
		m.logger.Info("already authenticated")
		// Note: state change here won't persist since Init returns nil,
		// but the Next() method checks auth.IsAuthenticated() directly
		return nil
	}

	return tea.Batch(
		m.spinner.Tick,
		m.startDeviceAuth(),
	)
}

// startDeviceAuth begins the device authorization flow.
func (m Model) startDeviceAuth() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		deviceAuth, err := m.auth.StartDeviceAuth(ctx)
		return deviceAuthMsg{deviceAuth: deviceAuth, err: err}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case deviceAuthMsg:
		if msg.err != nil {
			m.logger.Error("failed to start device authorization", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

		m.deviceAuth = msg.deviceAuth
		m.state = stateReady
		m.err = nil
		m.logger.Info("device authorization started")

		// Auto-open browser
		err := browser.OpenURL(m.deviceAuth.VerificationURIComplete)
		if err != nil {
			m.logger.Warn("failed to auto-open browser", "error", err)
			m.openFailed = true
			return m, nil
		}

		m.logger.Debug("auto-opened browser for auth")
		m.polling = true
		return m, m.pollForAuth()

	case authCompleteMsg:
		m.polling = false

		if msg.err != nil {
			m.logger.Error("authentication failed", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

		m.authResult = msg.result
		m.state = stateComplete
		m.logger.Info("authentication complete", "user_email", m.authResult.User.Email)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if m.state == stateReady {
			switch msg.String() {
			case "enter":
				err := browser.OpenURL(m.deviceAuth.VerificationURIComplete)
				if err != nil {
					m.logger.Warn("failed to open browser", "error", err)
					m.openFailed = true
					return m, nil
				}

				m.openFailed = false
				m.err = nil
				if !m.polling {
					m.polling = true
					return m, m.pollForAuth()
				}
				return m, nil

			case "c":
				err := clipboard.WriteAll(m.deviceAuth.VerificationURIComplete)
				if err != nil {
					m.logger.Error("failed to copy to clipboard", "error", err)
					return m, nil
				}
				m.copiedToClipboard = true
				m.logger.Debug("auth URL copied to clipboard")

				m.err = nil
				if !m.polling {
					m.polling = true
					return m, m.pollForAuth()
				}
				return m, nil

			case "r":
				if m.err != nil && isRecoverableError(m.err) {
					m.polling = false
					m.err = nil
					m.openFailed = false
					m.copiedToClipboard = false
					m.state = stateInitializing
					return m, m.Init()
				}
			}
		}
	}

	return m, nil
}

// pollForAuth starts the background polling process.
func (m Model) pollForAuth() tea.Cmd {
	return func() tea.Msg {
		interval := time.Duration(m.deviceAuth.Interval) * time.Second
		result, err := m.auth.WaitForAuth(m.ctx, m.deviceAuth.DeviceCode, interval)
		return authCompleteMsg{result: result, err: err}
	}
}

// View renders the auth step.
func (m Model) View() string {
	styles := m.theme.Styles
	colors := m.theme.Colors

	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch m.state {
	case stateInitializing:
		return styles.Title.Render("Initializing authentication...")

	case stateReady:
		if m.deviceAuth == nil {
			return styles.Title.Render("Loading...")
		}

		var parts []string

		parts = append(parts, styles.Title.Render("Authenticate with Tero"), "")
		parts = append(parts,
			styles.Body.Render("Your code: ")+styles.Title.Render(m.deviceAuth.UserCode),
			"",
		)
		parts = append(parts,
			styles.Body.Render("Visit this URL to sign in:"),
			styles.URL.Render(m.deviceAuth.VerificationURIComplete),
			"",
		)
		parts = append(parts,
			styles.Body.Render("Confirm the code matches, then click \"Confirm\" in your browser."),
			"",
		)

		if m.polling {
			parts = append(parts, m.spinner.View()+" "+mutedStyle.Render("Waiting for authentication..."))
		} else if m.openFailed {
			parts = append(parts, styles.Error.Render("Couldn't open browser. Press 'c' to copy URL"))
		} else if m.copiedToClipboard {
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

// isRecoverableError checks if the error allows the user to retry.
func isRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "device code expired - press 'r' to restart" ||
		msg == "user denied authorization"
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns true while waiting for authentication.
func (m Model) IsBusy() bool {
	return m.state == stateInitializing || m.polling
}

// HasError returns true if there was an error.
func (m Model) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m Model) Error() error {
	return m.err
}

// Help returns key bindings for the auth step.
func (m Model) Help() help.KeyMap {
	if m.state == stateReady {
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

		if m.err != nil && isRecoverableError(m.err) {
			keys = append(keys, key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "restart"),
			))
		}

		return keymap.Simple{Keys: keys}
	}

	return keymap.Simple{Keys: []key.Binding{}}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.state != stateComplete {
		return nil, step.ErrNotReady
	}

	// Create API services - client is created internally with auth.Auth for fresh tokens per request
	services := api.NewServices(m.apiEndpoint+"/graphql", m.auth, m.logger)

	return role.New(m.ctx, m.theme, services, m.prefs, m.auth, m.logger), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
