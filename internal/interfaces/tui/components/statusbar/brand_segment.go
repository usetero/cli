package statusbar

import (
	"strings"

	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func (m *Model) renderBrand(includeEnv bool) string {
	brand := m.theme.Shell.HeaderLead.Render("╱╱ ") + m.theme.Shell.HeaderBrand.Render(strings.ToUpper(theme.AppName))
	if !includeEnv || m.env == "" || m.env == productionEnv {
		return brand
	}
	return brand + " " + m.theme.Text.Warning.Render(strings.ToUpper(m.env))
}
