package organizationselect

import (
	"context"
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	bubbleslist "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/loader"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	accountselect "github.com/usetero/cli/internal/tui/onboarding/account/select"
	organizationcreate "github.com/usetero/cli/internal/tui/onboarding/organization/create"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const createNewOrgID domain.OrganizationID = "__CREATE_NEW__"

// OrgListItem wraps domain.Organization to implement list.Item.
type OrgListItem struct {
	domain.Organization
}

func (i OrgListItem) FilterValue() string { return i.Name }

// orgDelegate renders each organization in the list.
type orgDelegate struct {
	theme *styles.Theme
}

func (d orgDelegate) Height() int                                    { return 1 }
func (d orgDelegate) Spacing() int                                   { return 0 }
func (d orgDelegate) Update(_ tea.Msg, _ *bubbleslist.Model) tea.Cmd { return nil }
func (d orgDelegate) Render(w io.Writer, m bubbleslist.Model, index int, item bubbleslist.Item) {
	i, ok := item.(OrgListItem)
	if !ok {
		return
	}

	colors := d.theme.Colors
	str := i.Name
	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Accent).Bold(true).Render("> "+str))
	} else {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Page.Text).Render("  "+str))
	}
}

// tokenRefreshMsg is sent when token refresh completes.
type tokenRefreshMsg struct {
	accessToken string
	err         error
}

// Model handles selecting an organization.
type Model struct {
	ctx   context.Context
	theme *styles.Theme
	role  string

	services api.APIServices
	prefs    preferences.Preferences
	auth     auth.Auth
	logger   log.Logger

	remoteList          remotelist.Model
	refreshLoader       loader.Model
	orgs                []domain.Organization
	selectedOrgID       domain.OrganizationID
	selectedWorkosOrgID domain.WorkosOrganizationID
	tokenRefreshed      bool
	refreshingToken     bool
	width               int
	height              int
}

// New creates a new organization select model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	services api.APIServices,
	prefs preferences.Preferences,
	authService auth.Auth,
	logger log.Logger,
) Model {
	delegate := orgDelegate{theme: theme}

	return Model{
		ctx:        ctx,
		theme:      theme,
		role:       role,
		services:   services,
		prefs:      prefs,
		auth:       authService,
		logger:     logger,
		remoteList: remotelist.New(theme, delegate, "Loading organizations", logger),
		width:      80,
	}
}

// Init starts loading organizations.
func (m Model) Init() tea.Cmd {
	_, cmd := m.remoteList.InitWithLoader(m.loadOrganizations())
	return cmd
}

// loadOrganizations returns a command that loads organizations.
func (m Model) loadOrganizations() tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("loading organizations")

		orgs, err := m.services.Organizations.List(m.ctx)
		if err != nil {
			m.logger.Error("failed to load organizations", "error", err)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		m.logger.Info("organizations loaded", "count", len(orgs))

		items := make([]list.Item, len(orgs))
		for i, org := range orgs {
			items[i] = OrgListItem{Organization: org}
		}

		return remotelist.LoadResultMsg{Items: items, Err: nil}
	}
}

// refreshToken returns a command that refreshes the token with org scope.
func (m Model) refreshToken() tea.Cmd {
	return func() tea.Msg {
		accessToken, err := m.auth.RefreshTokenWithOrganization(m.ctx, m.selectedWorkosOrgID)
		return tokenRefreshMsg{accessToken: accessToken, err: err}
	}
}

