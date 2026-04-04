package present

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

const (
	panelPadLeft  = 1
	panelPadRight = 1
	panelPadY     = 1
)

// PanelInnerWidth returns the content width available inside the shared panel surface.
func PanelInnerWidth(width int) int {
	innerWidth := width - panelPadLeft - panelPadRight
	if innerWidth < 1 {
		return 1
	}
	return innerWidth
}

// Panel renders a surfaced block at the provided outer width.
// It owns only the block's internal padding. Callers remain responsible for any
// outer borders, gutters, or layout margins around the block.
func Panel(appTheme theme.Theme, width int, content string) string {
	content = restoreBackground(strings.TrimRight(content, "\n"), appTheme)

	return lipgloss.NewStyle().
		Width(PanelInnerWidth(width)).
		Padding(panelPadY, panelPadRight, panelPadY, panelPadLeft).
		Background(appTheme.Background).
		Foreground(appTheme.Palette.Text).
		Render(content)
}

func restoreBackground(value string, appTheme theme.Theme) string {
	r, g, b, _ := appTheme.Background.RGBA()
	bgSeq := fmt.Sprintf("\033[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
	value = strings.ReplaceAll(value, "\033[0m", "\033[0m"+bgSeq)
	value = strings.ReplaceAll(value, "\033[m", "\033[m"+bgSeq)
	return value
}
