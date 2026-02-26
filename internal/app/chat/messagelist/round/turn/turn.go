package turn

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/user"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// State represents the current state of a turn.
type State int

const (
	StateIdle State = iota
	StateStreaming
	StateAwaitingToolResults
	StateComplete
)

const dbOpTimeout = 2 * time.Second

// Model represents a single user→assistant exchange.
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme styles.Theme
	scope log.Scope

	conversationID domain.ConversationID
	accountID      domain.AccountID

	userMessage      *user.Model
	assistantMessage *assistant.Model

	state  State
	width  int
	stream *streamState

	// Tool result collection
	pendingTools     int
	pendingToolIDs   map[string]bool // tool_use IDs belonging to this turn
	toolResults      []tools.Result
	persisted        bool // true after assistantMessage is written to DB
	firedToolResults bool // true after fireToolResults has been called

	// Protocol guard telemetry (incremented on dropped/malformed lifecycle events).
	protocolViolationCount int

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// streamState holds the channel for receiving stream updates.
type streamState struct {
	updates chan streamUpdate
	cancel  context.CancelCauseFunc
	done    bool
}

// streamUpdate is sent through the channel as the stream progresses.
type streamUpdate struct {
	message *domain.Message
	status  chatclient.StreamStatus
	abort   string
	result  *chatclient.StreamResult // final result, only set on done
	err     error
	done    bool
}

var errUserCancelled = errors.New("user_cancelled")

// streamUpdateMsg is the internal message for stream handling.
type streamUpdateMsg struct {
	turnID domain.MessageID
	update streamUpdate
}

// assistantPersisted is fired after the assistant message is written to the DB.
// It gates fireToolResults to prevent racing with persistence.
type assistantPersisted struct {
	turnID    domain.MessageID
	messageID domain.MessageID
}

