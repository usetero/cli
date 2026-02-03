package chat

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/messages"
)

// eventMsg wraps a streaming event from the Chat API.
type eventMsg struct {
	event chat.Event
}

// errorMsg is sent when the Chat API returns an error.
type errorMsg struct {
	err error
}

// persistedMsg is sent when the assistant message is persisted.
type persistedMsg struct {
	messageID domain.MessageID
	err       error
}

// messagesLoadedMsg is sent when messages are loaded from SQLite.
type messagesLoadedMsg struct {
	conversationID domain.ConversationID
	messages       []domain.Message
	err            error
}

// Model is the chat page.
// It owns the conversation state and handles chat interactions.
type Model struct {
	ctx    context.Context
	theme  *styles.Theme
	db     sqlite.Database
	client chat.Client
	logger log.Logger

	// Identity
	accountID   domain.AccountID
	workspaceID domain.WorkspaceID

	// Conversation state
	conversationID domain.ConversationID
	messages       []domain.Message

	// Streaming state
	streaming *chat.Accumulator
	eventCh   chan tea.Msg

	width  int
	height int
}

// New creates a new chat model.
func New(ctx context.Context, theme *styles.Theme, db sqlite.Database, client chat.Client, accountID domain.AccountID, workspaceID domain.WorkspaceID, logger log.Logger) Model {
	return Model{
		ctx:         ctx,
		theme:       theme,
		db:          db,
		client:      client,
		logger:      logger,
		accountID:   accountID,
		workspaceID: workspaceID,
	}
}

// Init initializes the chat model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.SubmitMsg:
		return m.handleSubmit(msg.Text)

	case eventMsg:
		m.streaming.Handle(msg.event)

		// Check if streaming is done
		if msg.event.Done {
			var assistantMsg *domain.Message
			m, assistantMsg = m.finishStreaming()
			if assistantMsg != nil {
				cmds = append(cmds, m.persistAssistantMessage(assistantMsg))
			}
			m.eventCh = nil
			return m, tea.Batch(cmds...)
		}

		// Continue reading from the channel
		if m.eventCh != nil {
			cmds = append(cmds, m.waitForNextEvent())
		}
		return m, tea.Batch(cmds...)

	case errorMsg:
		m.logger.Error("chat API error", "error", msg.err)
		m.eventCh = nil
		m.streaming = nil
		// TODO: show error to user
		return m, nil

	case persistedMsg:
		if msg.err != nil {
			m.logger.Error("failed to persist assistant message", "error", msg.err)
		} else {
			m.logger.Debug("assistant message persisted", "id", msg.messageID)
		}
		return m, nil

	case messagesLoadedMsg:
		if msg.err != nil {
			m.logger.Error("failed to load messages", "error", msg.err)
			return m, nil
		}
		if msg.conversationID == m.conversationID {
			m.messages = msg.messages
		}
		return m, nil
	}

	return m, nil
}

// View renders the chat.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Empty state
	if len(m.messages) == 0 && m.streaming == nil {
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")
	}

	// Render all messages
	var rendered []string
	for _, msg := range m.messages {
		rendered = append(rendered, m.renderMessage(msg))
	}

	// Render streaming message if present
	if m.streaming != nil {
		rendered = append(rendered, m.renderStreamingMessage())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rendered...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content)
}

// renderMessage renders a single message.
func (m Model) renderMessage(msg domain.Message) string {
	switch msg.Role {
	case domain.RoleUser:
		return NewUserMessage(m.theme, msg).SetWidth(m.width).View()
	case domain.RoleAssistant:
		return NewAssistantMessage(m.theme, msg).SetWidth(m.width).View()
	default:
		return ""
	}
}

