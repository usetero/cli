package commandbar

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar/palette"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

var commandsBinding = key.NewBinding(
	key.WithKeys("/"),
	key.WithHelp("/", "commands"),
)

func (m *Model) SetCommands(commands []core.Command) {
	m.commands = append([]core.Command(nil), commands...)
}

func (m *Model) handleCommandKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if key.Matches(msg, commandsBinding) {
		if m.CanOpenPalette() && len(m.commands) > 0 {
			m.OpenPalette(m.commands)
			return nil, true
		}
		return nil, false
	}

	if m.paletteOpen && palette.KeyMatchesClose(msg) {
		m.closePalette()
		return nil, true
	}

	return nil, false
}

func (m *Model) commandShortHelp() []key.Binding {
	if m.paletteOpen {
		return []key.Binding{palette.CloseBinding()}
	}
	if len(m.commands) > 0 && !m.CapturingInput() {
		return []key.Binding{commandsBinding}
	}
	return nil
}

func (m *Model) CanOpenPalette() bool {
	if m.err != nil || m.busy != nil {
		return false
	}
	if m.paletteOpen {
		return true
	}
	if len(m.commands) == 0 {
		return false
	}
	if m.CapturingInput() {
		return false
	}
	return true
}

func (m *Model) OpenPalette(commands []core.Command) {
	if !m.CanOpenPalette() {
		return
	}
	m.paletteOpen = true
	m.localNotice = nil
	m.mode = ModeSelect
	m.children.action.palette.SetCommands(commands)
	m.children.action.SetActive(m.children.action.palette)
	m.children.visor.Clear()
}

func (m *Model) closePalette() {
	m.paletteOpen = false
	m.applyState(m.input, m.busy, m.err, m.notice)
}
