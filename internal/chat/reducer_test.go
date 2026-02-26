package chat

import "testing"

func TestReducer(t *testing.T) {
	t.Parallel()

	t.Run("builds completed snapshot with metadata", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")

		_, err := r.apply(event{
			ConversationID: "conv-1",
			TurnID:         "turn-1",
			Seq:            1,
			Type:           EventTypeMessageStart,
			MessageStart:   &messageStart{Model: "claude-3", ContextWindow: 200000},
		})
		if err != nil {
			t.Fatalf("apply(message_start) error = %v", err)
		}

		_, err = r.apply(event{
			ConversationID: "conv-1",
			TurnID:         "turn-1",
			Seq:            2,
			Type:           EventTypeTextDelta,
			Text:           &textContent{Content: "hello"},
		})
		if err != nil {
			t.Fatalf("apply(text_delta) error = %v", err)
		}

		_, err = r.apply(event{
			ConversationID: "conv-1",
			TurnID:         "turn-1",
			Seq:            3,
			Type:           EventTypeMessageStop,
			MessageStop:    &messageStop{StopReason: "end_turn", InputTokens: 10, OutputTokens: 3},
		})
		if err != nil {
			t.Fatalf("apply(message_stop) error = %v", err)
		}

		snap, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 4, Done: true})
		if err != nil {
			t.Fatalf("apply(done) error = %v", err)
		}

		if !snap.Done {
			t.Fatal("Done = false, want true")
		}
		if snap.Status != StreamStatusCompleted {
			t.Fatalf("Status = %q, want %q", snap.Status, StreamStatusCompleted)
		}
		if snap.ConversationID != "conv-1" {
			t.Fatalf("ConversationID = %q, want conv-1", snap.ConversationID)
		}
		if snap.TurnID != "turn-1" {
			t.Fatalf("TurnID = %q, want turn-1", snap.TurnID)
		}
		if snap.Message == nil || len(snap.Message.Content) != 1 {
			t.Fatalf("Message blocks = %v, want 1 block", snap.Message)
		}
		if snap.Metadata == nil {
			t.Fatal("Metadata = nil, want non-nil")
		}
		if snap.Metadata.ContextWindow != 200000 {
			t.Fatalf("ContextWindow = %d, want 200000", snap.Metadata.ContextWindow)
		}
	})

	t.Run("maps tool_use stop reason to tool_use status", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, _ = r.apply(event{Seq: 1, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use"}})
		snap, err := r.apply(event{Seq: 2, Done: true})
		if err != nil {
			t.Fatalf("apply(done) error = %v", err)
		}
		if snap.Status != StreamStatusToolUse {
			t.Fatalf("Status = %q, want %q", snap.Status, StreamStatusToolUse)
		}
	})

	t.Run("rejects non-monotonic seq", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, err := r.apply(event{Seq: 2, Type: EventTypeTextDelta, Text: &textContent{Content: "a"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = r.apply(event{Seq: 2, Type: EventTypeTextDelta, Text: &textContent{Content: "b"}})
		if err == nil {
			t.Fatal("expected non-monotonic seq error, got nil")
		}
	})

	t.Run("rejects turn mismatch", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, err := r.apply(event{TurnID: "turn-1", Type: EventTypeTextDelta, Text: &textContent{Content: "a"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = r.apply(event{TurnID: "turn-2", Type: EventTypeTextDelta, Text: &textContent{Content: "b"}})
		if err == nil {
			t.Fatal("expected turn mismatch error, got nil")
		}
	})

	t.Run("abortSnapshot emits terminal aborted status", func(t *testing.T) {
		t.Parallel()

		r := newReducer("conv-1")
		_, err := r.apply(event{TurnID: "turn-1", Seq: 1, Type: EventTypeTextDelta, Text: &textContent{Content: "a"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		snap := r.abortSnapshot("user_cancelled")
		if snap == nil {
			t.Fatal("abortSnapshot = nil, want snapshot")
		}
		if !snap.Done {
			t.Fatal("Done = false, want true")
		}
		if snap.Status != StreamStatusAborted {
			t.Fatalf("Status = %q, want %q", snap.Status, StreamStatusAborted)
		}
		if snap.AbortReason != "user_cancelled" {
			t.Fatalf("AbortReason = %q, want user_cancelled", snap.AbortReason)
		}
		if snap.TurnID != "turn-1" {
			t.Fatalf("TurnID = %q, want turn-1", snap.TurnID)
		}

		if again := r.abortSnapshot("user_cancelled"); again != nil {
			t.Fatal("second abortSnapshot should return nil")
		}
	})
}
