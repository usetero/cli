package chat

import "testing"

func TestReducerProperty_SequenceMustBeStrictlyIncreasing(t *testing.T) {
	t.Parallel()

	// Check a broad set of short sequences deterministically.
	// Property: any non-increasing step must be rejected.
	for a := 1; a <= 4; a++ {
		for b := 1; b <= 4; b++ {
			for c := 1; c <= 4; c++ {
				seqs := []int{a, b, c}
				r := newReducer("conv-1")
				errAt := -1
				for i, s := range seqs {
					_, err := r.apply(event{
						ConversationID: "conv-1",
						TurnID:         "turn-1",
						Seq:            s,
						Type:           EventTypeTextDelta,
						Text:           &textContent{Content: "x"},
					})
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
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Seq:            1,
		Type:           EventTypeTextDelta,
		Text:           &textContent{Content: "a"},
	}); err != nil {
		t.Fatalf("unexpected error on first turn event: %v", err)
	}

	// Same turn is always valid.
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Seq:            2,
		Type:           EventTypeTextDelta,
		Text:           &textContent{Content: "b"},
	}); err != nil {
		t.Fatalf("unexpected error for same turn: %v", err)
	}

	// Different turn must be rejected.
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-2",
		Seq:            3,
		Type:           EventTypeTextDelta,
		Text:           &textContent{Content: "c"},
	}); err == nil {
		t.Fatal("expected turn mismatch error, got nil")
	}
}

func TestReducerProperty_NoEventsAfterTerminal(t *testing.T) {
	t.Parallel()

	r := newReducer("conv-1")
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Seq:            1,
		Type:           EventTypeMessageStop,
		MessageStop:    &messageStop{StopReason: "end_turn"},
	}); err != nil {
		t.Fatalf("unexpected error before done: %v", err)
	}
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Seq:            2,
		Done:           true,
	}); err != nil {
		t.Fatalf("unexpected done error: %v", err)
	}
	if _, err := r.apply(event{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Seq:            3,
		Type:           EventTypeTextDelta,
		Text:           &textContent{Content: "late"},
	}); err == nil {
		t.Fatal("expected error for event after terminal, got nil")
	}
}
