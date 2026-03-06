package helpbar

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// Model renders the app short-help row.
type Model struct {
	help help.Model
}

// New creates a help bar model.
func New() *Model {
	return &Model{help: help.New()}
}

// SetWidth configures help layout width.
func (m *Model) SetWidth(width int) {
	m.help.SetWidth(width)
}

// Short renders short help for the provided bindings.
func (m *Model) Short(bindings []key.Binding) string {
	return m.help.ShortHelpView(bindings)
}