// selectOrg handles org selection.
func (m Model) selectOrg(orgID domain.OrganizationID, workosOrgID domain.WorkosOrganizationID) (Model, tea.Cmd) {
	m.selectedOrgID = orgID
	m.selectedWorkosOrgID = workosOrgID

	if orgID == createNewOrgID {
		m.tokenRefreshed = true
		return m, nil
	}

	if workosOrgID != "" {
		m.refreshingToken = true

		orgName := "organization"
		for _, org := range m.orgs {
			if org.ID == orgID {
				orgName = org.Name
				break
			}
		}
		m.refreshLoader = loader.New(m.theme, "Selecting "+orgName)

		return m, tea.Batch(m.refreshLoader.Init(), m.refreshToken())
	}

	m.tokenRefreshed = true
	return m, nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	if m.refreshingToken {
		var cmd tea.Cmd
		m.refreshLoader, cmd = m.refreshLoader.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tokenRefreshMsg:
		m.refreshingToken = false
		m.tokenRefreshed = true
		if msg.err != nil {
			m.logger.Error("failed to refresh token with organization", "error", msg.err)
		} else {
			m.logger.Debug("token refreshed with organization scope")
			// Token is now stored in auth service - subsequent API calls will use it automatically
		}
		return m, nil

	case remotelist.LoadResultMsg:
		if msg.Err == nil {
			m.orgs = make([]domain.Organization, 0, len(msg.Items))
			for _, item := range msg.Items {
				if orgItem, ok := item.(OrgListItem); ok {
					m.orgs = append(m.orgs, orgItem.Organization)
				}
			}

			userPref := m.prefs.GetDefaultOrgID()

			// No orgs → auto-select create
			if len(m.orgs) == 0 {
				m.logger.Debug("auto-selected create organization", "reason", "no organizations found")
				m.selectedOrgID = createNewOrgID
				m.tokenRefreshed = true
			}

			// Has preference and exists → auto-select
			if userPref != "" {
				for _, org := range m.orgs {
					if org.ID == userPref {
						m.logger.Debug("auto-selected organization from preference", "id", userPref, "name", org.Name)
						var cmd tea.Cmd
						m, cmd = m.selectOrg(org.ID, org.WorkosOrganizationID)
						cmds = append(cmds, cmd)
						break
					}
				}
			}

			// No preference and only 1 org → auto-select and save
			if userPref == "" && len(m.orgs) == 1 {
				m.logger.Info("auto-selected organization", "id", m.orgs[0].ID, "name", m.orgs[0].Name, "reason", "only one available")
				if err := m.prefs.SetDefaultOrgID(m.orgs[0].ID); err != nil {
					m.logger.Error("failed to save organization preference", "error", err)
				}
				var cmd tea.Cmd
				m, cmd = m.selectOrg(m.orgs[0].ID, m.orgs[0].WorkosOrganizationID)
				cmds = append(cmds, cmd)
			}
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.remoteList.HasError() {
				m.logger.Info("user requested retry")
				var cmd tea.Cmd
				m.remoteList, cmd = m.remoteList.Retry()
				cmds = append(cmds, cmd)
			}
		case "enter":
			if !m.remoteList.IsLoaded() {
				break
			}
			selected := m.remoteList.SelectedItem()
			if orgItem, ok := selected.(OrgListItem); ok {
				m.logger.Info("organization selected", "id", orgItem.ID, "name", orgItem.Name)
				if err := m.prefs.SetDefaultOrgID(orgItem.ID); err != nil {
					m.logger.Error("failed to save organization preference", "error", err)
				}
				var cmd tea.Cmd
				m, cmd = m.selectOrg(orgItem.ID, orgItem.WorkosOrganizationID)
				cmds = append(cmds, cmd)
			}
		case "n":
			if !m.remoteList.IsLoaded() {
				break
			}
			m.logger.Info("user chose to create new organization")
			var cmd tea.Cmd
			m, cmd = m.selectOrg(createNewOrgID, "")
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	m.remoteList, cmd = m.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the organization selection UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.remoteList.IsBusy() {
		return m.remoteList.View()
	}

	if m.refreshingToken {
		return m.refreshLoader.View()
	}

	if m.remoteList.HasError() {
		return m.remoteList.View()
	}

	title := themeStyles.Title.Render("Select your organization")
	subtitle := themeStyles.Help.Render("This will be your default workspace")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.remoteList.View(),
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	m.remoteList = m.remoteList.SetWidth(width)
	return m
}

// IsBusy returns true while loading or refreshing token.
func (m Model) IsBusy() bool {
	return !m.remoteList.IsLoaded() || m.refreshingToken
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.remoteList.HasError()
}

// Error returns the current error.
func (m Model) Error() error {
	return m.remoteList.Error()
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.remoteList.IsBusy() {
		return keymap.Simple{Keys: []key.Binding{}}
	}

	if m.remoteList.HasError() {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	listKeys := m.remoteList.KeyMap()
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

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.selectedOrgID == "" || !m.tokenRefreshed {
		return nil, step.ErrNotReady
	}

	if m.selectedOrgID == createNewOrgID {
		return organizationcreate.New(
			m.ctx,
			m.theme,
			m.role,
			m.services,
			m.prefs,
			m.auth,
			m.logger,
		), nil
	}

	var selectedOrg domain.Organization
	for _, org := range m.orgs {
		if org.ID == m.selectedOrgID {
			selectedOrg = org
			break
		}
	}

	return accountselect.New(
		m.ctx,
		m.theme,
		m.role,
		selectedOrg,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
