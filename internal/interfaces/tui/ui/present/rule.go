package present

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Rule renders a labeled horizontal divider on the active background.
func Rule(appTheme theme.Theme, width int, label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return lipgloss.NewStyle().
			Width(width).
			Foreground(appTheme.Palette.Border).
			Background(appTheme.Background).
			Render(strings.Repeat("─", width))
	}

	prefix := lipgloss.NewStyle().
		Foreground(appTheme.Palette.TextSubtle).
		Background(appTheme.Background).
		Render(label + " ")
	lineWidth := width - lipgloss.Width(prefix)
	if lineWidth < 0 {
		lineWidth = 0
	}

	line := lipgloss.NewStyle().
		Foreground(appTheme.Palette.Border).
		Background(appTheme.Background).
		Render(strings.Repeat("─", lineWidth))

	return lipgloss.NewStyle().
		Width(width).
		Background(appTheme.Background).
		Render(prefix + line)
}
