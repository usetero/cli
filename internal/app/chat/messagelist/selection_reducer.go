package messagelist

import tea "charm.land/bubbletea/v2"

type selectionState struct {
	mouseDown      bool
	mouseDownBlock int
	mouseDownX     int
	mouseDownY     int
	mouseDragBlock int
	mouseDragX     int
	mouseDragY     int
}

type selectionPoint struct {
	block int
	x     int
	y     int
}

type clickDecision struct {
	setFocusIdx bool
	focusIdx    int
}

type releaseDecision struct {
	handle     bool
	clickBlock int
	clickY     int
}

type releaseAction int

const (
	releaseActionNoop releaseAction = iota
	releaseActionCopy
	releaseActionClick
)

func reduceSelectionClick(state selectionState, button tea.MouseButton, point selectionPoint, hit bool) (selectionState, clickDecision) {
	if button != tea.MouseLeft || !hit {
		return state, clickDecision{}
	}

	state.mouseDown = true
	state.mouseDownBlock = point.block
	state.mouseDownX = point.x
	state.mouseDownY = point.y
	state.mouseDragBlock = point.block
	state.mouseDragX = point.x
	state.mouseDragY = point.y

	return state, clickDecision{setFocusIdx: true, focusIdx: point.block}
}

func reduceSelectionMotion(state selectionState, button tea.MouseButton, point selectionPoint, hit bool) selectionState {
	if !state.mouseDown || button != tea.MouseLeft || !hit {
		return state
	}

	state.mouseDragBlock = point.block
	state.mouseDragX = point.x
	state.mouseDragY = point.y
	return state
}

func reduceSelectionRelease(state selectionState) (selectionState, releaseDecision) {
	if !state.mouseDown {
		return state, releaseDecision{}
	}

	decision := releaseDecision{
		handle:     true,
		clickBlock: state.mouseDownBlock,
		clickY:     state.mouseDownY,
	}
	state.mouseDown = false
	return state, decision
}

func reduceReleaseAction(hasHighlight bool, highlightedText string) releaseAction {
	if !hasHighlight {
		return releaseActionClick
	}
	if highlightedText == "" {
		// Treat empty extraction as a click so collapsible blocks still toggle.
		return releaseActionClick
	}
	return releaseActionCopy
}
