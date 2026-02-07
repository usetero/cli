package messagelist

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/block"
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

func TestUpdate_StreamCompleted(t *testing.T) {
	t.Parallel()

	t.Run("thinking animation removed after StreamCompleted", func(t *testing.T) {
		t.Parallel()

		theme := styles.NewTheme(true)
		scope := logtest.NewScope(t)
		db := dbtest.OpenTestDB(t)

		client := &chattest.MockClient{
			StreamFunc: func(_ context.Context, _ chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error) {
				msg := &domain.Message{
					ID:         "assistant-1",
					StopReason: "end_turn",
					Content: []domain.Block{
						{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}},
					},
				}
				onMessage(msg)
				return &chat.StreamResult{}, nil
			},
		}

		m := New(theme, db, client, nil, scope)
		m.SetSize(80, 40)

		userMsgID := domain.MessageID("user-1")
		m.StartTurn("conv-1", "acct-1", userMsgID, msgs.UserSubmittedInput{Text: "hi"}, nil, nil)

		// After StartTurn, thinking animation should be present.
		if !hasBlockKind(m.blocks, block.KindThinkingAnimation) {
			t.Fatal("expected thinking animation block after StartTurn")
		}

		// StreamCompleted must update streaming state before rebuildBlocks reads it.
		m.Update(msgs.StreamCompleted{
			TurnID:     userMsgID,
			StopReason: "end_turn",
			Message: domain.Message{
				ID:         "assistant-1",
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
