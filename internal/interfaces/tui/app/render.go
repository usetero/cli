package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/cursor"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

func (m *Model) View() tea.View {
	contentWidth := m.innerWidth()
	header := m.theme.Shell.HeaderBar.Width(contentWidth).Render(normalizeLine(m.children.statusbar.View().Content))
	body := m.renderBody(contentWidth)
	footer := m.renderFooter(contentWidth)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.width > 0 && m.height > 0 {
		content = m.theme.Shell.Outer.Width(m.width).Height(m.height).Render(content)
	} else {
		content = m.theme.Shell.Outer.Render(content)
	}

	view := tea.NewView(content)
	clean, cur := cursor.Extract(view.Content)
	if cur != nil {
		cur.Color = m.theme.Palette.AccentAlt
	}
	view.Content = clean
	view.Cursor = cur
	view.AltScreen = true
	view.WindowTitle = windowTitle
	view.MouseMode = tea.MouseModeNone
	if m.body.Busy() != nil {
		view.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, 0)
	}
	view.BackgroundColor = m.theme.Background
	return view
}

func (m *Model) renderBody(width int) string {
	height := m.bodyHeight(m.commandbarHeight(width, m.helpbarHeight(width)), m.helpbarHeight(width))
	innerWidth := width - m.theme.Shell.Body.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := height - m.theme.Shell.Body.GetVerticalFrameSize()
	if innerHeight < 1 {
		innerHeight = 1
	}

	bodyContent := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Bottom).
		Render(normalizeBody(m.body.View().Content))

	bodyBlock := m.theme.Shell.Body.Width(width).Height(height).Render(bodyContent)
	return lipgloss.NewStyle().Width(width).Height(height).Render(bodyBlock)
}

func (m *Model) renderFooter(width int) string {
	rule := present.Rule(m.theme, width, "◇ Tero")
	parts := []string{rule, "", normalizeBody(m.children.commandbar.View().Content)}
	if help := normalizeBody(m.children.helpbar.View().Content); strings.TrimSpace(help) != "" {
		parts = append(parts, "", help)
	}
	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}
