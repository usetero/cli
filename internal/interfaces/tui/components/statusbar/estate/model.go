package estate

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model renders the estate scale segment.
type Model struct {
	theme theme.Theme
}

func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

func (m *Model) Init() tea.Cmd                       { return nil }
func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *Model) SetSize(_, _ int)                    {}

func (m *Model) View() tea.View {
	return tea.NewView(m.Segment())
}

func (m *Model) Segment() string {
	return m.theme.Text.Subtle.Render("184 services, 91k facts, 12 policies")
}
