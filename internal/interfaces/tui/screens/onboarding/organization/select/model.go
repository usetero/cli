package organizationselect

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the organization-select page state.
type Model struct {
	options []core.Option
}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(theme.Theme) *Model { return &Model{} }

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case selectlist.SelectedMsg:
		return m, func() tea.Msg { return SelectedMsg{OrganizationID: tenancy.OrganizationID(typed.Option.ID)} }
	default:
		return m, nil
	}
}

func (m *Model) View() tea.View { return tea.NewView("") }

func (m *Model) SetSize(width, height int) {}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Kind:    core.InputSelect,
		Label:   "Choose your organization.",
		Options: append([]core.Option(nil), m.options...),
	}
}

func (m *Model) SetOrganizations(orgs []tenancy.Organization) {
	options := make([]core.Option, 0, len(orgs))
	for _, org := range orgs {
		options = append(options, core.Option{
			ID:    string(org.ID),
			Label: org.Name,
		})
	}
	m.options = options
}

func Selection(id string) preferences.OrganizationSelection {
	return preferences.OrganizationSelection{OrganizationID: tenancy.OrganizationID(id)}
}
