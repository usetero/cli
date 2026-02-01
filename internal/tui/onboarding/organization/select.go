package organization

import (
	"context"
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/loader"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const createNewOrgID = "__CREATE_NEW__"

// OrgItem implements list.Item for the list component.
// Exported for testing.
type OrgItem struct {
	ID                   string
	Name                 string
	WorkosOrganizationID string
}

func (i OrgItem) FilterValue() string { return i.Name }

// orgDelegate renders each organization in the list
type orgDelegate struct {
	theme *styles.Theme
}

func (d orgDelegate) Height() int                             { return 1 }
func (d orgDelegate) Spacing() int                            { return 0 }
func (d orgDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d orgDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(OrgItem)
	if !ok {
		return
	}

	colors := d.theme.Colors

	str := i.Name
	if index == m.Index() {
		fn := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			Render
		_, _ = fmt.Fprint(w, fn("> "+str))
	} else {
		fn := lipgloss.NewStyle().
			Foreground(colors.Page.Text).
			Render
		_, _ = fmt.Fprint(w, fn("  "+str))
	}
}

// SelectStep handles selecting an organization or choosing to create one.
type SelectStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme
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
	remoteList          *remotelist.Component
	orgs                []api.Organization
	selectedOrgID       string
	selectedWorkosOrgID string
	tokenRefreshed      bool
	refreshingToken     bool
	refreshLoader       *loader.Component
	width               int
	globalBindings      []key.Binding
}

// tokenRefreshMsg is sent when token refresh completes
type tokenRefreshMsg struct {
	accessToken string
	err         error
}

