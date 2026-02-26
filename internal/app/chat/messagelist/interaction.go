package messagelist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type keyDecision struct {
	handle      bool
	focusDelta  int
	scrollDelta int
}

func reduceKeyPress(msg tea.KeyPressMsg, focused bool) keyDecision {
	if !focused {
		return keyDecision{}
	}
	switch {
	case key.Matches(msg, focusPrevKey):
		return keyDecision{handle: true, focusDelta: -1}
	case key.Matches(msg, focusNextKey):
		return keyDecision{handle: true, focusDelta: 1}
	case key.Matches(msg, scrollUpKey):
		return keyDecision{handle: true, scrollDelta: -1}
	case key.Matches(msg, scrollDownKey):
		return keyDecision{handle: true, scrollDelta: 1}
	default:
		return keyDecision{}
	}
}

func reduceMouseWheel(button tea.MouseButton) int {
	switch button {
	case tea.MouseWheelUp:
		return -5
	case tea.MouseWheelDown:
		return 5
	default:
		return 0
	}
}
