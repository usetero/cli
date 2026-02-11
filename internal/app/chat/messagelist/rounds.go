package messagelist

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
)

// HasActiveRound returns true if the last round is still active.
func (m *Model) HasActiveRound() bool {
	if len(m.rounds) == 0 {
		return false
	}
	return m.rounds[len(m.rounds)-1].IsActive()
}

// CancelActiveRound cancels the last round if it is still active.
func (m *Model) CancelActiveRound() {
	if len(m.rounds) == 0 {
		return
	}
	last := m.rounds[len(m.rounds)-1]
	if last.IsActive() {
		last.Cancel()
		m.rebuildBlocks()
	}
}

// StartTurn creates a new round and begins streaming.
func (m *Model) StartTurn(
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	messages []domain.Message,
	context []domain.ContextEntity,
) tea.Cmd {
	m.scope.Debug("starting turn", "user_message_id", userMessageID)

	r := round.New(
		m.theme,
		conversationID,
		accountID,
		userMessageID,
		input,
		m.contentWidth(),
		m.db,
		m.chatClient,
		m.toolRegistry,
		m.scope,
	)

	m.rounds = append(m.rounds, r)
	m.rebuildBlocks()

	startCmd := func() tea.Msg {
		return msgs.TurnStarted{
			UserMessageID:  userMessageID,
			ConversationID: conversationID,
		}
	}

	return tea.Batch(r.Init(), r.StartStream(messages, context), startCmd)
}

// rebuildBlocks collects blocks from the round hierarchy into a flat list
// and syncs the viewport with current heights/gaps.
func (m *Model) rebuildBlocks() {
	var entries []blockEntry
	for i, r := range m.rounds {
		for _, b := range r.Blocks() {
			entries = append(entries, blockEntry{block: b, roundIndex: i})
		}
	}
	m.blocks = entries
	m.syncViewportItems()
}

// syncViewportItems rebuilds the viewport's height/gap slices from the
// current block list. Called after rebuildBlocks and after toggles
// (which change block heights without changing the block list).
func (m *Model) syncViewportItems() {
	n := len(m.blocks)
	heights := make([]int, n)
	gaps := make([]int, n)
	for i := range m.blocks {
		heights[i] = m.blockHeight(i)
		if i > 0 {
			gaps[i] = m.gapSize(i)
		}
	}
	m.vp.SetItems(heights, gaps)
	m.vp.SetTrailingHeight(m.trailingHeight())
}

// updateRoundWidths sets the width on all rounds.
func (m *Model) updateRoundWidths() {
	w := m.contentWidth()
	for _, r := range m.rounds {
		r.SetWidth(w)
	}
}
