package round

import (
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func newTestRound(t *testing.T) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	input := msgs.UserSubmittedInput{Text: "hello"}
	return New(theme, "conv-1", "acct-1", "user-1", input, 80, nil, nil, nil, scope)
}

func hasBlockKind(blocks []block.Block, kind block.Kind) bool {
	for _, b := range blocks {
		if b.Kind() == kind {
			return true
		}
	}
	return false
}

func TestNew(t *testing.T) {
	t.Parallel()

	m := newTestRound(t)

	if m.State() != StateActive {
		t.Errorf("expected StateActive, got %d", m.State())
	}
	if m.ID() != "user-1" {
		t.Errorf("expected ID user-1, got %s", m.ID())
	}
	if len(m.turns) != 1 {
		t.Errorf("expected 1 turn, got %d", len(m.turns))
	}
}

func TestBlocks(t *testing.T) {
	t.Parallel()

	t.Run("includes thinking animation while active", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		blocks := m.Blocks()
		if !hasBlockKind(blocks, block.KindThinkingAnimation) {
			t.Error("expected thinking animation block while active")
		}
	})

	t.Run("excludes thinking animation when complete", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:     "user-1",
			StopReason: "end_turn",
			Message:    domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		if m.State() != StateComplete {
			t.Fatalf("expected StateComplete, got %d", m.State())
		}
		if hasBlockKind(m.Blocks(), block.KindThinkingAnimation) {
			t.Error("thinking animation should be removed after completion")
		}
	})

	t.Run("excludes thinking animation when cancelled", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()

		if m.State() != StateCancelled {
			t.Fatalf("expected StateCancelled, got %d", m.State())
		}
		if hasBlockKind(m.Blocks(), block.KindThinkingAnimation) {
			t.Error("thinking animation should be removed after cancel")
		}
	})
}

func TestCancel(t *testing.T) {
	t.Parallel()

	t.Run("sets state to cancelled", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()

		if m.State() != StateCancelled {
			t.Errorf("expected StateCancelled, got %d", m.State())
		}
	})

	t.Run("sets end time", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()

		if m.endTime.IsZero() {
			t.Error("expected endTime to be set after cancel")
		}
	})

	t.Run("propagates to turns", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()

		for i, turn := range m.turns {
			if turn.State() != 3 { // turn.StateComplete
				t.Errorf("turn %d: expected StateComplete, got %d", i, turn.State())
			}
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("StreamCompleted with end_turn completes round", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:     "user-1",
			StopReason: "end_turn",
			Message:    domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		if m.State() != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.State())
		}
	})

	t.Run("StreamCompleted with tool_use stays active", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:     "user-1",
			StopReason: "tool_use",
			Message:    domain.Message{ID: "asst-1", StopReason: "tool_use"},
		})

		if m.State() != StateActive {
			t.Errorf("expected StateActive after tool_use, got %d", m.State())
		}
	})

	t.Run("ignores messages for unknown turns", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:     "unknown-turn",
			StopReason: "end_turn",
			Message:    domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		if m.State() != StateActive {
			t.Errorf("expected StateActive (unchanged), got %d", m.State())
		}
	})

	t.Run("skips forwarding to turns after cancel", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()

		// StreamCompleted for our turn should not change state back
		m.Update(msgs.StreamCompleted{
			TurnID:     "user-1",
			StopReason: "end_turn",
			Message:    domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		// State should still be cancelled, not complete
		if m.State() != StateCancelled {
			t.Errorf("expected StateCancelled (unchanged), got %d", m.State())
		}
	})
}

func TestDuration(t *testing.T) {
	t.Parallel()

	t.Run("returns positive duration while active", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		if m.Duration() <= 0 {
			t.Error("expected positive duration while active")
		}
	})

	t.Run("returns fixed duration after cancel", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Cancel()
		d1 := m.Duration()
		d2 := m.Duration()

		if d1 != d2 {
			t.Error("expected fixed duration after cancel")
		}
	})
}
