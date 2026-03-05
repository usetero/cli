package messagelist

import (
	"testing"

	msgs "github.com/usetero/cli/internal/app/chat/events"
)

func TestReduceLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("turn started always rebuilds and scrolls", func(t *testing.T) {
		t.Parallel()
		d := reduceLifecycle(msgs.TurnStarted{}, false)
		if !d.handle || !d.rebuild || !d.clearSelection || !d.scrollToBottom || !d.focusLastAtBottom {
			t.Fatalf("unexpected decision: %+v", d)
		}
		if d.forwardRounds {
			t.Fatalf("forwardRounds = true, want false: %+v", d)
		}
	})

	t.Run("assistant update forwards and preserves bottom stickiness", func(t *testing.T) {
		t.Parallel()
		dTop := reduceLifecycle(msgs.AssistantContentUpdated{}, false)
		if !dTop.handle || !dTop.forwardRounds || !dTop.rebuild {
			t.Fatalf("unexpected top decision: %+v", dTop)
		}
		if dTop.scrollToBottom || dTop.focusLastAtBottom {
			t.Fatalf("top decision should not force bottom: %+v", dTop)
		}

		dBottom := reduceLifecycle(msgs.StreamCompleted{}, true)
		if !dBottom.scrollToBottom || !dBottom.focusLastAtBottom {
			t.Fatalf("bottom decision should stick bottom: %+v", dBottom)
		}
	})

	t.Run("unrelated messages are ignored", func(t *testing.T) {
		t.Parallel()
		d := reduceLifecycle(struct{}{}, true)
		if d.handle {
			t.Fatalf("expected no-op decision, got %+v", d)
		}
	})
}
