package messages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/components/thinking"
)

// StartThinking shows the thinking indicator.
func (m Model) StartThinking() (Model, tea.Cmd) {
	m.thinking = thinking.New(m.theme, "Thinking")
	m.showThinking = true
	return m, m.thinking.Init()
}

// StopThinking hides the thinking indicator.
func (m Model) StopThinking() Model {
	m.showThinking = false
	return m
}

// IsThinking returns true if the thinking indicator is shown.
func (m Model) IsThinking() bool {
	return m.showThinking
}

// renderThinking renders the thinking indicator with label.
func (m Model) renderThinking() string {
	colors := m.theme.Colors
	label := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientEnd).
		Bold(true).
		PaddingLeft(1).
		Render("Tero")
	content := lipgloss.NewStyle().PaddingLeft(1).Render(m.thinking.View())
	return lipgloss.JoinVertical(lipgloss.Left, label, content)
}
