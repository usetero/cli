package selectlist

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the current option list.
func (m *Model) View() tea.View {
	if m.width <= 0 {
		return tea.NewView("")
	}

	innerWidth := present.PanelInnerWidth(m.width)
	if len(m.options) == 0 {
		return tea.NewView(present.Panel(
			m.theme,
			m.width,
			lipgloss.NewStyle().
				Width(innerWidth).
				Foreground(m.theme.Palette.TextMuted).
				Background(m.theme.Background).
				Render("No options available."),
		))
	}

	lines := make([]string, 0, len(m.options))
	for i, option := range m.options {
		cursor := m.theme.Text.Subtle.Render("  ")
		label := m.theme.Text.Body.Render(option.Label)
		if i == m.index {
			cursor = m.theme.Input.Active.Render("> ")
			label = lipgloss.NewStyle().
				Inherit(m.theme.Text.Body).
				Foreground(m.theme.Palette.Accent).
				Bold(true).
				Render(option.Label)
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left, cursor, label)
		if subtitle := strings.TrimSpace(option.Subtitle); subtitle != "" {
			line = lipgloss.JoinHorizontal(
				lipgloss.Left,
				line,
				" ",
				m.theme.Text.Subtle.Render(subtitle),
			)
		}

		lines = append(lines, lipgloss.NewStyle().
			Width(innerWidth).
			Background(m.theme.Background).
			Render(line))
	}

	return tea.NewView(present.Panel(m.theme, m.width, strings.Join(lines, "\n")))
}
