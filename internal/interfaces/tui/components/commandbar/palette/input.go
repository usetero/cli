package palette

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var closeBinding = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "back"),
)

func (m *Model) ShortHelp() []key.Binding {
	return append(m.list.ShortHelp(), closeBinding)
}

func (m *Model) ConsumesKey(msg tea.KeyPressMsg) bool { return m.list.ConsumesKey(msg) }

func KeyMatchesClose(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, closeBinding)
}

func CloseBinding() key.Binding {
	return closeBinding
}
