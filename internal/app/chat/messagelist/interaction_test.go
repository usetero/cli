package messagelist

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestReduceKeyPress(t *testing.T) {
	t.Parallel()

	t.Run("ignored when not focused", func(t *testing.T) {
		t.Parallel()
		d := reduceKeyPress(tea.KeyPressMsg{Code: tea.KeyUp}, false)
		if d.handle || d.focusDelta != 0 || d.scrollDelta != 0 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("focus prev", func(t *testing.T) {
		t.Parallel()
		d := reduceKeyPress(tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModShift}, true)
		if !d.handle || d.focusDelta != -1 || d.scrollDelta != 0 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("focus next", func(t *testing.T) {
		t.Parallel()
		d := reduceKeyPress(tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift}, true)
		if !d.handle || d.focusDelta != 1 || d.scrollDelta != 0 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("scroll up", func(t *testing.T) {
		t.Parallel()
		d := reduceKeyPress(tea.KeyPressMsg{Code: tea.KeyUp}, true)
		if !d.handle || d.focusDelta != 0 || d.scrollDelta != -1 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("scroll down", func(t *testing.T) {
		t.Parallel()
		d := reduceKeyPress(tea.KeyPressMsg{Code: tea.KeyDown}, true)
		if !d.handle || d.focusDelta != 0 || d.scrollDelta != 1 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})
}

func TestReduceMouseWheel(t *testing.T) {
	t.Parallel()

	if got := reduceMouseWheel(tea.MouseWheelUp); got != -5 {
		t.Fatalf("MouseWheelUp=%d, want -5", got)
	}
	if got := reduceMouseWheel(tea.MouseWheelDown); got != 5 {
		t.Fatalf("MouseWheelDown=%d, want 5", got)
	}
	if got := reduceMouseWheel(tea.MouseLeft); got != 0 {
		t.Fatalf("MouseLeft=%d, want 0", got)
	}
}
