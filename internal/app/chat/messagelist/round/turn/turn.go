package turn

import (
	"context"

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
	pendingTools int
	toolResults  []tools.Result
	persisted    bool // true after assistantMessage is written to DB

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// streamState holds the channel for receiving stream updates.
type streamState struct {
	updates chan streamUpdate
	cancel  context.CancelFunc
	done    bool
}

// streamUpdate is sent through the channel as the stream progresses.
type streamUpdate struct {
	message *domain.Message
	result  *chatclient.StreamResult // final result, only set on done
	err     error
	done    bool
}

// streamUpdateMsg is the internal message for stream handling.
type streamUpdateMsg struct {
	turnID domain.MessageID
	update streamUpdate
}

// assistantPersisted is fired after the assistant message is written to the DB.
// It gates fireToolResults to prevent racing with persistence.
type assistantPersisted struct {
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
		assistantMessage: assistant.New(theme, "", width, toolRegistry, scope),
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
		cmds = append(cmds, m.handleToolCompleted(msg.GetToolUseID(), msg.GetResult()))

	case assistantPersisted:
		m.persisted = true
		// If tools are already done, fire now that persist is complete.
		if m.state == StateComplete && m.pendingTools > 0 && len(m.toolResults) >= m.pendingTools {
			return m.fireToolResults()
		}
		return nil
	}

	cmds = append(cmds, m.userMessage.Update(msg))
	cmds = append(cmds, m.assistantMessage.Update(msg))

	return tea.Batch(cmds...)
}

func (m *Model) handleToolCompleted(toolUseID string, result tools.Result) tea.Cmd {
	// Collect results during streaming or awaiting - tools may complete before StreamCompleted
	m.toolResults = append(m.toolResults, result)
	m.scope.Info("tool completed", "tool_use_id", toolUseID, "collected", len(m.toolResults), "pending", m.pendingTools)

	// Only fire results once we're awaiting and have all of them
	if m.state != StateAwaitingToolResults {
		return nil
	}

	if len(m.toolResults) >= m.pendingTools {
		m.scope.Info("all tools completed")
		m.state = StateComplete
		if m.persisted {
			return m.fireToolResults()
		}
		return nil // assistantPersisted handler will fire tool results
	}
	return nil
}

// StartStream begins streaming the assistant response.
func (m *Model) StartStream(messages []domain.Message, chatContext []domain.ContextEntity) tea.Cmd {
	m.scope.Debug("starting stream", "message_count", len(messages))
	m.state = StateStreaming

	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan streamUpdate, 10)
	m.stream = &streamState{updates: updates, cancel: cancel}

	go func() {
		defer close(updates)

		req := chatclient.Request{
			ConversationID: m.conversationID.String(),
			Messages:       messages,
			Context:        chatContext,
		}

		var lastMessage *domain.Message
		result, err := m.chatClient.Stream(ctx, req, func(msg *domain.Message) {
			lastMessage = msg
			updates <- streamUpdate{message: msg}
		})

		if err != nil {
			updates <- streamUpdate{err: err, done: true}
		} else {
			updates <- streamUpdate{message: lastMessage, result: result, done: true}
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

// Cancel stops the in-flight stream and marks the turn complete.
// The partial content remains rendered but nothing is persisted.
func (m *Model) Cancel() {
	if m.stream != nil && !m.stream.done {
		m.stream.cancel()
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
		m.scope.Error("stream error", "error", update.err)
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
			m.pendingTools = countToolUseBlocks(update.message.Content)
			m.scope.Info("awaiting tool results", "pending", m.pendingTools, "already_collected", len(m.toolResults))

			// Check if tools already completed during streaming.
			// Don't fire tool results yet — wait for persistAssistantMessage
			// to complete first (via assistantPersisted handler).
			if len(m.toolResults) >= m.pendingTools {
				m.scope.Info("all tools already completed")
				m.state = StateComplete
				return tea.Batch(
					func() tea.Msg {
						return msgs.StreamCompleted{
							TurnID:        m.userMessage.ID(),
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
			m.state = StateAwaitingToolResults
		} else {
			m.state = StateComplete
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
		ctx := context.Background()

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
		return assistantPersisted{messageID: msgID}
	}
}

// fireToolResults fires ToolResultsReady for the round to handle.
func (m *Model) fireToolResults() tea.Cmd {
	turnID := m.userMessage.ID()
	results := m.toolResults
	return func() tea.Msg {
		return msgs.ToolResultsReady{
			TurnID:  turnID,
			Results: results,
		}
	}
}

// countToolUseBlocks counts the number of tool_use blocks in content.
func countToolUseBlocks(content []domain.Block) int {
	count := 0
	for _, b := range content {
		if b.Type == domain.BlockTypeToolUse {
			count++
		}
	}
	return count
}
