package palette

import (
	tea "charm.land/bubbletea/v2"
	baselist "github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns searchable command-palette state.
type Model struct {
	list *baselist.Model
}

func New(appTheme theme.Theme) *Model {
	return &Model{
		list: baselist.New(appTheme, true, "Type a command..."),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.list.Update(msg)
	if typed, ok := msg.(baselist.SelectedMsg); ok {
		selected, ok := typed.Item.(item)
		if !ok {
			return m, cmd
		}
		return m, func() tea.Msg {
			return SubmittedMsg{Command: selected.command}
		}
	}
	if updated, ok := next.(*baselist.Model); ok {
		m.list = updated
	}

	return m, cmd
}

func (m *Model) PreferredHeight(width int) int {
	provider, ok := any(m.list).(core.HeightProvider)
	if !ok {
		return 1
	}
	return provider.PreferredHeight(width)
}
