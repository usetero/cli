package check

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/auth/authenticate"
	"github.com/usetero/cli/internal/tui/onboarding/role"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// checkAuthMsg is sent when check completes.
type checkAuthMsg struct {
	hasValidAuth bool
	accessToken  string
	err          error
}

// Model checks if the user has a valid auth token.
type Model struct {
	ctx         context.Context
	theme       *styles.Theme
	auth        auth.Auth
	prefs       preferences.Preferences
	apiEndpoint string
	logger      log.Logger

	checking     bool
	checked      bool
	hasValidAuth bool
	accessToken  string
	err          error
	width        int
	height       int
}

// New creates a new auth check model.
func New(ctx context.Context, theme *styles.Theme, authService auth.Auth, prefs preferences.Preferences, apiEndpoint string, logger log.Logger) Model {
	return Model{
		ctx:         ctx,
		theme:       theme,
		auth:        authService,
		prefs:       prefs,
		apiEndpoint: apiEndpoint,
		logger:      logger,
		width:       80,
	}
}

// Init starts checking for valid auth.
func (m Model) Init() tea.Cmd {
	m.checking = true
	return m.checkAuth()
}

// checkAuth checks if there's a valid access token.
func (m Model) checkAuth() tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("checking authentication")

		if !m.auth.IsAuthenticated() {
			return checkAuthMsg{hasValidAuth: false}
		}

		accessToken, err := m.auth.GetAccessToken(m.ctx)
		if err != nil {
			m.logger.Warn("failed to get access token, clearing tokens", "error", err)
			_ = m.auth.ClearTokens()
			return checkAuthMsg{hasValidAuth: false}
		}

		return checkAuthMsg{hasValidAuth: true, accessToken: accessToken}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case checkAuthMsg:
		m.checking = false
		if msg.err != nil {
			m.logger.Error("failed to check auth", "error", msg.err)
			m.err = msg.err
			return m, nil
		}

		m.err = nil
		m.checked = true
		m.hasValidAuth = msg.hasValidAuth
		m.accessToken = msg.accessToken

		if m.hasValidAuth {
			m.logger.Info("valid authentication found")
		} else {
			m.logger.Info("no valid authentication found")
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.err != nil {
				m.logger.Info("retrying authentication check")
				m.err = nil
				m.checking = true
				return m, m.checkAuth()
			}
		}
	}

	return m, nil
}

// View renders the check UI.
func (m Model) View() string {
	styles := m.theme.Styles

	if m.checking {
		return styles.Title.Render("Checking authentication...")
	}

	if m.hasValidAuth {
		return styles.Success.Render("Already authenticated")
	}

	return styles.Title.Render("No authentication found")
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns true while checking.
func (m Model) IsBusy() bool {
	return m.checking
}

// HasError returns true if there was an error.
func (m Model) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m Model) Error() error {
	return m.err
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}
	return keymap.Simple{Keys: []key.Binding{}}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.checked {
		return nil, step.ErrNotReady
	}

	if !m.hasValidAuth {
		// No valid auth - go to authenticate step
		return authenticate.New(m.ctx, m.theme, m.auth, m.prefs, m.apiEndpoint, m.logger), nil
	}

	// Has valid auth - create API services and go to role selection
	services := api.NewServices(m.apiEndpoint+"/graphql", m.auth, m.logger)

	return role.New(m.ctx, m.theme, services, m.prefs, m.auth, m.logger), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
