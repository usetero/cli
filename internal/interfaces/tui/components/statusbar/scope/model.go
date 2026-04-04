package scope

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model renders the current estate scope.
type Model struct {
	theme        theme.Theme
	organization tenancy.Organization
}

func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case events.OrganizationSelectedMsg:
		m.organization = typed.Organization
	case events.AccountSelectedMsg:
		if typed.Scope.Organization.ID != "" {
			m.organization = typed.Scope.Organization
		}
	case events.AccountRuntimeUpdatedMsg:
		if typed.Status.Scope.Organization.ID != "" {
			m.organization = typed.Status.Scope.Organization
		}
	}
	return m, nil
}

func (m *Model) View() tea.View {
	return tea.NewView(m.Segment())
}

func (m *Model) Segment() string {
	if name := strings.TrimSpace(m.organization.Name); name != "" {
		return m.theme.Shell.HeaderLead.Render(name)
	}
	if id := strings.TrimSpace(string(m.organization.ID)); id != "" {
		return m.theme.Shell.HeaderLead.Render(id)
	}
	return ""
}

func (m *Model) SetSize(_, _ int) {}
