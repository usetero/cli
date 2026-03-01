package turn

import (
	tea "charm.land/bubbletea/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/domain/tools"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case turnStreamUpdateMsg:
		if msg.turnID != m.userMessage.ID() {
			return nil
		}
		cmds = append(cmds, m.handleStreamUpdate(msg.update))

	case msgs.AssistantContentUpdated:
		if msg.TurnID != m.userMessage.ID() {
			return nil
		}
		cmds = append(cmds, m.assistantMessage.Update(msg))
		return tea.Batch(cmds...)

	case msgs.StreamCompleted:
		if msg.TurnID != m.userMessage.ID() {
			return nil
		}
		cmds = append(cmds, m.assistantMessage.Update(msg))
		return tea.Batch(cmds...)

	case msgs.ToolCompleted:
		if msg.TurnID != m.userMessage.ID() {
			// Message buses fan out to all turns; non-owner turns ignore.
			return nil
		}
		cmds = append(cmds, m.handleToolCompleted(msg.ToolUseID, msg.ResultOrError()))

	case assistantPersisted:
		if msg.turnID != m.userMessage.ID() {
			// Internal completion events are broadcast; non-owner turns ignore.
			return nil
		}
		if msg.messageID != "" {
			m.assistantMessage.SetID(msg.messageID)
		}
		m.toolTracker.markPersisted()
		if m.toolTracker.shouldFire(m.state) {
			return m.fireToolResults()
		}
		return nil
	}

	cmds = append(cmds, m.userMessage.Update(msg))
	cmds = append(cmds, m.assistantMessage.Update(msg))

	return tea.Batch(cmds...)
}

func (m *Model) handleToolCompleted(toolUseID string, result tools.Result) tea.Cmd {
	// Ignore tools that don't belong to this turn.
	// Before pendingToolIDs is set (during streaming), accept all tools —
	// they'll be validated once the stream completes and IDs are known.
	if !m.toolTracker.accepts(toolUseID) {
		m.reportProtocolViolation(
			"tool_completed_unknown_tool_use_id",
			"tool_use_id", toolUseID,
			"pending", m.toolTracker.pendingCount(),
			"collected", m.toolTracker.collectedCount(),
		)
		return nil
	}

	// Collect results during streaming or awaiting - tools may complete before StreamCompleted
	m.toolTracker.collect(result)
	m.scope.Info("tool completed", "tool_use_id", toolUseID, "collected", m.toolTracker.collectedCount(), "pending", m.toolTracker.pendingCount())

	// Only fire results once we're awaiting and have all of them
	next := reduceOnToolCompleted(m.state, m.toolTracker.collectedCount(), m.toolTracker.pendingCount())
	if next == m.state {
		return nil
	}
	m.state = next
	if m.state == StateComplete {
		m.scope.Info("all tools completed")
		if m.toolTracker.shouldFire(m.state) {
			return m.fireToolResults()
		}
		return nil
	}
	return nil
}
