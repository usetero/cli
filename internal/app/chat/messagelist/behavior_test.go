package messagelist

import (
	"fmt"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
)

func addCompletedRound(t *testing.T, m *Model, turnID domain.MessageID, text string) {
	t.Helper()

	m.StartTurn("conv-1", "acct-1", turnID, msgs.UserSubmittedInput{Text: "prompt " + string(turnID)}, nil, nil)
	m.Update(msgs.StreamCompleted{
		TurnID:     turnID,
		StopReason: "end_turn",
		Message: domain.Message{
			ID:         domain.MessageID("asst-" + turnID),
			StopReason: "end_turn",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: text}},
			},
		},
	})
}

func seedHistoryWithActiveRound(t *testing.T, height int) *Model {
	t.Helper()

	m := newStreamingMessageList(t)
	m.SetSize(80, height)

	for i := range 6 {
		id := domain.MessageID(fmt.Sprintf("user-%d", i+1))
		addCompletedRound(t, m, id, fmt.Sprintf("history %d", i+1))
	}

	m.StartTurn("conv-1", "acct-1", "user-live", msgs.UserSubmittedInput{Text: "live"}, nil, nil)
	return m
}

func TestBehavior_CancelledRoundIgnoresStaleAssistantUpdates(t *testing.T) {
	t.Parallel()

	m := newStreamingMessageList(t)

	m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "first"}, nil, nil)
	m.CancelActiveRound()
	m.StartTurn("conv-1", "acct-1", "user-2", msgs.UserSubmittedInput{Text: "second"}, nil, nil)

	m.Update(msgs.AssistantContentUpdated{
		TurnID: "user-1",
		Message: domain.Message{
			ID: "asst-stale",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "stale text"}},
			},
		},
	})

	if len(m.rounds) != 2 {
		t.Fatalf("round count=%d, want 2", len(m.rounds))
	}
	if m.rounds[0].State() != round.StateCancelled {
		t.Fatalf("round[0] state=%v, want cancelled", m.rounds[0].State())
	}
	if !m.rounds[1].IsActive() {
		t.Fatalf("round[1] should still be active")
	}
	if strings.Contains(m.View(), "stale text") {
		t.Fatalf("stale update should not be rendered in view")
	}
}

func TestBehavior_StreamUpdateScrollPolicy(t *testing.T) {
	t.Parallel()

	t.Run("at bottom sticks to bottom on assistant updates", func(t *testing.T) {
		t.Parallel()
		m := seedHistoryWithActiveRound(t, 8)
		m.vp.ScrollToBottom()

		m.Update(msgs.AssistantContentUpdated{
			TurnID: "user-live",
			Message: domain.Message{
				ID: "asst-live",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "live update"}},
				},
			},
		})

		if !m.vp.AtBottom() {
			t.Fatalf("expected to remain at bottom after update")
		}
	})

	t.Run("scrolled up does not get yanked to bottom", func(t *testing.T) {
		t.Parallel()
		m := seedHistoryWithActiveRound(t, 8)
		m.vp.ScrollToBottom()
		m.vp.ScrollBy(-4)
		m.vp.UpdateFocusFromScroll()

		if m.vp.AtBottom() {
			t.Fatalf("precondition failed: expected scrolled-up viewport")
		}
		beforeIdx, beforeLine := m.vp.Offset()

		m.Update(msgs.AssistantContentUpdated{
			TurnID: "user-live",
			Message: domain.Message{
				ID: "asst-live",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "live update"}},
				},
			},
		})

		if m.vp.AtBottom() {
			t.Fatalf("viewport should stay scrolled up after update")
		}
		afterIdx, afterLine := m.vp.Offset()
		if beforeIdx != afterIdx || beforeLine != afterLine {
			t.Fatalf("expected offset stability while scrolled up: before=(%d,%d) after=(%d,%d)", beforeIdx, beforeLine, afterIdx, afterLine)
		}
	})
}
