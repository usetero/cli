package turn

// reduceOnStreamDone returns the turn state after the stream finishes.
func reduceOnStreamDone(stopReason string, collected, pending int) State {
	if stopReason != "tool_use" {
		return StateComplete
	}
	if collected >= pending {
		return StateComplete
	}
	return StateAwaitingToolResults
}

// reduceOnToolCompleted returns the turn state after collecting one tool result.
func reduceOnToolCompleted(current State, collected, pending int) State {
	if current != StateAwaitingToolResults {
		return current
	}
	if collected >= pending {
		return StateComplete
	}
	return StateAwaitingToolResults
}

// shouldFireToolResults returns true when results can be emitted to the round.
func shouldFireToolResults(state State, persisted bool, collected, pending int) bool {
	return state == StateComplete && persisted && pending > 0 && collected >= pending
}
