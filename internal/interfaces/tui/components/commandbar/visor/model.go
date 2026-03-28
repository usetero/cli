package visor

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the assistant response surface above the command input.
type Model struct {
	theme    theme.Theme
	width    int
	text     string
	expanded bool
}

var _ core.Model = (*Model)(nil)

// New constructs a visor model.
func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// Show updates the visible visor message.
func (m *Model) Show(text string, expanded bool) {
	m.text = text
	m.expanded = expanded
}

// Clear removes the visor message.
func (m *Model) Clear() {
	m.text = ""
	m.expanded = false
}

// Expand expands the visor message surface.
func (m *Model) Expand() { m.expanded = true }

// Collapse collapses the visor message surface.
func (m *Model) Collapse() { m.expanded = false }

// ApplySpec updates the visor state from a visor spec.
func (m *Model) ApplyInput(input *core.Input) {
	if input == nil || input.Label == "" {
		m.Clear()
		return
	}
	m.Show(input.Label, false)
}
