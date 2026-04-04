package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/cursor"
)

const (
	bodyPadX = 2
	bodyPadY = 1
)

func (m *Model) View() tea.View {
	contentWidth := m.innerWidth()
	header := m.theme.Shell.HeaderBar.Width(contentWidth).Render(normalizeLine(m.surface.shell.statusbar.View().Content))
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
		cur.Color = m.theme.Palette.Brand
	}
	view.Content = clean
	view.Cursor = cur
	view.AltScreen = true
	view.WindowTitle = m.windowTitle()
	view.MouseMode = tea.MouseModeNone
	if m.surface.body.Busy() != nil {
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

	contentViewport := lipgloss.NewStyle().
		Width(m.bodyContentWidth(width)).
		AlignHorizontal(lipgloss.Left).
		Render(normalizeBody(m.surface.body.View().Content))

	contentContainer := lipgloss.NewStyle().
		Width(innerWidth).
		Padding(bodyPadY, bodyPadX).
		Render(contentViewport)

	bodyStack := lipgloss.JoinVertical(
		lipgloss.Left,
		normalizeLine(m.surface.shell.divider.View().Content),
		contentContainer,
	)

	bodyBlock := m.theme.Shell.Body.Width(width).Height(height).Render(
		bottomAnchor(bodyStack, innerWidth, innerHeight),
	)

	return lipgloss.NewStyle().Width(width).Height(height).Render(bodyBlock)
}

func (m *Model) renderFooter(width int) string {
	parts := []string{normalizeBody(m.surface.shell.commandbar.View().Content)}
	if help := normalizeBody(m.surface.shell.helpbar.View().Content); strings.TrimSpace(help) != "" {
		parts = append(parts, "", help)
	}
	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func bottomAnchor(content string, width, height int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if height > 0 && len(lines) > height {
		lines = lines[len(lines)-height:]
	}
	content = lipgloss.JoinVertical(lipgloss.Left, lines...)

	if height > 0 {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			AlignHorizontal(lipgloss.Left).
			AlignVertical(lipgloss.Bottom).
			Render(content)
	}

	return lipgloss.NewStyle().
		Width(width).
		Render(content)
}
