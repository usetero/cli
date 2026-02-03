package organizationcreate

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// createOrgMsg is sent when org creation completes.
type createOrgMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}

// tokenRefreshMsg is sent when token refresh completes.
type tokenRefreshMsg struct {
	accessToken string
	err         error
}

// Model handles creating a new organization.
type Model struct {
	ctx   context.Context
	theme *styles.Theme
	role  string

	services api.APIServices
	prefs    preferences.Preferences
	auth     auth.Auth
	logger   log.Logger

	input           input.Model
	creating        bool
	created         bool
	createdResult   *api.OrganizationBootstrapResult
	refreshingToken bool
	tokenRefreshed  bool
	err             error
	width           int
	height          int
}

// New creates a new organization create model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	services api.APIServices,
	prefs preferences.Preferences,
	authService auth.Auth,
	logger log.Logger,
) Model {
	inp := input.New(theme, logger).
		SetPlaceholder("Acme Inc.").
		SetCharLimit(100)

	return Model{
		ctx:      ctx,
		theme:    theme,
		role:     role,
		services: services,
		prefs:    prefs,
		auth:     authService,
		logger:   logger,
		input:    inp,
		width:    80,
	}
}

// Init initializes the step.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.err != nil {
				m.err = nil
				m.creating = false
				return m, cmd
			}

			if !m.creating && !m.created {
				name := m.input.Value()
				if name != "" {
					m.creating = true
					return m, tea.Batch(cmd, m.createOrganization(name))
				}
			}
		}

	case createOrgMsg:
		m.creating = false
		if msg.err != nil {
			m.logger.Error("failed to create organization", "error", msg.err)
			m.err = msg.err
			return m, cmd
		}

		m.logger.Info("organization created", "id", msg.result.Organization.ID, "name", msg.result.Organization.Name)
		m.createdResult = msg.result

		if err := m.prefs.SetDefaultOrgID(msg.result.Organization.ID); err != nil {
			m.logger.Error("failed to save org preference", "error", err)
			m.err = err
			return m, cmd
		}

		if err := m.prefs.SetDefaultAccountID(msg.result.Account.ID); err != nil {
			m.logger.Error("failed to save account preference", "error", err)
			m.err = err
			return m, cmd
		}

		if msg.result.Organization.WorkosOrganizationID != "" {
			m.refreshingToken = true
			return m, tea.Batch(cmd, m.refreshToken(msg.result.Organization.WorkosOrganizationID))
		}

		m.created = true
		m.tokenRefreshed = true
		return m, cmd

	case tokenRefreshMsg:
		m.refreshingToken = false
		m.tokenRefreshed = true
		if msg.err != nil {
			m.logger.Error("failed to refresh token with organization", "error", msg.err)
		} else {
			m.logger.Debug("token refreshed with organization scope")
			// Token is now stored in auth service - subsequent API calls will use it automatically
		}
		m.created = true
		return m, cmd
	}

	return m, cmd
}

// createOrganization returns a command that creates the organization.
func (m Model) createOrganization(name string) tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("creating organization", "name", name)

		result, err := m.services.Organizations.Create(m.ctx, uuid.New(), name)
		if err != nil {
			return createOrgMsg{err: err}
		}

		return createOrgMsg{result: result}
	}
}

// refreshToken returns a command that refreshes the token with org scope.
func (m Model) refreshToken(workosOrgID domain.WorkosOrganizationID) tea.Cmd {
	return func() tea.Msg {
		accessToken, err := m.auth.RefreshTokenWithOrganization(m.ctx, workosOrgID)
		return tokenRefreshMsg{accessToken: accessToken, err: err}
	}
}

// View renders the create organization UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.creating {
		return themeStyles.Title.Render("Creating organization...")
	}

	title := themeStyles.Title.Render("Create a new organization")
	prompt := themeStyles.Body.Render("Enter your organization name")
	help := themeStyles.Help.Render("This will be your default workspace")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		"",
		m.input.View(),
		"",
		help,
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	if width > 10 {
		m.input = m.input.SetWidth(width - 6)
	}
	return m
}

// IsBusy returns true while creating or refreshing token.
func (m Model) IsBusy() bool {
	return m.creating || m.refreshingToken
}

// HasError returns true if there's an error.
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
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.created || !m.tokenRefreshed {
		return nil, step.ErrNotReady
	}

	// TODO: return datadog or workspace step
	// Bootstrap creates org + account + workspace, so we skip account selection
	return nil, nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
