package policytab

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
)

type testDetail struct {
	cursor int
	len    int
}

func (d *testDetail) Len() int        { return d.len }
func (d *testDetail) Cursor() int     { return d.cursor }
func (d *testDetail) SetCursor(v int) { d.cursor = v }
func (d *testDetail) Prompt() tea.Cmd { return func() tea.Msg { return "prompt" } }

func TestBase_ApplyIfChangedAndCursorClamp(t *testing.T) {
	t.Parallel()

	b := New("x")
	b.SetCursor(5)
	applied := b.ApplyIfChanged("a", 2, func() {})
	if !applied {
		t.Fatalf("expected applied=true")
	}
	if b.Cursor() != 1 {
		t.Fatalf("expected cursor clamped to 1, got %d", b.Cursor())
	}

	applied = b.ApplyIfChanged("a", 2, func() { t.Fatal("should not reapply") })
	if applied {
		t.Fatalf("expected applied=false on same key")
	}
}

func TestBase_HasList(t *testing.T) {
	t.Parallel()

	b := New("x")
	if b.HasList(1) {
		t.Fatalf("expected false without data")
	}
	b.SetHasData(true)
	if !b.HasList(1) {
		t.Fatalf("expected true with data and items")
	}
	if b.HasList(0) {
		t.Fatalf("expected false with zero items")
	}
}

func TestBase_NavController(t *testing.T) {
	t.Parallel()

	b := New("x")
	b.SetHasData(true)
	detail := &testDetail{len: 2}

	ctrl := b.NavController(
		func() int { return 3 },
		func(index int) tea.Cmd {
			return func() tea.Msg { return index }
		},
		func() listdetail.Detail { return detail },
		func() { detail = nil },
	)

	// Enter detail, then back should clear it.
	_ = ctrl.HandleKeyPress(tea.KeyPressMsg{})
}

func TestUpdatePoll_ForwardsToTabpollCycle(t *testing.T) {
	t.Parallel()

	b := New("test")
	b.SetHasData(true)

	// Poll without DB should be handled and no command.
	cmd, handled := UpdatePoll[int](&b, tabpoll.PollMsg{Source: "test"}, nil, func(int) {})
	if !handled {
		t.Fatalf("expected handled")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd when db missing")
	}
}