// New creates a new turn from a user submission.
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
	scope = scope.Child("turn")
	return &Model{
		theme:            theme,
		scope:            scope,
		conversationID:   conversationID,
		accountID:        accountID,
		userMessage:      user.New(theme.WithBg(theme.BgElevated), userMessageID, input, width-block.BorderWidth),
		assistantMessage: assistant.New(theme, userMessageID, "", width, toolRegistry, scope),
		state:            StateIdle,
		width:            width,
		db:               db,
		chatClient:       chatClient,
		toolRegistry:     toolRegistry,
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case streamUpdateMsg:
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
		if msg.GetTurnID() != m.userMessage.ID() {
			m.reportProtocolViolation(
				"tool_completed_turn_mismatch",
				"received_turn_id", msg.GetTurnID(),
				"expected_turn_id", m.userMessage.ID(),
				"tool_use_id", msg.GetToolUseID(),
			)
			return nil
		}
		cmds = append(cmds, m.handleToolCompleted(msg.GetToolUseID(), msg.GetResult()))

	case assistantPersisted:
		if msg.turnID != m.userMessage.ID() {
			m.reportProtocolViolation(
				"assistant_persisted_turn_mismatch",
				"received_turn_id", msg.turnID,
				"expected_turn_id", m.userMessage.ID(),
				"message_id", msg.messageID,
			)
			return nil
		}
		m.persisted = true
		if shouldFireToolResults(m.state, m.persisted, len(m.toolResults), m.pendingTools) {
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
	if m.pendingToolIDs != nil && !m.pendingToolIDs[toolUseID] {
		m.reportProtocolViolation(
			"tool_completed_unknown_tool_use_id",
			"tool_use_id", toolUseID,
			"pending", m.pendingTools,
			"collected", len(m.toolResults),
		)
		return nil
	}

	// Collect results during streaming or awaiting - tools may complete before StreamCompleted
	m.toolResults = append(m.toolResults, result)
	m.scope.Info("tool completed", "tool_use_id", toolUseID, "collected", len(m.toolResults), "pending", m.pendingTools)

	// Only fire results once we're awaiting and have all of them
	next := reduceOnToolCompleted(m.state, len(m.toolResults), m.pendingTools)
	if next == m.state {
		return nil
	}
	m.state = next
	if m.state == StateComplete {
		m.scope.Info("all tools completed")
		if shouldFireToolResults(m.state, m.persisted, len(m.toolResults), m.pendingTools) {
			return m.fireToolResults()
		}
		return nil
	}
	return nil
}

// StartStream begins streaming the assistant response.
func (m *Model) StartStream(messages []domain.Message, chatContext []domain.ContextEntity) tea.Cmd {
	m.scope.Debug("starting stream", "message_count", len(messages))
	m.state = StateStreaming

	ctx, cancel := context.WithCancelCause(context.Background())
	updates := make(chan streamUpdate, 10)
	m.stream = &streamState{updates: updates, cancel: cancel}

	go func() {
		defer close(updates)

		req := chatclient.Request{
			ConversationID: m.conversationID.String(),
			Messages:       messages,
			Context:        chatContext,
		}

		var lastSnapshot *chatclient.StreamSnapshot
		result, err := m.chatClient.StreamSnapshots(ctx, req, func(s chatclient.StreamSnapshot) {
			ss := s
			lastSnapshot = &ss
			if !s.Done {
				updates <- streamUpdate{message: s.Message, status: s.Status}
			}
		})

		if err != nil {
			updates <- streamUpdate{err: err, done: true}
		} else {
			var lastMessage *domain.Message
			var status chatclient.StreamStatus
			var abort string
			if lastSnapshot != nil {
				lastMessage = lastSnapshot.Message
				status = lastSnapshot.Status
				abort = lastSnapshot.AbortReason
			}
			updates <- streamUpdate{message: lastMessage, status: status, abort: abort, result: result, done: true}
		}
	}()

	return m.nextStreamUpdate()
}

// Blocks returns all visual blocks for the viewport.
// Includes user message (if visible) followed by assistant blocks.
func (m *Model) Blocks() []block.Block {
	var result []block.Block
	if m.userMessage.IsVisible() {
		result = append(result, m.userMessage)
	}
	result = append(result, m.assistantMessage.Blocks()...)
	return result
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	// User block gets content width minus border+padding, same as assistant blocks.
	// The border decoration is applied by renderBlock in the message list.
	m.userMessage.SetWidth(width - block.BorderWidth)
	m.assistantMessage.SetWidth(width)
}

// State returns the turn's current state.
func (m *Model) State() State {
	return m.state
}

// UserMessageID returns the user message ID.
func (m *Model) UserMessageID() domain.MessageID {
	return m.userMessage.ID()
}

// AssistantMessageID returns the persisted assistant message ID.
// Returns empty string if the assistant message was never persisted.
func (m *Model) AssistantMessageID() domain.MessageID {
	return m.assistantMessage.ID()
}

// Cancel stops the in-flight stream and marks the turn complete.
// The partial content remains rendered but nothing is persisted.
func (m *Model) Cancel() {
	if m.stream != nil && !m.stream.done {
		m.stream.cancel(errUserCancelled)
		m.stream.done = true
	}
	m.assistantMessage.Cancel()
	m.state = StateComplete
}

// handleStreamUpdate processes a stream update and fires messages.
func (m *Model) handleStreamUpdate(update streamUpdate) tea.Cmd {
	if update.err != nil {
		// context.Canceled is expected when Cancel() was called — don't show an error.
		if m.stream != nil && m.stream.done {
			return nil
		}
		errorClass := chatclient.ClassifyStreamError(update.err)
		m.scope.Error("stream error", "class", string(errorClass), "error", update.err)
		m.assistantMessage.Cancel()
		m.state = StateComplete
		turnID := m.userMessage.ID()
		return func() tea.Msg {
			return msgs.StreamFailed{TurnID: turnID, Err: update.err}
		}
	}

	if update.message == nil {
		return m.nextStreamUpdate()
	}

	// Set assistant message ID once we have it from the stream
	if m.assistantMessage.ID() == "" && update.message.ID != "" {
		m.assistantMessage.SetID(update.message.ID)
		// Populate content immediately to avoid empty render before message round-trips
		m.assistantMessage.SetContent(update.message.Content)
	}

	if update.done {
		if update.status == chatclient.StreamStatusAborted {
			reason := update.abort
			if reason == "" {
				reason = "context_canceled"
			}
			m.scope.Info("stream aborted", "reason", reason)
			m.assistantMessage.Cancel()
			m.state = StateComplete

			// User-cancelled stream: keep current behavior (do not persist).
			if reason == errUserCancelled.Error() {
				return nil
			}

			msg := update.message
			if msg == nil {
				msg = &domain.Message{}
			}
			msg.StopReason = "aborted"

			turnID := m.userMessage.ID()
			return tea.Batch(
				func() tea.Msg {
					return msgs.StreamCompleted{
						TurnID:     turnID,
						Message:    *msg,
						StopReason: msg.StopReason,
					}
				},
				m.persistAssistantMessage(msg),
			)
		}

		m.scope.Info("stream completed", "stop_reason", update.message.StopReason)

		// Extract metadata from stream result if present
		var title string
		var contextWindow, inputTokens, outputTokens int
		if update.result != nil && update.result.Metadata != nil {
			title = update.result.Metadata.Title
			contextWindow = update.result.Metadata.ContextWindow
			inputTokens = update.result.Metadata.InputTokens
			outputTokens = update.result.Metadata.OutputTokens
		}

		if update.message.StopReason == "tool_use" {
			m.pendingTools, m.pendingToolIDs = collectToolUseIDs(update.message.Content)
			m.scope.Info("awaiting tool results", "pending", m.pendingTools, "already_collected", len(m.toolResults))
			m.state = reduceOnStreamDone(update.message.StopReason, len(m.toolResults), m.pendingTools)
			if m.state == StateComplete {
				m.scope.Info("all tools already completed")
			}
		} else {
			m.state = reduceOnStreamDone(update.message.StopReason, len(m.toolResults), m.pendingTools)
		}

		// Fire StreamCompleted and persist
		turnID := m.userMessage.ID()
		return tea.Batch(
			func() tea.Msg {
				return msgs.StreamCompleted{
					TurnID:        turnID,
					Message:       *update.message,
					StopReason:    update.message.StopReason,
					Title:         title,
					ContextWindow: contextWindow,
					InputTokens:   inputTokens,
					OutputTokens:  outputTokens,
				}
			},
			m.persistAssistantMessage(update.message),
		)
	}

	// Fire AssistantContentUpdated and continue
	turnID := m.userMessage.ID()
	return tea.Batch(
		func() tea.Msg {
			return msgs.AssistantContentUpdated{TurnID: turnID, Message: *update.message}
		},
		m.nextStreamUpdate(),
	)
}

// nextStreamUpdate returns a command that waits for the next stream update.
func (m *Model) nextStreamUpdate() tea.Cmd {
	if m.stream == nil || m.stream.done {
		return nil
	}

	userMsgID := m.userMessage.ID()
	updates := m.stream.updates

	return func() tea.Msg {
		update, ok := <-updates
		if !ok {
			return streamUpdateMsg{
				turnID: userMsgID,
				update: streamUpdate{done: true},
			}
		}
		return streamUpdateMsg{
			turnID: userMsgID,
			update: update,
		}
	}
}

// persistAssistantMessage saves the assistant message to the database.
func (m *Model) persistAssistantMessage(msg *domain.Message) tea.Cmd {
	if msg == nil {
		return nil
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
		defer cancel()

		msgID, err := m.db.Messages().CreateAssistantMessage(
			ctx,
			m.accountID,
			m.conversationID,
			msg.Model,
		)
		if err != nil {
			m.scope.Error("failed to create assistant message", "error", err)
			return nil
		}

		content, err := domain.EncodeBlocks(msg.Content)
		if err != nil {
			m.scope.Error("failed to encode blocks", "error", err)
			return nil
		}

		if err := m.db.Messages().UpdateContent(ctx, msgID, content); err != nil {
			m.scope.Error("failed to update content", "error", err)
			return nil
		}

		if err := m.db.Messages().UpdateMeta(ctx, msgID, msg.Model, msg.StopReason); err != nil {
			m.scope.Error("failed to update meta", "error", err)
			return nil
		}

		m.scope.Info("assistant message persisted", "message_id", msgID)
		return assistantPersisted{
			turnID:    m.userMessage.ID(),
			messageID: msgID,
		}
	}
}

// fireToolResults fires ToolResultsReady for the round to handle.
// Guarded by firedToolResults to ensure it can only fire once per turn.
func (m *Model) fireToolResults() tea.Cmd {
	if m.firedToolResults {
		m.scope.Warn("fireToolResults called twice, ignoring")
		return nil
	}
	m.firedToolResults = true
	turnID := m.userMessage.ID()
	results := m.toolResults
	return func() tea.Msg {
		return msgs.ToolResultsReady{
			TurnID:  turnID,
			Results: results,
		}
	}
}

// collectToolUseIDs returns the count and set of tool_use IDs in content.
func collectToolUseIDs(content []domain.Block) (int, map[string]bool) {
	ids := make(map[string]bool)
	for _, b := range content {
		if b.Type == domain.BlockTypeToolUse && b.ToolUse != nil && b.ToolUse.ID != "" {
			ids[b.ToolUse.ID] = true
		}
	}
	return len(ids), ids
}

func (m *Model) reportProtocolViolation(reason string, kv ...any) {
	m.protocolViolationCount++
	fields := []any{
		"reason", reason,
		"count", m.protocolViolationCount,
	}
	fields = append(fields, kv...)
	m.scope.Warn("protocol violation", fields...)
}
