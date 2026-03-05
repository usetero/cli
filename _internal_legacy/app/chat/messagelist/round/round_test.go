package round

import (
	"errors"
	"testing"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/usecase"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func newTestRound(t *testing.T) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	input := msgs.UserSubmittedInput{Text: "hello"}
	return New(theme, "conv-1", "acct-1", "user-1", input, 80, usecase.RuntimeDeps{}, nil, scope)
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
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
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
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		if m.State() != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.State())
		}
	})

	t.Run("StreamCompleted with tool_use stays active", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "tool_use"},
		})

		if m.State() != StateActive {
			t.Errorf("expected StateActive after tool_use, got %d", m.State())
		}
	})

	t.Run("ignores messages for unknown turns", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamCompleted{
			TurnID:  "unknown-turn",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
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
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		// State should still be cancelled, not complete
		if m.State() != StateCancelled {
			t.Errorf("expected StateCancelled (unchanged), got %d", m.State())
		}
	})
}

func TestToolResultsReadyDoesNotDoubleFire(t *testing.T) {
	t.Parallel()

	t.Run("transitions to StateAwaitingNextTurn", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.ToolResultsReady{
			TurnID:  "user-1",
			Results: []tools.Result{{ToolUseID: "tool-1"}},
		})

		if m.state != StateAwaitingNextTurn {
			t.Fatalf("expected StateAwaitingNextTurn, got %d", m.state)
		}
	})

	t.Run("second ToolResultsReady is ignored", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.ToolResultsReady{
			TurnID:  "user-1",
			Results: []tools.Result{{ToolUseID: "tool-1"}},
		})
		m.Update(msgs.ToolResultsReady{
			TurnID:  "user-1",
			Results: []tools.Result{{ToolUseID: "tool-1"}},
		})

		if m.state != StateAwaitingNextTurn {
			t.Fatalf("expected StateAwaitingNextTurn (unchanged), got %d", m.state)
		}
		if len(m.turns) != 1 {
			t.Errorf("expected 1 turn (no duplicate), got %d", len(m.turns))
		}
	})
}

func TestStateAwaitingNextTurn(t *testing.T) {
	t.Parallel()

	t.Run("IsActive returns true", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)
		m.state = StateAwaitingNextTurn

		if !m.IsActive() {
			t.Error("expected IsActive() true for StateAwaitingNextTurn")
		}
	})

	t.Run("shows thinking animation", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)
		m.state = StateAwaitingNextTurn

		if !hasBlockKind(m.Blocks(), block.KindThinkingAnimation) {
			t.Error("expected thinking animation in StateAwaitingNextTurn")
		}
	})

	t.Run("cancel works from awaiting state", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)
		m.state = StateAwaitingNextTurn

		m.Cancel()

		if m.State() != StateCancelled {
			t.Errorf("expected StateCancelled, got %d", m.State())
		}
	})

	t.Run("nextTurnReady ignored after cancel", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)
		m.state = StateAwaitingNextTurn

		m.Cancel()

		m.Update(nextTurnReady{
			roundID:   "user-1",
			messageID: "tool-result-1",
			messages:  []domain.Message{},
		})

		if m.State() != StateCancelled {
			t.Errorf("expected StateCancelled (unchanged), got %d", m.State())
		}
		if len(m.turns) != 1 {
			t.Errorf("expected 1 turn (nextTurnReady ignored), got %d", len(m.turns))
		}
	})
}

func TestStreamFailed(t *testing.T) {
	t.Parallel()

	t.Run("transitions to StateFailed", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamFailed{TurnID: "user-1", Err: errors.New("connection lost")})

		if m.State() != StateFailed {
			t.Errorf("expected StateFailed, got %d", m.State())
		}
	})

	t.Run("stores the error", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		err := errors.New("connection lost")
		m.Update(msgs.StreamFailed{TurnID: "user-1", Err: err})

		if !errors.Is(m.Err(), err) {
			t.Errorf("expected stored error %v, got %v", err, m.Err())
		}
	})

	t.Run("excludes thinking animation", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamFailed{TurnID: "user-1", Err: errors.New("fail")})

		if hasBlockKind(m.Blocks(), block.KindThinkingAnimation) {
			t.Error("thinking animation should be removed after failure")
		}
	})

	t.Run("ignores subsequent messages", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		m.Update(msgs.StreamFailed{TurnID: "user-1", Err: errors.New("fail")})
		m.Update(msgs.StreamCompleted{
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		if m.State() != StateFailed {
			t.Errorf("expected StateFailed (unchanged), got %d", m.State())
		}
	})
}

func TestHasAssistantContent(t *testing.T) {
	t.Parallel()

	t.Run("false for fresh round", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		if m.HasAssistantContent() {
			t.Error("expected false for fresh round with no assistant content")
		}
	})
}

func TestLastTurnMessageIDs(t *testing.T) {
	t.Parallel()

	t.Run("returns user message ID for turn 1", func(t *testing.T) {
		t.Parallel()
		m := newTestRound(t)

		ids := m.LastTurnMessageIDs()

		if len(ids) != 1 {
			t.Fatalf("expected 1 ID, got %d", len(ids))
		}
		if ids[0] != "user-1" {
			t.Errorf("expected user-1, got %s", ids[0])
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

func TestSummarizeToolResults(t *testing.T) {
	t.Parallel()

	results := []tools.Result{
		{
			ToolUseID: "tool-1",
			Content: map[string]any{
				"rows": []map[string]any{{"service_id": "ad"}, {"service_id": "email"}},
			},
		},
		{
			ToolUseID: "tool-2",
			Error:     &tools.ErrorResult{Message: "boom"},
		},
	}

	summaries := summarizeToolResults(results)
	if len(summaries) != 2 {
		t.Fatalf("len(summaries)=%d, want 2", len(summaries))
	}
	if got := summaries[0]; got != "tool_use_id=tool-1 is_error=false rows=2" {
		t.Fatalf("summaries[0]=%q", got)
	}
	if got := summaries[1]; got != "tool_use_id=tool-2 is_error=true" {
		t.Fatalf("summaries[1]=%q", got)
	}
}
