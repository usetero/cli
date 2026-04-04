package visor

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const collapsedLines = 3
const visorPadX = 1

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the visor output surface.
func (m *Model) View() tea.View {
	title := strings.TrimSpace(m.title)
	detail := strings.TrimSpace(m.detail)
	if title == "" && detail == "" {
		return tea.NewView("")
	}

	titleBlock := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Text).
		Background(m.theme.Background).
		Bold(true).
		Render(title)

	block := titleBlock
	if detail != "" {
		renderedDetail := strings.TrimSpace(detail)
		if !m.expanded {
			lines := strings.Split(renderedDetail, "\n")
			if len(lines) > collapsedLines-2 {
				limit := collapsedLines - 2
				if limit < 1 {
					limit = 1
				}
				renderedDetail = strings.Join(lines[:limit], "\n") + "\n" + m.theme.Text.Subtle.Render("...")
			}
		}
		block = lipgloss.JoinVertical(
			lipgloss.Left,
			titleBlock,
			"",
			m.theme.Text.Body.Render(renderedDetail),
		)
	}

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
