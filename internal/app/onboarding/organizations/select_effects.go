package organizations

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/loader"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

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

func (m *SelectModel) selectOrg(org domain.Organization) tea.Cmd {
	m.selectedOrg = &org

	if org.WorkosOrganizationID != "" {
		m.refreshingToken = true
		m.refreshLoader = loader.New(m.theme, "Selecting "+org.Name)
		return tea.Batch(m.refreshLoader.Init(), m.refreshToken(org))
	}

	m.scope.Info("organization selected", "id", org.ID, "name", org.Name)
	return func() tea.Msg { return bootstrap.OrgSelected{Org: org} }
}
