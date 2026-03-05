package messagelist

import (
	tea "charm.land/bubbletea/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
)

type lifecycleDecision struct {
	handle            bool
	forwardRounds     bool
	rebuild           bool
	clearSelection    bool
	scrollToBottom    bool
	focusLastAtBottom bool
}

func reduceLifecycle(msg tea.Msg, wasAtBottom bool) lifecycleDecision {
	switch msg.(type) {
	case msgs.TurnStarted:
		return lifecycleDecision{
			handle:            true,
			rebuild:           true,
			clearSelection:    true,
			scrollToBottom:    true,
			focusLastAtBottom: true,
		}
	case msgs.AssistantContentUpdated, msgs.StreamCompleted, msgs.StreamFailed:
		return lifecycleDecision{
			handle:            true,
			forwardRounds:     true,
			rebuild:           true,
			scrollToBottom:    wasAtBottom,
			focusLastAtBottom: wasAtBottom,
		}
	default:
		return lifecycleDecision{}
	}
}
