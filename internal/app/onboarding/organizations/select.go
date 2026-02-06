// Package organizations provides organization selection and creation steps.
package organizations

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/loader"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// tokenRefreshedMsg is sent when token refresh completes.
type tokenRefreshedMsg struct {
	err error
}

// SelectModel handles organization selection.
type SelectModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	auth     auth.Auth
	scope    log.Scope

	list            *remotelist.Model
	refreshLoader   *loader.Model
	orgs            []domain.Organization
	selectedOrg     *domain.Organization
	refreshingToken bool
	width           int
	height          int
}

// NewSelect creates a new organization select step.
func NewSelect(
	ctx context.Context,
	theme *styles.Theme,
	services api.APIServices,
	prefs preferences.Preferences,
	authService auth.Auth,
	scope log.Scope,
) *SelectModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}

	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		auth:     authService,
		scope:    scope,
		list:     remotelist.New(theme, "Loading organizations"),
	}
}

// Init starts loading organizations.
func (m *SelectModel) Init() tea.Cmd {
	m.scope.Info("loading organizations")
	return m.list.InitWithLoader(m.loadOrgs())
}

func (m *SelectModel) loadOrgs() tea.Cmd {
	return func() tea.Msg {
		orgs, err := m.services.Organizations.List(m.ctx)
		if err != nil {
			return remotelist.LoadResult{Err: err}
		}

		items := make([]remotelist.Item, len(orgs))
		for i, org := range orgs {
			items[i] = org
		}
		return remotelist.LoadResult{Items: items}
	}
}

func (m *SelectModel) refreshToken(org domain.Organization) tea.Cmd {
	return func() tea.Msg {
		_, err := m.auth.RefreshTokenWithOrganization(m.ctx, org.WorkosOrganizationID)
		return tokenRefreshedMsg{err: err}
	}
}

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		if msg.Err != nil {
			m.scope.Error("failed to load organizations", "error", msg.Err)
			return tea.Batch(m.list.Update(msg), appmsg.ErrorCmd("Failed to load organizations", msg.Err, false))
		}

		m.orgs = make([]domain.Organization, len(msg.Items))
		for i, item := range msg.Items {
			if org, ok := item.(domain.Organization); ok {
				m.orgs[i] = org
			}
		}
		m.scope.Info("organizations loaded", "count", len(m.orgs))

		if len(m.orgs) == 0 {
			m.scope.Debug("no organizations found")
			return func() tea.Msg { return msgs.NoOrgs{} }
		}

		prefID := m.prefs.GetDefaultOrgID()
		if prefID != "" {
			for _, org := range m.orgs {
				if org.ID == prefID {
					m.scope.Debug("using saved organization preference", "id", org.ID)
					return m.selectOrg(org)
				}
			}
		}

		if len(m.orgs) == 1 {
			m.scope.Debug("auto-selected organization (only one)")
			_ = m.prefs.SetDefaultOrgID(m.orgs[0].ID)
			return m.selectOrg(m.orgs[0])
		}

		return m.list.Update(msg)

	case tokenRefreshedMsg:
		m.refreshingToken = false
		if msg.err != nil {
			m.scope.Warn("token refresh failed", "error", msg.err)
		}
		org := *m.selectedOrg
		m.scope.Info("organization selected", "id", org.ID, "name", org.Name)
		return func() tea.Msg { return msgs.OrgSelected{Org: org} }

	case tea.KeyPressMsg:
		if m.refreshingToken || m.list.IsLoading() {
			return nil
		}
		switch msg.String() {
		case "enter":
			if item := m.list.SelectedItem(); item != nil {
				if org, ok := item.(domain.Organization); ok {
					_ = m.prefs.SetDefaultOrgID(org.ID)
					return m.selectOrg(org)
				}
			}
		case "n":
			return func() tea.Msg { return msgs.NoOrgs{} }
		case "r":
			if m.list.HasError() {
				m.scope.Debug("retrying organization load")
				return m.list.Retry()
			}
		}
	}

	if m.refreshingToken && m.refreshLoader != nil {
		return m.refreshLoader.Update(msg)
	}

	return m.list.Update(msg)
}

func (m *SelectModel) selectOrg(org domain.Organization) tea.Cmd {
	m.selectedOrg = &org

	if org.WorkosOrganizationID != "" {
		m.refreshingToken = true
		m.refreshLoader = loader.New(m.theme, "Selecting "+org.Name)
		return tea.Batch(m.refreshLoader.Init(), m.refreshToken(org))
	}

	m.scope.Info("organization selected", "id", org.ID, "name", org.Name)
	return func() tea.Msg { return msgs.OrgSelected{Org: org} }
}

// View renders the organization selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.refreshingToken && m.refreshLoader != nil {
		return m.refreshLoader.View()
	}

	if m.list.HasError() {
		return s.Error.Render("Failed to load organizations. Press 'r' to retry.")
	}

	title := s.Title.Render("Select your organization")
	subtitle := s.Help.Render("Press 'n' to create a new organization")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}

// SetSize updates dimensions.
func (m *SelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *SelectModel) ShortHelp() []key.Binding {
	if m.refreshingToken || m.list.IsLoading() {
		return nil
	}
	if m.list.HasError() {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	// Delegate to list, add step-specific bindings
	bindings := m.list.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new org")),
	)
	return bindings
}
