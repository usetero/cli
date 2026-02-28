package round

import "testing"

func TestRoundReducer(t *testing.T) {
	t.Parallel()

	t.Run("stream completed tool_use stays active", func(t *testing.T) {
		next, changed := reduceOnStreamCompleted(StateActive, true, "tool_use")
		if next != StateActive || changed {
			t.Fatalf("next=%v changed=%v", next, changed)
		}
	})

	t.Run("stream completed end_turn completes", func(t *testing.T) {
		next, changed := reduceOnStreamCompleted(StateActive, true, "end_turn")
		if next != StateComplete || !changed {
			t.Fatalf("next=%v changed=%v", next, changed)
		}
	})

	t.Run("tool results ready only from active", func(t *testing.T) {
		next, changed := reduceOnToolResultsReady(StateAwaitingNextTurn, true)
		if next != StateAwaitingNextTurn || changed {
			t.Fatalf("next=%v changed=%v", next, changed)
		}
	})

	t.Run("next turn ready only from awaiting", func(t *testing.T) {
		next, changed := reduceOnNextTurnReady(StateAwaitingNextTurn, true)
		if next != StateActive || !changed {
			t.Fatalf("next=%v changed=%v", next, changed)
		}
	})
}
