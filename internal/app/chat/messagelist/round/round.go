package round

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// State represents the current state of a round.
type State int

const (
	StateActive           State = iota
	StateAwaitingNextTurn       // async DB work in progress before next turn
	StateComplete
	StateCancelled
)

// IsActive returns true if the round is in-flight (active or awaiting next turn).
func (m *Model) IsActive() bool {
	return m.state == StateActive || m.state == StateAwaitingNextTurn
}

// Model represents a complete user→assistant exchange, potentially with multiple turns
// if tools are involved. A round starts with explicit user input and ends when the
// assistant stops (no more tool calls).
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme styles.Theme
	scope log.Scope

	id             domain.MessageID // first user message ID, identifies the round
	conversationID domain.ConversationID
	accountID      domain.AccountID

	turns    []*turn.Model
	thinking *thinking.Model
	state    State
	width    int

	startTime time.Time
	endTime   time.Time

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// New creates a new round from explicit user input.
func New(
	theme styles.Theme,
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	width int,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *chattools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("round")

	// Create first turn with user's explicit input
	firstTurn := turn.New(
		theme,
		conversationID,
		accountID,
		userMessageID,
		input,
		width,
		db,
		chatClient,
		toolRegistry,
		scope,
	)

	return &Model{
		theme:          theme,
		scope:          scope,
		id:             userMessageID,
		conversationID: conversationID,
		accountID:      accountID,
		turns:          []*turn.Model{firstTurn},
		thinking:       thinking.New(theme, thinking.Settings{Label: "Thinking"}),
		state:          StateActive,
		width:          width,
		startTime:      time.Now(),
		db:             db,
		chatClient:     chatClient,
		toolRegistry:   toolRegistry,
	}
}

// Init starts the thinking animation.
func (m *Model) Init() tea.Cmd {
	return m.thinking.Init()
}

// StartStream begins streaming for the first turn.
func (m *Model) StartStream(messages []domain.Message, context []domain.ContextEntity) tea.Cmd {
	if len(m.turns) == 0 {
		return nil
	}
	return m.turns[0].StartStream(messages, context)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	// Cancelled rounds are fully stopped — no state transitions, no forwarding.
	if m.state == StateCancelled {
		return nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.StreamCompleted:
		// Check if this is for our current turn
		if m.isOurTurn(msg.TurnID) {
			if msg.StopReason != "tool_use" {
				// Round complete - no more tool calls
				m.state = StateComplete
				m.endTime = time.Now()
				m.scope.Info("round complete", "stop_reason", msg.StopReason)
			}
			// If tool_use, turn will collect results and fire ToolResultsReady
		}

	case msgs.StreamFailed:
		if m.isOurTurn(msg.TurnID) {
			m.state = StateComplete
			m.endTime = time.Now()
			m.scope.Info("round failed", "error", msg.Err)
		}

	case msgs.ToolResultsReady:
		// Tool results collected by one of our turns - start next turn.
		// Transition to StateAwaitingNextTurn synchronously to prevent
		// a second ToolResultsReady from triggering a duplicate startNextTurn.
		if m.isOurTurn(msg.TurnID) && m.state == StateActive {
			m.state = StateAwaitingNextTurn
			cmds = append(cmds, m.startNextTurn(msg.Results))
		}

	case nextTurnReady:
		// Internal message after persistence - create and start next turn.
		// Only proceed if still awaiting (cancel may have intervened).
		if msg.roundID == m.id && m.state == StateAwaitingNextTurn {
			m.state = StateActive
			cmds = append(cmds, m.handleNextTurnReady(msg))
		}
	}

	// Forward thinking ticks while active
	if m.IsActive() {
		cmds = append(cmds, m.thinking.Update(msg))
	}

	// Forward to all turns
	for _, t := range m.turns {
		cmds = append(cmds, t.Update(msg))
	}

	return tea.Batch(cmds...)
}

// isOurTurn checks if the given turn ID belongs to this round.
func (m *Model) isOurTurn(turnID domain.MessageID) bool {
	for _, t := range m.turns {
		if t.UserMessageID() == turnID {
			return true
		}
	}
	return false
}

// startNextTurn persists tool results, loads messages, and creates the next turn.
func (m *Model) startNextTurn(results []domaintools.Result) tea.Cmd {
	m.scope.Info("starting next turn", "result_count", len(results))

	return func() tea.Msg {
		ctx := context.Background()

		// Convert to domain format and persist
		domainResults := make([]domain.ToolResult, len(results))
		for i, r := range results {
			domainResults[i] = domain.ToolResult{
				ToolUseID: r.ToolUseID,
				IsError:   r.IsError(),
				Content:   r.ToMap(),
			}
			if r.Error != nil {
				domainResults[i].Error = r.Error.Message
			}
		}

		msgID, err := m.db.Messages().CreateToolResultMessage(ctx, m.accountID, m.conversationID, domainResults)
		if err != nil {
			m.scope.Error("failed to create tool result message", "error", err)
			return nil
		}

		// Load all messages for the API call
		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.scope.Error("failed to load messages", "error", err)
			return nil
		}

		return nextTurnReady{
			roundID:   m.id,
			messageID: msgID,
			results:   results,
			messages:  messages,
		}
	}
}

// nextTurnReady is an internal message to create the next turn after persistence.
type nextTurnReady struct {
	roundID   domain.MessageID
	messageID domain.MessageID
	results   []domaintools.Result
	messages  []domain.Message
}

// handleNextTurnReady creates and starts the next turn.
func (m *Model) handleNextTurnReady(msg nextTurnReady) tea.Cmd {
	// Create input with tool results (empty text)
	input := msgs.UserSubmittedInput{
		ToolResults: msg.results,
	}

	nextTurn := turn.New(
		m.theme,
		m.conversationID,
		m.accountID,
		msg.messageID,
		input,
		m.width,
		m.db,
		m.chatClient,
		m.toolRegistry,
		m.scope,
	)

	m.turns = append(m.turns, nextTurn)

	return nextTurn.StartStream(msg.messages, nil)
}

// Blocks returns all visual blocks from all turns in this round.
// The thinking animation is appended at the end while the round is active.
func (m *Model) Blocks() []block.Block {
	var result []block.Block
	for _, t := range m.turns {
		result = append(result, t.Blocks()...)
	}
	if m.IsActive() {
		result = append(result, blocks.NewThinkingAnimBlock(m.thinking))
	}
	return result
}

// SetWidth sets the width for all turns.
func (m *Model) SetWidth(width int) {
	m.width = width
	for _, t := range m.turns {
		t.SetWidth(width)
	}
}

// Cancel stops all in-flight turns and marks the round cancelled.
func (m *Model) Cancel() {
	for _, t := range m.turns {
		t.Cancel()
	}
	m.state = StateCancelled
	m.endTime = time.Now()
	m.scope.Info("round cancelled")
}

// State returns the round's current state.
func (m *Model) State() State {
	return m.state
}

// ID returns the round's ID (first user message ID).
func (m *Model) ID() domain.MessageID {
	return m.id
}

// Duration returns the elapsed time for this round.
func (m *Model) Duration() time.Duration {
	if m.endTime.IsZero() {
		return time.Since(m.startTime)
	}
	return m.endTime.Sub(m.startTime)
}
