package commandbar

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/palette"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/visor"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type children struct {
	visor  *visor.Model
	action *actionSurface
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
	err          *core.Error
	notice       *core.Notice
	localNotice  *core.Notice
	paletteOpen  bool
	commands     []core.Command
	visorEnabled bool
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
		visorEnabled: true,
		children: children{
			visor:  visor.New(surfaceTheme),
			action: newActionSurface(surfaceTheme),
		},
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.children.visor.Init(),
		m.children.action.Init(),
	)
}

// Update handles command-bar mode changes and local child interaction.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.paletteOpen {
		switch typed := msg.(type) {
		case palette.SubmittedMsg:
			m.closePalette()
			return m, func() tea.Msg {
				return core.CommandSelectedMsg{ID: typed.Command.ID}
			}
		}
	}

	cmds := make([]tea.Cmd, 0, 3)
	if _, cmd := m.children.visor.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := m.children.action.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// SetState applies the current generic shell interaction state directly.
func (m *Model) SetState(input *core.Input, busy *core.Busy, err *core.Error, notice *core.Notice) {
	m.applyState(input, busy, err, notice)
}

func (m *Model) active() core.Model {
	return m.children.action.Active()
}

func (m *Model) FooterTitle(pageTitle string) string {
	if m.paletteOpen {
		return "esc • Commands"
	}
	title := pageTitle
	if title == "" {
		title = "Tero"
	}
	return title + " [/]"
}

func (m *Model) IsPaletteOpen() bool {
	return m.paletteOpen
}

func (m *Model) SetVisorEnabled(enabled bool) {
	m.visorEnabled = enabled
}

// HandleKey gives the active command-bar control first right of refusal on a
// key press. It returns whether the key was consumed.
func (m *Model) HandleKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if cmd, consumed := m.handleCommandKey(msg); consumed {
		return cmd, true
	}

	if m.err != nil || m.busy != nil {
		return nil, false
	}

	switch m.mode {
	case ModeInput:
		return m.children.action.HandleKey(msg)
	case ModeSelect:
		return m.children.action.HandleKey(msg)
	default:
		return nil, false
	}
}

func (m *Model) applyState(input *core.Input, busy *core.Busy, err *core.Error, notice *core.Notice) {
	m.input = input
	m.busy = busy
	m.err = err
	m.notice = notice
	if m.paletteOpen {
		return
	}
	m.children.visor.ApplyState(input)
	_ = m.children.action.busy.ApplyBusy(busy)
	m.children.action.err.ApplyError(err)
	if err != nil {
		m.mode = ModeError
		m.children.action.SetActive(m.children.action.err)
		return
	}
	if busy != nil {
		m.children.action.SetActive(m.children.action.busy)
		return
	}
	if input == nil {
		m.mode = ModeHidden
		m.children.action.SetActive(nil)
		return
	}
	switch input.Kind {
	case core.InputConfirm:
		m.mode = ModeAction
		m.children.action.SetActive(nil)
	case core.InputText:
		m.mode = ModeInput
		m.children.action.input.ApplyInput(*input)
		m.children.action.SetActive(m.children.action.input)
	case core.InputMultiline:
		m.mode = ModeInput
		m.children.action.input.ApplyInput(*input)
		m.children.action.SetActive(m.children.action.input)
	case core.InputSelect:
		m.mode = ModeSelect
		m.children.action.selectlist.ApplyInput(*input)
		m.children.action.SetActive(m.children.action.selectlist)
	default:
		m.mode = ModeHidden
		m.children.action.SetActive(nil)
	}
}
