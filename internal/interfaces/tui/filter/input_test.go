package filter

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewInputFilter_PassesThroughNonMouse(t *testing.T) {
	t.Parallel()

	f := NewInputFilter()
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	if got := f(nil, msg); got == nil {
		t.Fatal("expected key message to pass through")
	}
}

func TestNewInputFilter_ThrottlesMouseMotionAndWheel(t *testing.T) {
	t.Parallel()

	f := NewInputFilter()

	if got := f(nil, tea.MouseMotionMsg{}); got == nil {
		t.Fatal("expected first motion event to pass through")
	}
	if got := f(nil, tea.MouseMotionMsg{}); got != nil {
		t.Fatal("expected second motion event to be throttled")
	}

	// Wheel uses the same throttle bucket.
	if got := f(nil, tea.MouseWheelMsg{}); got != nil {
		t.Fatal("expected wheel event to be throttled in same interval")
	}
}
