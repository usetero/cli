package visor

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const collapsedLines = 3
const visorPadX = 2

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the visor output surface.
func (m *Model) View() tea.View {
	if strings.TrimSpace(m.text) == "" {
		return tea.NewView("")
	}

	content := strings.TrimSpace(m.text)
	if !m.expanded {
		lines := strings.Split(content, "\n")
		if len(lines) > collapsedLines {
			content = strings.Join(lines[:collapsedLines], "\n") + "\n" + m.theme.Text.Subtle.Render("...")
		}
	}

	block := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Text.Body.Render(content),
	)

	innerWidth := m.width - visorPadX*2
	if innerWidth < 1 {
		innerWidth = 1
	}

	return tea.NewView(
		lipgloss.NewStyle().
			Width(innerWidth).
			Padding(0, visorPadX).
			Background(m.theme.Background).
			Render(block),
	)
}
