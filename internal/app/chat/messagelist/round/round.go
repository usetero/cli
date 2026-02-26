package round

import (
	"context"
	"fmt"
	"strings"
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
	StateFailed
)

const dbOpTimeout = 2 * time.Second

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
	session  *chatclient.Session // authoritative in-memory history for active tool loop
	thinking *thinking.Model
	state    State
	lastErr  error
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
	m.session = chatclient.NewSession(m.conversationID, messages)
	return m.turns[0].StartStream(messages, context)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	// Terminal states — no state transitions, no forwarding.
	if m.state == StateCancelled || m.state == StateFailed {
		return nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.StreamCompleted:
		// Check if this is for our current turn
		if m.isOurTurn(msg.TurnID) {
			if m.session != nil {
				m.session.RecordAssistantMessage(msg.Message)
			}
			if msg.Message.StopReason != "tool_use" {
				// Round complete - no more tool calls
				m.state = StateComplete
				m.endTime = time.Now()
				m.scope.Info("round complete", "stop_reason", msg.Message.StopReason)
			}
			// If tool_use, turn will collect results and fire ToolResultsReady
		}

	case msgs.StreamFailed:
		if m.isOurTurn(msg.TurnID) {
			m.state = StateFailed
			m.lastErr = msg.Err
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

// HasTurn reports whether this round owns turnID.
func (m *Model) HasTurn(turnID domain.MessageID) bool {
	return m.isOurTurn(turnID)
}

// startNextTurn persists tool results and creates the next turn using in-memory history.
func (m *Model) startNextTurn(results []domaintools.Result) tea.Cmd {
	m.scope.Info("starting next turn", "result_count", len(results))
	for _, summary := range summarizeToolResults(results) {
		m.scope.Debug("next turn tool result", "summary", summary)
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()

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
			// Durability failure should not block the active chat loop.
			m.scope.Error("failed to create tool result message", "error", err)
			msgID = domain.NewMessageID()
		}

		if m.session == nil {
			m.session = chatclient.NewSession(m.conversationID, nil)
		}
		toolResultMessage := m.session.AppendUserToolResultsMessage(msgID, domainResults)
		messages := m.session.Messages()
		for _, summary := range summarizeHistory(messages) {
			m.scope.Debug("next turn history", "summary", summary)
		}

		return nextTurnReady{
			roundID:           m.id,
			messageID:         msgID,
			results:           results,
			messages:          messages,
			toolResultMessage: toolResultMessage,
		}
	}
}

func summarizeToolResults(results []domaintools.Result) []string {
	out := make([]string, 0, len(results))
	for _, r := range results {
		rows := -1
		if rawRows, ok := r.Content["rows"]; ok {
			if list, ok := rawRows.([]map[string]any); ok {
				rows = len(list)
			} else if listAny, ok := rawRows.([]any); ok {
				rows = len(listAny)
			}
		}
		if rows >= 0 {
			out = append(out, fmt.Sprintf("tool_use_id=%s is_error=%t rows=%d", r.ToolUseID, r.IsError(), rows))
			continue
		}
		out = append(out, fmt.Sprintf("tool_use_id=%s is_error=%t", r.ToolUseID, r.IsError()))
	}
	return out
}

func summarizeHistory(messages []domain.Message) []string {
	out := make([]string, 0, len(messages))
	for _, msg := range messages {
		blockKinds := make([]string, 0, len(msg.Content))
		for _, b := range msg.Content {
			blockKinds = append(blockKinds, string(b.Type))
		}
		out = append(out, fmt.Sprintf(
			"id=%s role=%s stop_reason=%s blocks=%d kinds=%s",
			msg.ID,
			msg.Role,
			msg.StopReason,
			len(msg.Content),
			strings.Join(blockKinds, ","),
		))
	}
	return out
}

// nextTurnReady is an internal message to create the next turn after persistence.
type nextTurnReady struct {
	roundID           domain.MessageID
	messageID         domain.MessageID
	results           []domaintools.Result
	messages          []domain.Message
	toolResultMessage domain.Message
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
	startStream := nextTurn.StartStream(msg.messages, nil)
	notifyPersist := func() tea.Msg {
		return msgs.ToolResultMessagePersisted{Message: msg.toolResultMessage}
	}
	return tea.Batch(startStream, notifyPersist)
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

// Err returns the error that caused the round to fail, or nil.
func (m *Model) Err() error {
	return m.lastErr
}

// HasAssistantContent returns true if any turn has assistant blocks.
func (m *Model) HasAssistantContent() bool {
	for _, t := range m.turns {
		if len(t.Blocks()) > 1 { // more than just the user message block
			return true
		}
	}
	return false
}

// LastTurnMessageIDs returns the message IDs that should be deleted on failure.
// For turn 1: the user message ID.
// For turn 2+: the tool result message ID (current turn) + the previous turn's assistant message ID.
func (m *Model) LastTurnMessageIDs() []domain.MessageID {
	if len(m.turns) == 0 {
		return nil
	}

	last := m.turns[len(m.turns)-1]

	if len(m.turns) == 1 {
		// Turn 1: just the user message
		return []domain.MessageID{last.UserMessageID()}
	}

	// Turn 2+: tool result message + previous assistant message
	prev := m.turns[len(m.turns)-2]
	ids := []domain.MessageID{last.UserMessageID()}
	if aid := prev.AssistantMessageID(); aid != "" {
		ids = append(ids, aid)
	}
	return ids
}

// Duration returns the elapsed time for this round.
func (m *Model) Duration() time.Duration {
	if m.endTime.IsZero() {
		return time.Since(m.startTime)
	}
	return m.endTime.Sub(m.startTime)
}
