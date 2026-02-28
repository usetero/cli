package organizations

import (
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
)

func (m *CreateModel) createOrg(name string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		result, err := m.services.Organizations.Create(m.ctx, api.CreateOrganizationInput{
			ID:   id,
			Name: name,
		})
		if err != nil {
			return orgCreatedMsg{err: err}
		}
		return orgCreatedMsg{result: result}
	}
}
