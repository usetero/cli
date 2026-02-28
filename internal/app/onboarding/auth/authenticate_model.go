package auth

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// AuthenticateModel handles the device code authentication flow.
type AuthenticateModel struct {
	ctx    context.Context
	theme  styles.Theme
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
func NewAuthenticate(ctx context.Context, theme styles.Theme, authService auth.Auth, scope log.Scope) *AuthenticateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg)

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
		interval := authPollInterval(m.device.Interval)
		m.scope.Debug("starting auth polling", "provider_interval_seconds", m.device.Interval, "poll_interval", interval)
		result, err := m.auth.WaitForAuth(m.ctx, m.device.DeviceCode, interval)
		return authCompleteMsg{result: result, err: err}
	}
}

func authPollInterval(providerIntervalSeconds int) time.Duration {
	interval := time.Duration(providerIntervalSeconds) * time.Second
	if interval <= 0 {
		return defaultPollInterval
	}
	if interval < minPollInterval {
		return minPollInterval
	}
	if interval > maxPollInterval {
		return maxPollInterval
	}
	return interval
}

// SetSize updates dimensions.
func (m *AuthenticateModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns the key bindings for the short help view.
func (m *AuthenticateModel) ShortHelp() []key.Binding {
	if m.state != stateReady {
		return nil
	}
	if m.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open browser")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy URL")),
	}
}
