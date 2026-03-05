package round

func reduceOnStreamCompleted(current State, ownsTurn bool, stopReason string) (State, bool) {
	if !ownsTurn {
		return current, false
	}
	if stopReason == "tool_use" {
		return current, false
	}
	if current == StateComplete {
		return current, false
	}
	return StateComplete, true
}

func reduceOnStreamFailed(current State, ownsTurn bool) (State, bool) {
	if !ownsTurn {
		return current, false
	}
	if current == StateFailed {
		return current, false
	}
	return StateFailed, true
}

func reduceOnToolResultsReady(current State, ownsTurn bool) (State, bool) {
	if !ownsTurn || current != StateActive {
		return current, false
	}
	return StateAwaitingNextTurn, true
}

func reduceOnNextTurnReady(current State, roundMatches bool) (State, bool) {
	if !roundMatches || current != StateAwaitingNextTurn {
		return current, false
	}
	return StateActive, true
}
