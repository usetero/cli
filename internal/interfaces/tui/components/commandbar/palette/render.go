package palette

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

var _ core.Model = (*Model)(nil)

type item struct {
	command core.Command
}

func (i item) FilterValue() string { return i.command.Title + " " + i.command.Description }
func (i item) Title() string       { return i.command.Title }
func (i item) Subtitle() string    { return i.command.Description }

func (m *Model) SetCommands(commands []core.Command) {
	items := make([]list.Item, 0, len(commands))
	for _, command := range commands {
		items = append(items, item{command: command})
	}
	m.list.SetItems(items)
}

func (m *Model) View() tea.View {
	return m.list.View()
}

func (m *Model) SetSize(width, height int) {
	innerWidth := present.PanelInnerWidth(width)
	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	m.list.SetSize(innerWidth, innerHeight)
}
