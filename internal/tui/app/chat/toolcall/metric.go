package toolcall

import (
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderMetric(theme *styles.Theme, input *block.ShowMetricInput, width int) string {
	if input == nil {
		return ""
	}

	colors := theme.Colors

	// Format: "Label: Value Unit"
	label := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Render(input.Label + ": ")

	value := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Render(input.Value)

	unit := ""
	if input.Unit != "" {
		unit = lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render(" " + input.Unit)
	}

	return lipgloss.NewStyle().
		PaddingLeft(2).
		Render(label + value + unit)
}