// renderStreamingMessage renders the in-progress assistant message.
func (m Model) renderStreamingMessage() string {
	if m.streaming == nil {
		return ""
	}

	msg := domain.Message{
		Role:    domain.RoleAssistant,
		Content: m.streaming.Blocks(),
		Model:   m.streaming.Model(),
	}

	return NewAssistantMessage(m.theme, msg).SetWidth(m.width).View()
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetConversation loads a conversation from SQLite.
func (m Model) SetConversation(conversationID domain.ConversationID) (Model, tea.Cmd) {
	m.conversationID = conversationID
	m.messages = nil
	m.streaming = nil

	if conversationID == "" {
		return m, nil
	}

	db := m.db
	return m, func() tea.Msg {
		messages, err := db.Messages().List(context.Background(), conversationID)
		return messagesLoadedMsg{
			conversationID: conversationID,
			messages:       messages,
			err:            err,
		}
	}
}

// handleSubmit processes a user message submission.
func (m Model) handleSubmit(text string) (Model, tea.Cmd) {
	ctx := m.ctx
	db := m.db
	client := m.client
	accountID := m.accountID
	workspaceID := m.workspaceID
	convID := m.conversationID

	// Create conversation if needed
	if convID == "" {
		convIDStr, err := db.Conversations().Create(ctx, accountID.String(), workspaceID.String())
		if err != nil {
			m.logger.Error("failed to create conversation", "error", err)
			return m, nil
		}
		convID = domain.ConversationID(convIDStr)
		m.conversationID = convID
	}

	// Create user message
	msgID := domain.NewMessageID()
	userMsg := domain.Message{
		ID:             msgID,
		ConversationID: convID,
		Role:           domain.RoleUser,
		Content:        []domain.Block{domain.NewTextBlock(text)},
		CreatedAt:      time.Now(),
	}

	// Append to state
	m.messages = append(m.messages, userMsg)

	// Start streaming
	m.streaming = chat.NewAccumulator()

	// Create channel for events
	m.eventCh = make(chan tea.Msg, 100)

	// Start the streaming goroutine
	go m.streamChat(ctx, db, client, accountID, convID, userMsg, m.eventCh)

	return m, m.waitForNextEvent()
}

// streamChat runs in a goroutine and sends events to the channel.
func (m Model) streamChat(ctx context.Context, db sqlite.Database, client chat.Client, accountID domain.AccountID, convID domain.ConversationID, userMsg domain.Message, eventCh chan<- tea.Msg) {
	defer close(eventCh)

	// Persist user message to SQLite
	_, err := db.Messages().CreateUserMessage(ctx, accountID, convID, userMsg.Content[0].Text.Content)
	if err != nil {
		eventCh <- errorMsg{err: err}
		return
	}

	// Build request with full message history
	req := chat.Request{
		ConversationID: convID.String(),
		Messages:       m.messages,
	}

	// Stream the response
	err = client.Send(ctx, req, func(event chat.Event) error {
		eventCh <- eventMsg{event: event}
		return nil
	})

	if err != nil {
		eventCh <- errorMsg{err: err}
	}
}

// waitForNextEvent returns a command that waits for the next event from the channel.
func (m Model) waitForNextEvent() tea.Cmd {
	ch := m.eventCh
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// finishStreaming finalizes the streaming message and adds it to state.
func (m Model) finishStreaming() (Model, *domain.Message) {
	if m.streaming == nil {
		return m, nil
	}

	msg := &domain.Message{
		ConversationID: m.conversationID,
		Role:           domain.RoleAssistant,
		Content:        m.streaming.Blocks(),
		Model:          m.streaming.Model(),
		StopReason:     m.streaming.StopReason(),
	}

	m.messages = append(m.messages, *msg)
	m.streaming = nil

	return m, msg
}

// persistAssistantMessage saves the completed assistant message to SQLite.
func (m Model) persistAssistantMessage(msg *domain.Message) tea.Cmd {
	ctx := m.ctx
	db := m.db
	accountID := m.accountID
	logger := m.logger

	return func() tea.Msg {
		content, err := domain.EncodeBlocks(msg.Content)
		if err != nil {
			return persistedMsg{err: err}
		}

		msgID, err := db.Messages().CreateAssistantMessage(ctx, accountID, msg.ConversationID, msg.Model)
		if err != nil {
			return persistedMsg{err: err}
		}

		err = db.Messages().UpdateContent(ctx, msgID, content)
		if err != nil {
			return persistedMsg{err: err}
		}

		err = db.Messages().UpdateMeta(ctx, msgID, msg.Model, msg.StopReason)
		if err != nil {
			return persistedMsg{err: err}
		}

		logger.Debug("persisted assistant message", "id", msgID)
		return persistedMsg{messageID: msgID}
	}
}

// IsStreaming returns true if currently receiving a response.
func (m Model) IsStreaming() bool {
	return m.streaming != nil
}
