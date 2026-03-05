package organizations

import (
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
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
			return appevents.PublishErrorToastCmd("Failed to create organization", msg.err, false)
		}
		_ = m.prefs.SetActiveOrgID(msg.result.Organization.ID)
		org := *msg.result.Organization
		m.scope.Info("organization created", "id", org.ID, "name", org.Name)
		return func() tea.Msg { return bootstrap.OrgCreated{Org: org} }

	case tea.KeyPressMsg:
		if name, ok := stepkit.ParseCreateSubmit(msg, m.creating, m.input.Value()); ok {
			m.creating = true
			m.err = nil
			m.scope.Info("creating organization", "name", name)
			return m.createOrg(name)
		}
	}

	return m.input.Update(msg)
}
