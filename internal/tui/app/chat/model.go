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
	"github.com/usetero/cli/internal/tui/app/tools"
	"github.com/usetero/cli/internal/tui/app/tools/query"
	"github.com/usetero/cli/internal/tui/components/thinking"
)

// turnEventMsg wraps a TurnEvent for the Bubble Tea message loop.
type turnEventMsg struct {
	event TurnEvent
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
	db     sqlite.DB
	turn   Turn
	logger log.Logger

	// Identity
	accountID   domain.AccountID
	workspaceID domain.WorkspaceID

	// Tools (global + view-specific, merged on construction)
	tools tools.Tools

	// Conversation state
	conversationID domain.ConversationID
	messages       []domain.Message

	// Streaming state
	streaming *chat.Accumulator
	thinking  thinking.Model
	eventCh   chan TurnEvent

	width  int
	height int
}

// New creates a new chat model.
func New(ctx context.Context, theme *styles.Theme, db sqlite.DB, turn Turn, accountID domain.AccountID, workspaceID domain.WorkspaceID, globalTools tools.Tools, logger log.Logger) Model {
	m := Model{
		ctx:         ctx,
		theme:       theme,
		db:          db,
		turn:        turn,
		logger:      logger,
		accountID:   accountID,
		workspaceID: workspaceID,
	}
	m.tools = globalTools.Merge(m.viewTools())
	return m
}

// viewTools returns the tools specific to the chat view.
func (m Model) viewTools() tools.Tools {
	return tools.Tools{
		query.Tool{DB: m.db},
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
	case thinking.TickMsg:
		// Forward ticks while streaming with no content yet
		if m.streaming != nil && len(m.streaming.Blocks()) == 0 {
			var cmd tea.Cmd
			m.thinking, cmd = m.thinking.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case messages.SubmitMsg:
		return m.handleSubmit(msg.Text)

	case turnEventMsg:
		return m.handleTurnEvent(msg.event)

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

// handleTurnEvent processes events from the Turn.
func (m Model) handleTurnEvent(event TurnEvent) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle errors
	if event.Error != nil {
		m.logger.Error("turn error", "error", event.Error)
		m.streaming = nil
		m.eventCh = nil
		return m, nil
	}

	// Handle streaming events (update accumulator for live display)
	if event.Event != nil {
		m.streaming.Handle(*event.Event)
	}

	// Handle completed assistant messages (update conversation state)
	if event.AssistantMessage != nil {
		m.messages = append(m.messages, *event.AssistantMessage)
		// Reset accumulator for potential next round (tool use)
		m.streaming = chat.NewAccumulator()
	}

	// Handle tool results (for display)
	if event.ToolResult != nil {
		// Tool results are added to messages by the Turn as part of the user message
		// We just log it here for debugging
		m.logger.Debug("tool result received", "tool_use_id", event.ToolResult.ToolResult.ToolUseID)
	}

	// Check if turn is complete
	if event.Done {
		m.streaming = nil
		m.eventCh = nil
		return m, nil
	}

	// Continue reading from the channel
	cmds = append(cmds, m.waitForNextEvent())
	return m, tea.Batch(cmds...)
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

	colors := m.theme.Colors
	blocks := m.streaming.Blocks()

	// Show thinking animation while waiting for content
	if len(blocks) == 0 {
		label := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientEnd).
			Bold(true).
			Render("Tero")
		return lipgloss.JoinVertical(lipgloss.Left, label, m.thinking.View())
	}

	msg := domain.Message{
		Role:    domain.RoleAssistant,
		Content: blocks,
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
	userMsg := domain.Message{
		ID:             domain.NewMessageID(),
		ConversationID: convID,
		Role:           domain.RoleUser,
		Content:        []domain.Block{domain.NewTextBlock(text)},
		CreatedAt:      time.Now(),
	}

	// Persist user message
	_, err := db.Messages().CreateUserMessage(ctx, accountID, convID, text)
	if err != nil {
		m.logger.Error("failed to persist user message", "error", err)
		return m, nil
	}

	// Append to state
	m.messages = append(m.messages, userMsg)

	// Start streaming
	m.streaming = chat.NewAccumulator()

	// Start thinking animation
	m.thinking = thinking.New(m.theme, "Thinking")

	// Create channel for turn events
	m.eventCh = make(chan TurnEvent, 100)

	// Start the turn
	go m.turn.Run(ctx, convID.String(), m.messages, m.tools, m.eventCh)

	return m, tea.Batch(m.waitForNextEvent(), m.thinking.Init())
}

// waitForNextEvent returns a command that waits for the next event from the channel.
func (m Model) waitForNextEvent() tea.Cmd {
	ch := m.eventCh
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return turnEventMsg{event: TurnEvent{Done: true}}
		}
		return turnEventMsg{event: event}
	}
}

// IsStreaming returns true if currently receiving a response.
func (m Model) IsStreaming() bool {
	return m.streaming != nil
}
