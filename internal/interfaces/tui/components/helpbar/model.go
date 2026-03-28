package helpbar

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model renders the app short-help row.
type Model struct {
	help  help.Model
	theme theme.Theme
	keys  []key.Binding
	width int
}

var _ core.Model = (*Model)(nil)

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

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update satisfies tea.Model. The help bar currently has no local event-loop state.
func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// SetBindings sets the bindings rendered in the help bar.
func (m *Model) SetBindings(bindings []key.Binding) {
	m.keys = append([]key.Binding(nil), bindings...)
}
