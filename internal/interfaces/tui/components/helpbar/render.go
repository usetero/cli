package helpbar

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const helpPadX = 2

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width

	innerWidth := width - helpPadX*2
	if innerWidth < 0 {
		innerWidth = 0
	}
	m.help.SetWidth(innerWidth)
}

// View satisfies tea.Model.
func (m *Model) View() tea.View {
	content := m.help.ShortHelpView(m.keys)
	if content == "" {
		return tea.NewView("")
	}
	return tea.NewView(
		lipgloss.NewStyle().
			Width(m.width).
			Padding(0, helpPadX).
			Background(m.theme.Background).
			Render(content),
	)
}
