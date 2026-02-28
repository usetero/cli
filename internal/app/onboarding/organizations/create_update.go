package organizations

import (
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *CreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orgCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.scope.Error("failed to create organization", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to create organization", msg.err, false)
		}
		_ = m.prefs.SetActiveOrgID(msg.result.Organization.ID)
		org := *msg.result.Organization
		m.scope.Info("organization created", "id", org.ID, "name", org.Name)
		return func() tea.Msg { return bootstrap.OrgCreated{Org: org} }

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		if msg.String() == "enter" {
			name := m.input.Value()
			if name == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.scope.Info("creating organization", "name", name)
			return m.createOrg(name)
		}
	}

	return m.input.Update(msg)
}
