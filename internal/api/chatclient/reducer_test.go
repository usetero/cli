package chat

import (
	"strings"
	"testing"
)

func TestReducer(t *testing.T) {
	t.Parallel()

	t.Run("requires message_start first", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, err := r.apply(event{Type: EventTypeTextDelta, Text: &textContent{Content: strPtr("x")}})
		if err == nil || !strings.Contains(err.Error(), "first event must be message_start") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("builds completed snapshot with metadata", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		events := []event{
			{ConversationID: "conv-1", TurnID: "turn-1", Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}},
			{ConversationID: "conv-1", TurnID: "turn-1", Seq: 2, Type: EventTypeTextDelta, Text: &textContent{Content: strPtr("hello")}},
			{ConversationID: "conv-1", TurnID: "turn-1", Seq: 3, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(10), OutputTokens: intPtr(3)}},
		}
		for _, e := range events {
			if _, err := r.apply(e); err != nil {
				t.Fatalf("apply(%s) error = %v", e.Type, err)
			}
		}

		snap, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 4, Done: true})
		if err != nil {
			t.Fatalf("apply(done) error = %v", err)
		}
		if !snap.Done || snap.Status != StreamStatusCompleted {
			t.Fatalf("unexpected final snapshot: %#v", snap)
		}
		if snap.Metadata == nil || snap.Metadata.ContextWindow != 200000 {
			t.Fatalf("metadata = %#v", snap.Metadata)
		}
	})

	t.Run("maps tool_use stop reason to tool_use status", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_, _ = r.apply(event{Seq: 2, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use", InputTokens: intPtr(10), OutputTokens: intPtr(2)}})
		snap, err := r.apply(event{Seq: 3, Done: true})
		if err != nil {
			t.Fatalf("apply(done) error = %v", err)
		}
		if snap.Status != StreamStatusToolUse {
			t.Fatalf("status = %q", snap.Status)
		}
	})

	t.Run("rejects content after message_stop", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_, _ = r.apply(event{Seq: 2, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(10), OutputTokens: intPtr(2)}})
		_, err := r.apply(event{Seq: 3, Type: EventTypeTextDelta, Text: &textContent{Content: strPtr("late")}})
		if err == nil || !strings.Contains(err.Error(), "after message_stop") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("allows metadata_update after message_stop before done", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_, _ = r.apply(event{Seq: 2, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(10), OutputTokens: intPtr(2)}})
		snap, err := r.apply(event{Seq: 3, Type: EventTypeMetadataUpdate, Metadata: &metadata{Title: "hello"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if snap == nil || snap.Metadata == nil || snap.Metadata.Title != "hello" {
			t.Fatalf("expected metadata title on snapshot, got %#v", snap)
		}
		_, err = r.apply(event{Seq: 4, Done: true})
		if err != nil {
			t.Fatalf("unexpected done error: %v", err)
		}
	})

	t.Run("rejects done before message_stop", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_, err := r.apply(event{Seq: 2, Done: true})
		if err == nil || !strings.Contains(err.Error(), "before message_stop") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects non-monotonic seq", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 2, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_, err := r.apply(event{Seq: 2, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(10), OutputTokens: intPtr(2)}})
		if err == nil {
			t.Fatal("expected non-monotonic seq error")
		}
	})

	t.Run("abortSnapshot emits terminal aborted status", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{TurnID: "turn-1", Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		snap := r.abortSnapshot("user_cancelled")
		if snap == nil || snap.Status != StreamStatusAborted || !snap.Done {
			t.Fatalf("unexpected abort snapshot: %#v", snap)
		}
		if again := r.abortSnapshot("user_cancelled"); again != nil {
			t.Fatal("second abortSnapshot should return nil")
		}
	})
}
