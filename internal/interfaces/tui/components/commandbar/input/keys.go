package input

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func KeyMatchesClose(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, closeBinding)
}

func CloseBinding() key.Binding {
	return closeBinding
}
