package selectlist

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

func (m *Model) SetSize(width, height int) {
	innerWidth := present.PanelInnerWidth(width)
	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	m.list.SetSize(innerWidth, innerHeight)
}

func (m *Model) View() tea.View {
	return m.list.View()
}
