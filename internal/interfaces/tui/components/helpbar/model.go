package helpbar

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model renders the app short-help row.
type Model struct {
	help  help.Model
	theme theme.Theme
}

// New creates a help bar model.
func New(appTheme theme.Theme) *Model {
	h := help.New()
	h.ShortSeparator = " • "
	h.Styles.ShortKey = appTheme.Text.Muted
	h.Styles.ShortDesc = appTheme.Text.Subtle
	h.Styles.ShortSeparator = appTheme.Text.Subtle
	h.Styles.Ellipsis = appTheme.Text.Subtle
	return &Model{help: h, theme: appTheme}
}

// SetWidth configures help layout width.
func (m *Model) SetWidth(width int) {
	m.help.SetWidth(width)
}

// Short renders short help for the provided bindings.
func (m *Model) Short(bindings []key.Binding) string {
	content := m.help.ShortHelpView(bindings)
	if content == "" {
		return ""
	}
	return content
}
