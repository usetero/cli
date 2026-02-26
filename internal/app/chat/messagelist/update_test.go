package messagelist

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/styles"
)

func hasBlockKind(entries []blockEntry, kind block.Kind) bool {
	for _, e := range entries {
		if e.block.Kind() == kind {
			return true
		}
	}
	return false
}

func countBlockKind(entries []blockEntry, kind block.Kind) int {
	n := 0
	for _, e := range entries {
		if e.block.Kind() == kind {
			n++
		}
	}
	return n
}

// newStreamingMessageList creates a messagelist with a mock client that streams
// one text block then blocks forever. Suitable for testing cancel and thinking lifecycle.
func newStreamingMessageList(t *testing.T) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)
	db := dbtest.OpenTestDB(t)

	client := &chattest.MockClient{
		StreamFunc: func(_ context.Context, _ chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error) {
			msg := &domain.Message{ID: "asst-1", Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}},
			}}
			onMessage(msg)
			// Block forever to simulate in-progress stream
			select {}
		},
	}

	m := New(theme, db, client, nil, scope)
	m.SetSize(80, 40)
	return m
}

func TestUpdate_StreamCompleted(t *testing.T) {
	t.Parallel()

	t.Run("thinking animation removed after StreamCompleted", func(t *testing.T) {
		t.Parallel()

		m := newStreamingMessageList(t)

		userMsgID := domain.MessageID("user-1")
		m.StartTurn("conv-1", "acct-1", userMsgID, msgs.UserSubmittedInput{Text: "hi"}, nil, nil)

		// After StartTurn, thinking animation should be present.
		if !hasBlockKind(m.blocks, block.KindThinkingAnimation) {
			t.Fatal("expected thinking animation block after StartTurn")
		}

		// StreamCompleted must update streaming state before rebuildBlocks reads it.
		m.Update(msgs.StreamCompleted{
			TurnID: userMsgID,
			Message: domain.Message{
				ID:         "asst-1",
				StopReason: "end_turn",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}},
				},
			},
		})

		// Thinking animation should be gone.
		if hasBlockKind(m.blocks, block.KindThinkingAnimation) {
			t.Error("thinking animation block should be removed after StreamCompleted")
		}

		// Text content should remain.
		if !hasBlockKind(m.blocks, block.KindAssistantText) {
			t.Error("expected text block to remain after StreamCompleted")
		}
	})
}

func TestCancelActiveRound(t *testing.T) {
	t.Parallel()

	t.Run("cancels active round and removes thinking", func(t *testing.T) {
		t.Parallel()
		m := newStreamingMessageList(t)

		m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "hi"}, nil, nil)

		if !hasBlockKind(m.blocks, block.KindThinkingAnimation) {
			t.Fatal("expected thinking animation block after StartTurn")
		}

		m.CancelActiveRound()

		if hasBlockKind(m.blocks, block.KindThinkingAnimation) {
			t.Error("thinking animation should be removed after cancel")
		}
		if m.rounds[0].State() != round.StateCancelled {
			t.Errorf("expected StateCancelled, got %d", m.rounds[0].State())
		}
	})

	t.Run("no-op when no rounds exist", func(t *testing.T) {
		t.Parallel()
		m := newStreamingMessageList(t)

		m.CancelActiveRound() // should not panic

		if len(m.rounds) != 0 {
			t.Error("expected no rounds")
		}
	})

	t.Run("no-op when last round is complete", func(t *testing.T) {
		t.Parallel()
		m := newStreamingMessageList(t)

		m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "hi"}, nil, nil)

		// Complete it via StreamCompleted
		m.Update(msgs.StreamCompleted{
			TurnID:  "user-1",
			Message: domain.Message{ID: "asst-1", StopReason: "end_turn"},
		})

		m.CancelActiveRound() // should not change state

		if m.rounds[0].State() != round.StateComplete {
			t.Errorf("expected StateComplete (unchanged), got %d", m.rounds[0].State())
		}
	})

	t.Run("each active round has one thinking block", func(t *testing.T) {
		t.Parallel()
		m := newStreamingMessageList(t)

		m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "first"}, nil, nil)
		m.StartTurn("conv-1", "acct-1", "user-2", msgs.UserSubmittedInput{Text: "second"}, nil, nil)

		count := countBlockKind(m.blocks, block.KindThinkingAnimation)
		if count != 2 {
			t.Errorf("expected 2 thinking blocks (one per active round), got %d", count)
		}
	})
}
