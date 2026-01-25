package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

var lastMouseEvent time.Time

// MouseEventFilter throttles mouse events to prevent trackpads from flooding
// the app with too many events. Use with tea.WithFilter when creating the program.
func MouseEventFilter(m tea.Model, msg tea.Msg) tea.Msg {
	switch msg.(type) {
	case tea.MouseWheelMsg, tea.MouseMotionMsg:
		now := time.Now()
		if now.Sub(lastMouseEvent) < 15*time.Millisecond {
			return nil
		}
		lastMouseEvent = now
	}
	return msg
}
