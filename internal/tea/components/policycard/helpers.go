package policycard

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

func alignLeftRight(left, right string, width int, theme styles.Theme) string {
	if left == "" && right == "" {
		return ""
	}
	if right == "" {
		return left
	}
	if left == "" {
		return right
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	pad := lipgloss.NewStyle().Background(theme.Bg).Render(strings.Repeat(" ", gap))
	return left + pad + right
}

func renderDivider(theme styles.Theme, width int, label string) string {
	s := lipgloss.NewStyle().Foreground(theme.TextSubtle).Background(theme.Bg)
	prefix := "── "
	suffix := " "
	used := lipgloss.Width(prefix) + lipgloss.Width(label) + lipgloss.Width(suffix)
	remaining := width - used
	if remaining < 1 {
		remaining = 1
	}
	line := prefix + label + suffix + strings.Repeat("─", remaining)
	return s.Render(line)
}

// renderExampleSection wraps evidence body with the "Example" divider
// and optional log body. Used by example-based evidence types.
func (m *Model) renderExampleSection(body string) string {
	var parts []string
	if logBody := m.viewLogBody(); logBody != "" {
		text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
		parts = append(parts, text.Render(logBody))
		parts = append(parts, "")
	}
	parts = append(parts, body)
	return renderDivider(m.theme, m.width, "Example") + "\n\n" + strings.Join(parts, "\n")
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
