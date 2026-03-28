package commandbar

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/input"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/visor"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type children struct {
	visor      *visor.Model
	input      *input.Model
	selectlist *selectlist.Model
}

// Model owns the global bottom interaction surface.
type Model struct {
	theme        theme.Theme
	surfaceTheme theme.Theme
	mode         Mode
	width        int
	height       int
	input        *core.Input
	busy         *core.Busy
	children     children
}

var _ core.Model = (*Model)(nil)

// New constructs a command bar model.
func New(appTheme theme.Theme) *Model {
	surfaceTheme := appTheme.OnSurface()

	return &Model{
		theme:        appTheme,
		surfaceTheme: surfaceTheme,
		mode:         ModeInput,
		children: children{
			visor:      visor.New(appTheme),
			input:      input.New(surfaceTheme),
			selectlist: selectlist.New(surfaceTheme),
		},
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.children.visor.Init(),
		m.children.input.Init(),
		m.children.selectlist.Init(),
	)
}

// Update handles command-bar mode changes and local child interaction.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 2)
	if _, cmd := m.children.visor.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if m.busy == nil {
		if child := m.active(); child != nil {
			if _, cmd := child.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

// SetState applies the current generic shell interaction state directly.
func (m *Model) SetState(input *core.Input, busy *core.Busy) {
	m.applyState(input, busy)
}

func (m *Model) active() core.Model {
	switch m.mode {
	case ModeInput:
		return m.children.input
	case ModeSelect:
		return m.children.selectlist
	default:
		return nil
	}
}

func (m *Model) applyState(input *core.Input, busy *core.Busy) {
	m.input = input
	m.busy = busy
	m.children.visor.ApplyInput(input)
	if input == nil {
		m.mode = ModeHidden
		return
	}
	switch input.Kind {
	case core.InputAction:
		m.mode = ModeAction
	case core.InputText:
		m.mode = ModeInput
		m.children.input.ApplyInput(*input)
	case core.InputSelect:
		m.mode = ModeSelect
		m.children.selectlist.ApplyInput(*input)
	default:
		m.mode = ModeHidden
	}
}
