package chrome

import (
	"strings"

	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// RenderAppWordmark renders the shared Tero wordmark using the brand gradient.
func RenderAppWordmark(t theme.Theme) string {
	return t.Gradients.Brand.Render(strings.ToUpper(theme.AppName), true)
}

// RenderSlashMotif renders a repeated slash motif using the shared shell gradient.
func RenderSlashMotif(t theme.Theme, count int) string {
	if count <= 0 {
		return ""
	}
	return t.Gradients.Motif.Render(strings.Repeat("╱", count), false)
}
