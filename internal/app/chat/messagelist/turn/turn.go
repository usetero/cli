package turn

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/user"
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
type Model struct {
	theme  *styles.Theme
	logger log.Logger

	conversationID domain.ConversationID
	accountID      domain.AccountID

	userMessage      *user.Model
	assistantMessage *assistant.Model

	state   State
	width   int
	stream  *streamState
	initCmd tea.Cmd // Command to start thinking animation

	// Tool result collection
	pendingTools int
	toolResults  []tools.Result

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// streamState holds the channel for receiving stream updates.
type streamState struct {
	updates chan streamUpdate
	done    bool
}

// streamUpdate is sent through the channel as the stream progresses.
type streamUpdate struct {
	message *domain.Message
	err     error
	done    bool
}

// streamUpdateMsg is the internal message for stream handling.
type streamUpdateMsg struct {
	turnID domain.MessageID
	update streamUpdate
}

// New creates a new turn from a user submission.
func New(
	theme *styles.Theme,
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	width int,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *chattools.Registry,
	logger log.Logger,
) *Model {
	l := logger.With("component", "turn", "user_message_id", userMessageID)
	assistantMsg := assistant.New(theme, "", width, toolRegistry, l)
	return &Model{
		theme:            theme,
		logger:           l,
		conversationID:   conversationID,
		accountID:        accountID,
		userMessage:      user.New(theme, userMessageID, input, width),
		assistantMessage: assistantMsg,
		initCmd:          assistantMsg.Init(), // Start thinking animation immediately
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

	case msgs.QueryCompleted:
		cmds = append(cmds, m.handleToolCompleted(msg.ToolUseID, tools.Result{
			ToolUseID: msg.ToolUseID,
			Query:     &msg.Result,
			Error:     errorResultFromErr(msg.Error),
		}))

	case msgs.StartJourneyCompleted:
		cmds = append(cmds, m.handleToolCompleted(msg.ToolUseID, tools.Result{
			ToolUseID:    msg.ToolUseID,
			StartJourney: &msg.Result,
			Error:        errorResultFromErr(msg.Error),
		}))

	case msgs.EndJourneyCompleted:
		cmds = append(cmds, m.handleToolCompleted(msg.ToolUseID, tools.Result{
			ToolUseID:  msg.ToolUseID,
			EndJourney: &msg.Result,
			Error:      errorResultFromErr(msg.Error),
		}))
	}

	cmds = append(cmds, m.userMessage.Update(msg))
	cmds = append(cmds, m.assistantMessage.Update(msg))

	return tea.Batch(cmds...)
}

func errorResultFromErr(err error) *tools.ErrorResult {
	if err == nil {
		return nil
	}
	return &tools.ErrorResult{Message: err.Error()}
}

func (m *Model) handleToolCompleted(toolUseID string, result tools.Result) tea.Cmd {
	// Collect results during streaming or awaiting - tools may complete before StreamCompleted
	m.toolResults = append(m.toolResults, result)
	m.logger.Info("tool completed", "tool_use_id", toolUseID, "collected", len(m.toolResults), "pending", m.pendingTools)

	// Only fire results once we're awaiting and have all of them
	if m.state != StateAwaitingToolResults {
		return nil
	}

	if len(m.toolResults) >= m.pendingTools {
		m.logger.Info("all tools completed")
		m.state = StateComplete
		return m.fireToolResults()
	}
	return nil
}

// View renders the turn.
func (m *Model) View() string {
	userView := m.userMessage.View()
	assistantView := m.assistantMessage.View()

	if userView == "" {
		return assistantView
	}
	return userView + "\n\n" + assistantView
}

// Height returns the number of lines this turn renders.
func (m *Model) Height(width int) int {
	view := m.View()
	if view == "" {
		return 0
	}
	return strings.Count(view, "\n") + 1
}

// StartStream begins streaming the assistant response.
func (m *Model) StartStream(messages []domain.Message, chatContext []domain.ContextEntity) tea.Cmd {
	m.logger.Debug("starting stream", "message_count", len(messages))
	m.state = StateStreaming

	// Capture init command (starts thinking animation)
	initCmd := m.initCmd
	m.initCmd = nil

	updates := make(chan streamUpdate, 10)
	m.stream = &streamState{updates: updates}

	go func() {
		defer close(updates)

		req := chatclient.Request{
			ConversationID: m.conversationID.String(),
			Messages:       messages,
			Context:        chatContext,
		}

		var lastMessage *domain.Message
		err := m.chatClient.Stream(context.Background(), req, func(msg *domain.Message) {
			lastMessage = msg
			updates <- streamUpdate{message: msg}
		})

		if err != nil {
			updates <- streamUpdate{err: err, done: true}
		} else {
			updates <- streamUpdate{message: lastMessage, done: true}
		}
	}()

	return tea.Batch(initCmd, m.nextStreamUpdate())
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.userMessage.SetWidth(width)
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

// handleStreamUpdate processes a stream update and fires messages.
func (m *Model) handleStreamUpdate(update streamUpdate) tea.Cmd {
	if update.err != nil {
		m.logger.Error("stream error", "error", update.err)
		m.state = StateComplete
		return nil
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
		m.logger.Info("stream completed", "stop_reason", update.message.StopReason)

		if update.message.StopReason == "tool_use" {
			m.pendingTools = countToolUseBlocks(update.message.Content)
			m.logger.Info("awaiting tool results", "pending", m.pendingTools, "already_collected", len(m.toolResults))

			// Check if tools already completed during streaming
			if len(m.toolResults) >= m.pendingTools {
				m.logger.Info("all tools already completed")
				m.state = StateComplete
				return tea.Batch(
					func() tea.Msg {
						return msgs.StreamCompleted{
							TurnID:     m.userMessage.ID(),
							Message:    *update.message,
							StopReason: update.message.StopReason,
						}
					},
					m.persistAssistantMessage(update.message),
					m.fireToolResults(),
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
					TurnID:     turnID,
					Message:    *update.message,
					StopReason: update.message.StopReason,
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
			m.logger.Error("failed to create assistant message", "error", err)
			return nil
		}

		content, err := domain.EncodeBlocks(msg.Content)
		if err != nil {
			m.logger.Error("failed to encode blocks", "error", err)
			return nil
		}

		if err := m.db.Messages().UpdateContent(ctx, msgID, content); err != nil {
			m.logger.Error("failed to update content", "error", err)
			return nil
		}

		if err := m.db.Messages().UpdateMeta(ctx, msgID, msg.Model, msg.StopReason); err != nil {
			m.logger.Error("failed to update meta", "error", err)
			return nil
		}

		m.logger.Info("assistant message persisted", "message_id", msgID)
		return msgs.AssistantMessageCreated{MessageID: msgID}
	}
}

// fireToolResults fires UserSubmittedInput with tool results.
func (m *Model) fireToolResults() tea.Cmd {
	results := m.toolResults
	return func() tea.Msg {
		return msgs.UserSubmittedInput{
			ToolResults: results,
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
