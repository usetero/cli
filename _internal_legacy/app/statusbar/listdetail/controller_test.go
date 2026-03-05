package listdetail

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestClampCursor(t *testing.T) {
	t.Parallel()

	if got := ClampCursor(10, 3); got != 2 {
		t.Fatalf("ClampCursor high = %d, want 2", got)
	}
	if got := ClampCursor(-1, 3); got != 0 {
		t.Fatalf("ClampCursor low = %d, want 0", got)
	}
	if got := ClampCursor(1, 3); got != 1 {
		t.Fatalf("ClampCursor inside = %d, want 1", got)
	}
	if got := ClampCursor(5, 0); got != 0 {
		t.Fatalf("ClampCursor empty = %d, want 0", got)
	}
}

func TestHandleKeyPressListSelect(t *testing.T) {
	t.Parallel()

	cursor := 0
	selected := -1
	c := Controller{
		HasList:       func() bool { return true },
		GetListCursor: func() int { return cursor },
		SetListCursor: func(v int) { cursor = v },
		ListLen:       func() int { return 3 },
		OnListSelect: func(index int) tea.Cmd {
			selected = index
			return nil
		},
	}

	_ = c.HandleKeyPress(tea.KeyPressMsg{Code: tea.KeyDown})
	if cursor != 1 {
		t.Fatalf("cursor after down = %d, want 1", cursor)
	}

	_ = c.HandleKeyPress(tea.KeyPressMsg{Code: tea.KeyEnter})
	if selected != 1 {
		t.Fatalf("selected = %d, want 1", selected)
	}
}
