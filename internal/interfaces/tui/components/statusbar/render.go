package statusbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the full status bar line.
func (m *Model) View() tea.View {
	left := m.leftSection()
	if m.width <= 0 {
		return tea.NewView(left)
	}

	if lipgloss.Width(left) >= m.width {
		return tea.NewView(ansi.Truncate(left, m.width, "…"))
	}

	fillWidth := m.width - lipgloss.Width(left) - 1
	if fillWidth <= 0 {
		return tea.NewView(left)
	}
	return tea.NewView(lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		" ",
		m.renderTrailSlashes(fillWidth),
	))
}

func (m *Model) renderLeadSlashes(count int) string {
	if count <= 0 {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(m.theme.Palette.Brand).
		Background(m.theme.Background).
		Render(strings.Repeat("╱", count))
}

func (m *Model) renderSeparator() string {
	return lipgloss.NewStyle().
		Foreground(m.theme.Palette.Brand).
		Background(m.theme.Background).
		Render(" // ")
}

func (m *Model) renderTrailSlashes(count int) string {
	if count <= 0 {
		return ""
	}
	return theme.Gradient{
		Start: m.theme.Palette.Brand,
		End:   m.theme.Palette.Accent,
	}.Render(strings.Repeat("╱", count), false)
}

func truncateLabel(label string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(label)
	if len(runes) <= width {
		return label
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}