// NewSelectStep creates a new organization selection step
func NewSelectStep(ctx context.Context, theme *styles.Theme, role string, organizations api.Organizations, apiClient api.Client, prefs preferences.Preferences, authService auth.Auth, logger log.Logger, globalBindings []key.Binding) step.Step {
	if organizations == nil {
		panic("organizations cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if prefs == nil {
		panic("preferences cannot be nil")
	}
	if authService == nil {
		panic("auth cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	delegate := orgDelegate{theme: theme}
	remoteList := remotelist.New(theme, delegate, "Loading organizations", logger)

	return &SelectStep{
		ctx:            ctx,
		theme:          theme,
		role:           role,
		organizations:  organizations,
		preferences:    prefs,
		auth:           authService,
		apiClient:      apiClient,
		logger:         logger,
		remoteList:     remoteList,
		width:          80,
		globalBindings: globalBindings,
	}
}

// Init starts loading organizations
func (s *SelectStep) Init() tea.Cmd {
	return s.remoteList.InitWithLoader(func() tea.Msg {
		s.logger.Info("loading organizations")

		orgs, err := s.organizations.List(s.ctx)
		if err != nil {
			s.logger.Error("failed to load organizations", "error", err)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		s.logger.Info("organizations loaded", log.Int("count", len(orgs)))

		// Build list items from orgs
		items := make([]list.Item, len(orgs))
		for i, org := range orgs {
			items[i] = OrgItem{ID: org.ID, Name: org.Name, WorkosOrganizationID: org.WorkosOrganizationID}
		}

		return remotelist.LoadResultMsg{Items: items, Err: nil}
	})
}

// refreshToken returns a command that refreshes the token with org scope
func (s *SelectStep) refreshToken() tea.Cmd {
	return func() tea.Msg {
		accessToken, err := s.auth.RefreshTokenWithOrganization(s.ctx, s.selectedWorkosOrgID)
		return tokenRefreshMsg{accessToken: accessToken, err: err}
	}
}

// selectOrg handles org selection - sets state and returns refresh command if needed
func (s *SelectStep) selectOrg(orgID, workosOrgID string) tea.Cmd {
	s.selectedOrgID = orgID
	s.selectedWorkosOrgID = workosOrgID

	// "Create new" doesn't need token refresh
	if orgID == createNewOrgID {
		s.tokenRefreshed = true
		return nil
	}

	// Existing org needs token refresh
	if workosOrgID != "" {
		s.refreshingToken = true

		// Create loader with org name
		orgName := "organization"
		for _, org := range s.orgs {
			if org.ID == orgID {
				orgName = org.Name
				break
			}
		}
		s.refreshLoader = loader.New(s.theme, "Selecting "+orgName)

		return tea.Batch(s.refreshLoader.Init(), s.refreshToken())
	}

	// No workos org ID - skip refresh
	s.tokenRefreshed = true
	return nil
}

// Update handles messages
func (s *SelectStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	// Update loader if refreshing
	if s.refreshingToken && s.refreshLoader != nil {
		loaderCmd := s.refreshLoader.Update(msg)
		cmds = append(cmds, loaderCmd)
	}

	// Handle token refresh completion
	switch msg := msg.(type) {
	case tokenRefreshMsg:
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
		return s, nil
	}

	// Handle remotelist's LoadResultMsg to apply auto-selection logic
	switch msg := msg.(type) {
	case remotelist.LoadResultMsg:
		if msg.Err == nil {
			// Extract orgs from items
			s.orgs = make([]api.Organization, 0, len(msg.Items))
			for _, item := range msg.Items {
				if orgItem, ok := item.(OrgItem); ok {
					s.orgs = append(s.orgs, api.Organization{ID: orgItem.ID, Name: orgItem.Name, WorkosOrganizationID: orgItem.WorkosOrganizationID})
				}
			}

			// Apply auto-selection logic
			userPref := s.preferences.GetDefaultOrgID()

			// Case 1: No orgs → auto-select "create" to fast-forward
			if len(s.orgs) == 0 {
				s.logger.Debug("auto-selected create organization", "reason", "no organizations found")
				cmd := s.selectOrg(createNewOrgID, "")
				cmds = append(cmds, cmd)
			}

			// Case 2: Has preference AND exists → auto-select
			if userPref != "" {
				for _, org := range s.orgs {
					if org.ID == userPref {
						s.logger.Debug("auto-selected organization from preference", "id", userPref, "name", org.Name)
						cmd := s.selectOrg(org.ID, org.WorkosOrganizationID)
						cmds = append(cmds, cmd)
					}
				}
			}

			// Case 3: No preference AND only 1 org → auto-select and save
			if userPref == "" && len(s.orgs) == 1 {
				s.logger.Info("auto-selected organization", "id", s.orgs[0].ID, "name", s.orgs[0].Name, "reason", "only one available")
				if err := s.preferences.SetDefaultOrgID(s.orgs[0].ID); err != nil {
					s.logger.Error("failed to save organization preference", "error", err)
				}
				cmd := s.selectOrg(s.orgs[0].ID, s.orgs[0].WorkosOrganizationID)
				cmds = append(cmds, cmd)
			}

			// Case 4: User must select from list (selectedOrgID remains empty)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Retry loading if there's an error
			if s.remoteList.HasError() {
				s.logger.Info("user requested retry")
				cmd := s.remoteList.Retry()
				cmds = append(cmds, cmd)
			}
		case "enter":
			if !s.remoteList.IsLoaded() {
				break
			}
			selected := s.remoteList.SelectedItem()
			if org, ok := selected.(OrgItem); ok {
				s.logger.Info("organization selected", "id", org.ID, "name", org.Name)
				if err := s.preferences.SetDefaultOrgID(org.ID); err != nil {
					s.logger.Error("failed to save organization preference", "error", err)
				}
				cmd := s.selectOrg(org.ID, org.WorkosOrganizationID)
				cmds = append(cmds, cmd)
			}
		case "n":
			if !s.remoteList.IsLoaded() {
				break
			}
			// User pressed 'n' to create new organization
			s.logger.Info("user chose to create new organization")
			cmd := s.selectOrg(createNewOrgID, "")
			cmds = append(cmds, cmd)
		}
	}

	// Update remote list (handles loading, error, and list navigation)
	cmd := s.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

// View renders the organization selection UI
func (s *SelectStep) View() string {
	themeStyles := s.theme.Styles

	// Show loading state during initial load or token refresh
	if s.remoteList.IsBusy() {
		return s.remoteList.View()
	}

	// Show loading during token refresh (after auto-selection)
	if s.refreshingToken && s.refreshLoader != nil {
		return s.refreshLoader.View()
	}

	// Show error state
	if s.remoteList.HasError() {
		return s.remoteList.View()
	}

	title := themeStyles.Title.Render("Select your organization")
	subtitle := themeStyles.Help.Render("This will be your default workspace")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		s.remoteList.View(),
	)
}

// SetSize sets the width available for rendering
func (s *SelectStep) SetSize(width, height int) {
	s.width = width
	s.remoteList.SetWidth(width)
}

// IsBusy returns true while loading organizations or refreshing token
func (s *SelectStep) IsBusy() bool {
	return !s.remoteList.IsLoaded() || s.refreshingToken
}

// HasError returns true if the remotelist has an error
func (s *SelectStep) HasError() bool {
	return s.remoteList.HasError()
}

// Error returns the remotelist's error, or nil if no error
func (s *SelectStep) Error() error {
	return s.remoteList.Error()
}

// Next returns the next step after organization selection
func (s *SelectStep) Next() (step.Step, error) {
	// Not ready yet
	if s.selectedOrgID == "" || !s.tokenRefreshed {
		return nil, step.ErrNotReady
	}

	// User wants to create new org
	if s.selectedOrgID == createNewOrgID {
		organizationService := api.NewOrganizationService(s.apiClient, s.logger)
		return NewCreateStep(s.ctx, s.theme, s.role, organizationService, s.preferences, s.auth, s.apiClient, s.logger, s.globalBindings), nil
	}

	// Find the selected org
	var selectedOrg api.Organization
	for _, org := range s.orgs {
		if org.ID == s.selectedOrgID {
			selectedOrg = org
			break
		}
	}

	// Create account service for next step
	accountService := api.NewAccountService(s.apiClient, s.logger)

	// User selected existing org - pass role, org, and services forward
	return account.NewSelectStep(s.ctx, s.theme, s.role, selectedOrg, accountService, s.preferences, s.apiClient, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *SelectStep) Help() help.KeyMap {
	// If loading, no keybindings
	if s.remoteList.IsBusy() {
		return keymap.Simple{Keys: []key.Binding{}}
	}

	// If error, show retry keybinding
	if s.remoteList.HasError() {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	// Normal state: show list navigation and actions
	listKeys := s.remoteList.KeyMap()
	return keymap.Simple{
		Keys: []key.Binding{
			listKeys.CursorUp,
			listKeys.CursorDown,
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "create new"),
			),
		},
	}
}

// Close releases any resources held by the step.
func (s *SelectStep) Close() error {
	return nil
}
