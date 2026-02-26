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

// LastRound returns the last round or nil.
func (m *Model) LastRound() *round.Model {
	if len(m.rounds) == 0 {
		return nil
	}
	return m.rounds[len(m.rounds)-1]
}

// HasTurn returns true when any round owns the given turn/user-message ID.
func (m *Model) HasTurn(turnID domain.MessageID) bool {
	for _, r := range m.rounds {
		if r.HasTurn(turnID) {
			return true
		}
	}
	return false
}

// RemoveLastRound removes the last round and rebuilds blocks.
func (m *Model) RemoveLastRound() {
	if len(m.rounds) == 0 {
		return
	}
	m.rounds = m.rounds[:len(m.rounds)-1]
	m.rebuildBlocks()
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
	items := projectItems(m.blocks, m.blockHeight)
	m.layout = projectLayout(items, func(roundIndex int) bool {
		return m.rounds[roundIndex].IsActive()
	})
	m.vp.SetItems(m.layout.heights, m.layout.gapHeights())
	m.vp.SetTrailingHeight(m.layout.trailingHeight())
}

// updateRoundWidths sets the width on all rounds.
func (m *Model) updateRoundWidths() {
	w := m.contentWidth()
	for _, r := range m.rounds {
		r.SetWidth(w)
	}
}
