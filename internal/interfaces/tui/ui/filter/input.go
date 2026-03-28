package filter

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const mouseThrottleInterval = 15 * time.Millisecond

// NewInputFilter returns a Bubble Tea filter that throttles high-frequency
// mouse motion/wheel events to keep the event loop responsive on trackpads.
func NewInputFilter() func(tea.Model, tea.Msg) tea.Msg {
	lastMouseEventAt := time.Time{}

	return func(_ tea.Model, msg tea.Msg) tea.Msg {
		switch msg.(type) {
		case tea.MouseWheelMsg, tea.MouseMotionMsg:
			now := time.Now()
			if now.Sub(lastMouseEventAt) < mouseThrottleInterval {
				return nil
			}
			lastMouseEventAt = now
		}
		return msg
	}
}
