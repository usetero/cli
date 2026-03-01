package organizations

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		return m.handleLoadResult(msg)

	case orgTokenRefreshedMsg:
		m.refreshingToken = false
		if msg.err != nil {
			m.scope.Warn("token refresh failed", "error", msg.err)
		}
		org := *m.selectedOrg
		m.scope.Info("organization selected", "id", org.ID, "name", org.Name)
		return func() tea.Msg { return bootstrap.OrgSelected{Org: org} }

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	if m.refreshingToken && m.refreshLoader != nil {
		return m.refreshLoader.Update(msg)
	}

	return m.list.Update(msg)
}

func (m *SelectModel) handleLoadResult(msg remotelist.LoadResult) tea.Cmd {
	if msg.Err != nil {
		m.scope.Error("failed to load organizations", "error", msg.Err)
		return tea.Batch(m.list.Update(msg), appevents.ErrorCmd("Failed to load organizations", msg.Err, false))
	}

	m.orgs = stepkit.CastItems[domain.Organization](msg.Items)
	m.scope.Info("organizations loaded", "count", len(m.orgs))

	if len(m.orgs) == 0 {
		m.scope.Debug("no organizations found")
		return func() tea.Msg { return bootstrap.NoOrgs{} }
	}

	if prefOrg := findOrgByID(m.orgs, m.prefs.GetActiveOrgID()); prefOrg != nil {
		m.scope.Debug("using saved organization preference", "id", prefOrg.ID)
		return m.selectOrg(*prefOrg)
	}

	if len(m.orgs) == 1 {
		m.scope.Debug("auto-selected organization (only one)")
		_ = m.prefs.SetActiveOrgID(m.orgs[0].ID)
		return m.selectOrg(m.orgs[0])
	}

	return m.list.Update(msg)
}

func (m *SelectModel) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.refreshingToken || m.list.IsLoading() {
		return nil
	}

	switch msg.String() {
	case "enter":
		if item := m.list.SelectedItem(); item != nil {
			if org, ok := item.(domain.Organization); ok {
				_ = m.prefs.SetActiveOrgID(org.ID)
				return m.selectOrg(org)
			}
		}
	case "n":
		return func() tea.Msg { return bootstrap.NoOrgs{} }
	case "r":
		if m.list.HasError() {
			m.scope.Debug("retrying organization load")
			return m.list.Retry()
		}
	}
	return nil
}
