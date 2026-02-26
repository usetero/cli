package chat

import "testing"

func TestReducerProperty_SequenceMustBeStrictlyIncreasing(t *testing.T) {
	t.Parallel()

	for a := 1; a <= 4; a++ {
		for b := 1; b <= 4; b++ {
			for c := 1; c <= 4; c++ {
				seqs := []int{a, b, c}
				r := newReducer("conv-1")
				_, _ = r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 0, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})

				errAt := -1
				for i, s := range seqs {
					_, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: s, Type: EventTypeMetadataUpdate, Metadata: &metadata{}})
					if err != nil {
						errAt = i
						break
					}
				}

				hasNonIncreasing := b <= a || c <= b
				if hasNonIncreasing && errAt == -1 {
					t.Fatalf("expected error for non-increasing sequence %v", seqs)
				}
				if !hasNonIncreasing && errAt != -1 {
					t.Fatalf("unexpected error for increasing sequence %v", seqs)
				}
			}
		}
	}
}

func TestReducerProperty_TurnIDMustStayStable(t *testing.T) {
	t.Parallel()

	r := newReducer("conv-1")
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}}); err != nil {
		t.Fatalf("unexpected error on first turn event: %v", err)
	}
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 2, Type: EventTypeMetadataUpdate, Metadata: &metadata{}}); err != nil {
		t.Fatalf("unexpected error for same turn: %v", err)
	}
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-2", Seq: 3, Type: EventTypeMetadataUpdate, Metadata: &metadata{}}); err == nil {
		t.Fatal("expected turn mismatch error, got nil")
	}
}

func TestReducerProperty_NoEventsAfterTerminal(t *testing.T) {
	t.Parallel()

	r := newReducer("conv-1")
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 1, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}}); err != nil {
		t.Fatalf("unexpected error before done: %v", err)
	}
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 2, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(10), OutputTokens: intPtr(2)}}); err != nil {
		t.Fatalf("unexpected message_stop error: %v", err)
	}
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 3, Done: true}); err != nil {
		t.Fatalf("unexpected done error: %v", err)
	}
	if _, err := r.apply(event{ConversationID: "conv-1", TurnID: "turn-1", Seq: 4, Type: EventTypeMetadataUpdate, Metadata: &metadata{}}); err == nil {
		t.Fatal("expected error for event after terminal, got nil")
	}
}
