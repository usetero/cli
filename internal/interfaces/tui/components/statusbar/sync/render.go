package sync

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	return tea.NewView(m.Segment())
}

func (m *Model) Segment() string {
	return m.style().Render("●")
}

func (m *Model) style() lipgloss.Style {
	switch m.state {
	case stateReady:
		return m.theme.Text.Success
	case stateConnecting:
		return lipgloss.NewStyle().
			Foreground(m.theme.Palette.Accent).
			Background(m.theme.Background)
	case stateReconnecting:
		return m.theme.Text.Warning
	case stateSyncing:
		return lipgloss.NewStyle().
			Foreground(m.theme.Palette.Accent).
			Background(m.theme.Background)
	case stateError:
		return m.theme.Text.Error
	default:
		return m.theme.Text.Muted
	}
}
