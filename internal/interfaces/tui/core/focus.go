package core

import tea "charm.land/bubbletea/v2"

// FocusHandler allows a parent to notify a child when it gains or loses focus.
// Components that don't need explicit focus transitions do not need to
// implement it.
type FocusHandler interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
}
