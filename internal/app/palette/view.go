package palette

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/usetero/cli/internal/styles"
)

// View renders the palette.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	colors := m.theme
	innerWidth := m.width - 2 // borders
	contentWidth := innerWidth - 2*paddingH

	// Header: "Commands ╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱".
	header := m.renderHeader(contentWidth)

	// Input line (cursor marker inserted by input component).
	inputView := m.input.View()

	// Items.
	visible := min(len(m.matches), maxVisible)
	var items []string
	for i := range visible {
		items = append(items, m.renderItem(i, contentWidth))
	}

	// Separator between input and items.
	sep := lipgloss.NewStyle().
		Foreground(colors.Border).
		Background(colors.Bg).
		Render(strings.Repeat("─", contentWidth))

	// Stack: header + gap + input + separator + items.
	var sections []string
	sections = append(sections, header, "", inputView)
	if len(items) > 0 {
		sections = append(sections, sep)
		sections = append(sections, items...)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(colors.Bg).
		Foreground(colors.Text).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Accent).
		BorderBackground(colors.Bg).
		Padding(0, paddingH).
		Render(content)
}

// renderHeader renders "Commands ╱╱╱╱╱╱╱╱" with gradient slashes.
// When nested, shows the parent command name instead.
func (m *Model) renderHeader(width int) string {
	colors := m.theme
	titleText := "Commands"
	if len(m.stack) > 0 {
		// Show the name of the command we drilled into.
		titleText = m.stack[len(m.stack)-1].title
	}
	title := lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg).Bold(true).Render(titleText)
	titleWidth := lipgloss.Width(title)

	remaining := width - titleWidth - 1 // 1 for space
	if remaining <= 0 {
		return title
	}

	slashes := strings.Repeat(diag, remaining)
	slashes = styles.ApplyForegroundGrad(slashes, colors.Accent, colors.AccentAlt)

	return title + " " + slashes
}

// renderItem renders a single list item.
func (m *Model) renderItem(index, width int) string {
	colors := m.theme
	item := m.matches[index]
	name := item.command.Name
	hasChildren := len(item.command.Children) > 0
	isSelected := index == m.selected

	// Reserve space for "›" suffix on parent commands.
	nameWidth := width
	if hasChildren {
		nameWidth = width - 2 // space + "›"
	}

	// Truncate name to fit.
	name = ansi.Truncate(name, nameWidth, "…")

	if isSelected {
		// Highlight matched characters in selected item.
		styled := m.highlightMatches(name, item.matchIndexes,
			lipgloss.NewStyle().Foreground(colors.Bg).Background(colors.Accent).Bold(true),
			lipgloss.NewStyle().Foreground(colors.Bg).Background(colors.Accent),
		)
		if hasChildren {
			suffix := lipgloss.NewStyle().Foreground(colors.Bg).Background(colors.Accent).Render(" ›")
			styled += suffix
		}
		return lipgloss.NewStyle().
			Width(width).
			Background(colors.Accent).
			Foreground(colors.Bg).
			Render(styled)
	}

	// Highlight matched characters in normal item.
	styled := m.highlightMatches(name, item.matchIndexes,
		lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg).Bold(true),
		lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg),
	)
	if hasChildren {
		suffix := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg).Render(" ›")
		styled += suffix
	}
	return lipgloss.NewStyle().
		Width(width).
		Background(colors.Bg).
		Foreground(colors.Text).
		Render(styled)
}

// highlightMatches applies matchStyle to matched rune positions and baseStyle to the rest.
func (m *Model) highlightMatches(text string, indexes []int, matchStyle, baseStyle lipgloss.Style) string {
	if len(indexes) == 0 {
		return baseStyle.Render(text)
	}

	indexSet := make(map[int]bool, len(indexes))
	for _, idx := range indexes {
		indexSet[idx] = true
	}

	var b strings.Builder
	for i, r := range text {
		if indexSet[i] {
			b.WriteString(matchStyle.Render(string(r)))
		} else {
			b.WriteString(baseStyle.Render(string(r)))
		}
	}
	return b.String()
}
