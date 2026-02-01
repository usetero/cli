package organization

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// CreateStep handles creating a new organization
type CreateStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role string

	// Services
	organizations api.Organizations
	preferences   preferences.Preferences
	auth          auth.Auth

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	input           *input.Component
	creating        bool
	created         bool
	createdResult   *api.OrganizationBootstrapResult
	refreshingToken bool
	tokenRefreshed  bool
	err             error
	width           int
	globalBindings  []key.Binding
}

// NewCreateStep creates a new organization creation step
func NewCreateStep(ctx context.Context, theme *styles.Theme, role string, organizations api.Organizations, prefs preferences.Preferences, authService auth.Auth, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if organizations == nil {
		panic("organizations cannot be nil")
	}
	if prefs == nil {
		panic("preferences cannot be nil")
	}
	if authService == nil {
		panic("auth cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	inp := input.New(theme, logger)
	inp.SetPlaceholder("Acme Inc.")
	inp.SetCharLimit(100)

	return &CreateStep{
		ctx:            ctx,
		theme:          theme,
		role:           role,
		organizations:  organizations,
		preferences:    prefs,
		auth:           authService,
		apiClient:      apiClient,
		logger:         logger,
		input:          inp,
		width:          80,
		globalBindings: globalBindings,
	}
}

// createOrgMsg is sent when org creation completes
type createOrgMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}

// createTokenRefreshMsg is sent when token refresh completes after org creation
type createTokenRefreshMsg struct {
	accessToken string
	err         error
}

// Init focuses the input
func (s *CreateStep) Init() tea.Cmd {
	return nil // Input is already focused in constructor
}

// Update handles messages
func (s *CreateStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	// Always update input for cursor blinking
	inputCmd := s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Retry on error
			if s.err != nil {
				s.err = nil
				s.creating = false
				return s, inputCmd
			}

			// Submit if not already creating
			if !s.creating && !s.created {
				name := s.input.Value()
				if name != "" {
					s.creating = true
					return s, tea.Batch(inputCmd, s.createOrganization(name))
				}
			}
		}

	case createOrgMsg:
		s.creating = false
		if msg.err != nil {
			s.logger.Error("failed to create organization", "error", msg.err)
			s.err = msg.err
			return s, inputCmd
		}

		s.logger.Info("organization created", "id", msg.result.Organization.ID, "name", msg.result.Organization.Name, "accountID", msg.result.Account.ID)
		s.createdResult = msg.result

		// Save organization to preferences
		if err := s.preferences.SetDefaultOrgID(msg.result.Organization.ID); err != nil {
			s.logger.Error("failed to save org preference", "error", err)
			s.err = err
			return s, inputCmd
		}
		s.logger.Debug("organization saved to preferences", "orgID", msg.result.Organization.ID)

		// Organization bootstrap also creates an account - save it too
		if err := s.preferences.SetDefaultAccountID(msg.result.Account.ID); err != nil {
			s.logger.Error("failed to save account preference", "error", err)
			s.err = err
			return s, inputCmd
		}
		s.logger.Debug("account saved to preferences", "accountID", msg.result.Account.ID)

		// Refresh token with organization scope
		if msg.result.Organization.WorkosOrganizationID != "" {
			s.refreshingToken = true
			return s, tea.Batch(inputCmd, s.refreshToken(msg.result.Organization.WorkosOrganizationID))
		}

		// No workos org ID - skip refresh, mark complete
		s.created = true
		s.tokenRefreshed = true
		return s, inputCmd

	case createTokenRefreshMsg:
		s.refreshingToken = false
		s.tokenRefreshed = true
		if msg.err != nil {
			s.logger.Error("failed to refresh token with organization", "error", msg.err)
			// Continue anyway - token refresh is best-effort
		} else {
			s.logger.Debug("token refreshed with organization scope")
			// Update the API client with the new token
			s.apiClient.SetAccessToken(msg.accessToken)
		}
		s.created = true
		return s, inputCmd
	}

	return s, inputCmd
}

// refreshToken returns a command that refreshes the token with org scope
func (s *CreateStep) refreshToken(workosOrgID string) tea.Cmd {
	return func() tea.Msg {
		accessToken, err := s.auth.RefreshTokenWithOrganization(s.ctx, workosOrgID)
		return createTokenRefreshMsg{accessToken: accessToken, err: err}
	}
}

// createOrganization creates a new organization via the API
func (s *CreateStep) createOrganization(name string) tea.Cmd {
	return func() tea.Msg {
		s.logger.Info("creating organization", log.String("name", name))

		result, err := s.organizations.Create(s.ctx, name)
		if err != nil {
			return createOrgMsg{err: err}
		}

		return createOrgMsg{result: result}
	}
}

// View renders the create organization UI
func (s *CreateStep) View() string {
	themeStyles := s.theme.Styles

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			themeStyles.Title.Render("Creating organization..."),
		)
	}

	title := themeStyles.Title.Render("Create a new organization")
	prompt := themeStyles.Body.Render("Enter your organization name")
	help := themeStyles.Help.Render("This will be your default workspace")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		"",
		s.input.View(),
		"",
		help,
	)

	return content
}

// SetSize sets the width available for rendering
func (s *CreateStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsBusy returns true while creating the organization or refreshing token
func (s *CreateStep) IsBusy() bool {
	return s.creating || s.refreshingToken
}

// HasError returns true if there was an error creating the organization
func (s *CreateStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *CreateStep) Error() error {
	return s.err
}

// Next returns the next step after creating organization
func (s *CreateStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.created || !s.tokenRefreshed {
		return nil, step.ErrNotReady
	}

	// Skip account selection since bootstrap creates it automatically
	// Go to Datadog region selection
	return datadog.NewSelectRegionStep(s.ctx, s.theme, s.role, *s.createdResult.Organization, *s.createdResult.Account, s.apiClient, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *CreateStep) Help() help.KeyMap {
	// Show retry hint if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	// Normal state: show submit
	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
		},
	}
}

// Close releases any resources held by the step.
func (s *CreateStep) Close() error {
	return nil
}
