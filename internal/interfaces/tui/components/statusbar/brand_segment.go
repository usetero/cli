package statusbar

import (
	"strings"

	"github.com/usetero/cli/internal/interfaces/tui/chrome"
)

func (m *Model) renderBrand() string {
	return chrome.RenderAppWordmark(m.theme)
}

func (m *Model) renderEnv(includeEnv bool) string {
	if !includeEnv || m.env == "" || m.env == productionEnv {
		return ""
	}
	return m.theme.Shell.HeaderLead.Copy().Foreground(m.theme.Warning).Bold(true).Render(strings.ToUpper(m.env))
}
