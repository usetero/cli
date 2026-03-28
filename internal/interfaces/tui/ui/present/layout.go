package present

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Stack vertically joins rendered content.
func Stack(children ...string) string { return StackGap(0, children...) }

// StackGap vertically joins rendered content with blank lines between items.
func StackGap(gap int, children ...string) string {
	filtered := make([]string, 0, len(children))
	for _, child := range children {
		if strings.TrimSpace(child) != "" {
			filtered = append(filtered, child)
		}
	}
	if len(filtered) == 0 {
		return ""
	}

	lines := make([]string, 0, len(filtered)*2)
	for i, child := range filtered {
		if i > 0 && gap > 0 {
			for range gap {
				lines = append(lines, "")
			}
		}
		lines = append(lines, child)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Row horizontally joins rendered content.
func Row(children ...string) string {
	filtered := make([]string, 0, len(children))
	for _, child := range children {
		if child != "" {
			filtered = append(filtered, child)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, filtered...)
}
