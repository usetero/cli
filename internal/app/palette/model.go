package palette

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// Model is the command palette.
type Model struct {
	theme    styles.Theme
	input    *input.Model
	commands []Command // current command list (top-level or children)
	stack    []level   // previous levels for back navigation
	matches  []match   // filtered results
	selected int       // index into matches
	width    int       // available width (set by caller)
}

// New creates a new command palette.
func New(theme styles.Theme, commands []Command) *Model {
	ti := input.New(theme)
	ti.SetPlaceholder("Type a command...")
	ti.SetCharLimit(128)

	m := &Model{
		theme:    theme,
		input:    ti,
		commands: commands,
	}
	m.filter()
	return m
}

// Init returns the initial command (focus + blink).
func (m *Model) Init() tea.Cmd {
	return m.input.Focus()
}

// SetWidth sets the palette width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.SetWidth(width - 2*paddingH - 2) // 2 for "> " prompt
}

// Height returns the rendered height of the palette.
func (m *Model) Height() int {
	visible := min(len(m.matches), maxVisible)
	// border(2) + header + gap + input + separator + items
	return 2 + headerHeight + 1 + inputHeight + 1 + visible
}
