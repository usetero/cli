package create

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	cmdinput "github.com/usetero/cli/internal/interfaces/tui/components/commandbar/input"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the organization-create page state.
type Model struct{}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(theme.Theme) *Model { return &Model{} }

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case cmdinput.SubmittedMsg:
		return m, func() tea.Msg { return CreatedMsg{Name: typed.Text} }
	default:
		return m, nil
	}
}

func (m *Model) View() tea.View { return tea.NewView("") }

func (m *Model) SetSize(width, height int) {}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Kind:        core.InputText,
		Label:       "Create your organization.",
		Placeholder: "Acme",
	}
}

func Submission(name string) tenancy.OrganizationCreate {
	return tenancy.OrganizationCreate{Name: name}
}
