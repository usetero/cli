package auth

import (
	"context"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

type authState int

const (
	stateInitializing authState = iota
	stateReady
	statePolling
	stateComplete
)

// deviceAuthMsg is sent when device auth flow starts.
type deviceAuthMsg struct {
	deviceAuth *auth.DeviceAuth
	err        error
}

// authCompleteMsg is sent when auth completes.
type authCompleteMsg struct {
	err error
}

// AuthenticateModel handles the device code authentication flow.
type AuthenticateModel struct {
	ctx    context.Context
	theme  *styles.Theme
	auth   auth.Auth
	scope  log.Scope
	state  authState
	device *auth.DeviceAuth
	err    error

	spinner           spinner.Model
	copiedToClipboard bool
	browserFailed     bool
	width             int
	height            int
}

// NewAuthenticate creates a new authenticate step.
func NewAuthenticate(ctx context.Context, theme *styles.Theme, authService auth.Auth, scope log.Scope) *AuthenticateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return &AuthenticateModel{
		ctx:     ctx,
		theme:   theme,
		auth:    authService,
		scope:   scope,
		state:   stateInitializing,
		spinner: sp,
	}
}

// Init starts the device auth flow.
func (m *AuthenticateModel) Init() tea.Cmd {
	m.scope.Info("starting device auth flow")
	return tea.Batch(m.spinner.Tick, m.startDeviceAuth())
}

func (m *AuthenticateModel) startDeviceAuth() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		deviceAuth, err := m.auth.StartDeviceAuth(ctx)
		return deviceAuthMsg{deviceAuth: deviceAuth, err: err}
	}
}

func (m *AuthenticateModel) pollForAuth() tea.Cmd {
	return func() tea.Msg {
		interval := time.Duration(m.device.Interval) * time.Second
		_, err := m.auth.WaitForAuth(m.ctx, m.device.DeviceCode, interval)
		return authCompleteMsg{err: err}
	}
}

// Update handles messages.
func (m *AuthenticateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case deviceAuthMsg:
		if msg.err != nil {
			m.scope.Error("device auth failed", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to start authentication", msg.err, false)
		}
		m.device = msg.deviceAuth
		m.state = stateReady
		m.scope.Debug("device code received", "code", m.device.UserCode)

		if err := browser.OpenURL(m.device.VerificationURIComplete); err != nil {
			m.scope.Warn("failed to open browser", "error", err)
			m.browserFailed = true
			return nil
		}
		m.scope.Debug("browser opened")
		m.state = statePolling
		return m.pollForAuth()

	case authCompleteMsg:
		if msg.err != nil {
			m.scope.Error("authentication failed", "error", msg.err)
			m.err = msg.err
			m.state = stateReady
			return appmsg.ErrorCmd("Authentication failed", msg.err, false)
		}
		m.state = stateComplete
		m.scope.Info("user authenticated")
		return func() tea.Msg {
			return msgs.Authenticated{}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd

	case tea.KeyPressMsg:
		if m.state != stateReady {
			return nil
		}
		switch msg.String() {
		case "enter":
			if err := browser.OpenURL(m.device.VerificationURIComplete); err != nil {
				m.scope.Warn("failed to open browser", "error", err)
				m.browserFailed = true
				return nil
			}
			m.scope.Debug("browser opened")
			m.browserFailed = false
			m.state = statePolling
			return m.pollForAuth()

		case "c":
			if err := clipboard.WriteAll(m.device.VerificationURIComplete); err != nil {
				m.scope.Warn("failed to copy to clipboard", "error", err)
				return nil
			}
			m.scope.Debug("url copied to clipboard")
			m.copiedToClipboard = true
			m.state = statePolling
			return m.pollForAuth()

		case "r":
			if m.err != nil {
				m.scope.Debug("retrying authentication")
				m.err = nil
				m.state = stateInitializing
				return m.Init()
			}
		}
	}
	return nil
}

// View renders the authenticate UI.
func (m *AuthenticateModel) View() string {
	s := m.theme.Styles
	colors := m.theme.Colors

	if m.state == stateInitializing {
		return s.Title.Render("Initializing authentication...")
	}

	if m.device == nil {
		return s.Title.Render("Loading...")
	}

	var parts []string
	parts = append(parts, s.Title.Render("Authenticate with Tero"), "")
	parts = append(parts,
		s.Body.Render("Your code: ")+s.Title.Render(m.device.UserCode),
		"",
	)
	parts = append(parts,
		s.Body.Render("Visit this URL to sign in:"),
		s.URL.Render(m.device.VerificationURIComplete),
		"",
	)
	parts = append(parts,
		s.Body.Render("Confirm the code matches, then click \"Confirm\" in your browser."),
		"",
	)

	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch {
	case m.state == statePolling:
		parts = append(parts, m.spinner.View()+" "+mutedStyle.Render("Waiting for authentication..."))
	case m.browserFailed:
		parts = append(parts, s.Error.Render("Couldn't open browser. Press 'c' to copy URL"))
	case m.copiedToClipboard:
		parts = append(parts, s.Success.Render("URL copied to clipboard"))
	case m.err != nil:
		parts = append(parts, s.Error.Render("Authentication failed. Press 'r' to retry"))
	default:
		parts = append(parts, s.Action.Render("Press Enter to open in browser, or 'c' to copy URL"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *AuthenticateModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
