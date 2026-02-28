package turn

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
)

type toolResultTracker struct {
	pendingTools   int
	pendingToolIDs map[string]bool
	results        []tools.Result
	persisted      bool
	fired          bool
}

func (t *toolResultTracker) accepts(toolUseID string) bool {
	// Before pendingToolIDs is set (during streaming), accept all tools.
	if t.pendingToolIDs == nil {
		return true
	}
	return t.pendingToolIDs[toolUseID]
}

func (t *toolResultTracker) collect(result tools.Result) {
	t.results = append(t.results, result)
}

func (t *toolResultTracker) collectedCount() int {
	return len(t.results)
}

func (t *toolResultTracker) pendingCount() int {
	return t.pendingTools
}

func (t *toolResultTracker) setPendingFromContent(content []domain.Block) {
	t.pendingTools, t.pendingToolIDs = collectToolUseIDs(content)
}

func (t *toolResultTracker) markPersisted() {
	t.persisted = true
}

func (t *toolResultTracker) shouldFire(state State) bool {
	return shouldFireToolResults(state, t.persisted, len(t.results), t.pendingTools)
}

func (t *toolResultTracker) fire(turnID domain.MessageID, scopeReporter func(msg string)) tea.Cmd {
	if t.fired {
		if scopeReporter != nil {
			scopeReporter("fireToolResults called twice, ignoring")
		}
		return nil
	}
	t.fired = true
	results := t.results
	return func() tea.Msg {
		return msgs.ToolResultsReady{
			TurnID:  turnID,
			Results: results,
		}
	}
}
