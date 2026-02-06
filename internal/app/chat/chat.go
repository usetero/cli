package chat

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/commandbar"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/msgs"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// Model is the main chat model.
// It is a flexible component - it renders exactly the size given by SetSize.
type Model struct {
	scope log.Scope

	commandBar  *commandbar.Model
	messageList *messagelist.Model

	// Conversation is created lazily on first message
	conversationID domain.ConversationID

	account   domain.Account
	workspace domain.Workspace
	theme     *styles.Theme
	width     int
	height    int

	// Dependencies
	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *tools.Registry
}

// New creates a new chat model.
func New(
	account domain.Account,
	workspace domain.Workspace,
	theme *styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("chat")

	return &Model{
		scope:        scope,
		commandBar:   commandbar.New(theme, scope),
		messageList:  messagelist.New(theme, db, chatClient, toolRegistry, scope),
		account:      account,
		workspace:    workspace,
		theme:        theme,
		db:           db,
		chatClient:   chatClient,
		toolRegistry: toolRegistry,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return m.commandBar.Init()
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle messages this model cares about
	switch msg := msg.(type) {
	case msgs.UserSubmittedInput:
		cmds = append(cmds, m.handleUserInput(msg))

	case conversationCreated:
		m.scope.Info("conversation created", "id", msg.conversationID)
		m.conversationID = msg.conversationID
		cmds = append(cmds, m.persistUserMessage(msg.input))

	case userMessagePersisted:
		m.scope.Debug("received userMessagePersisted", "message_id", msg.messageID)
		cmds = append(cmds, m.handlePersistedMessage(msg))
	}

	// Always forward to children
	cmds = append(cmds, m.commandBar.Update(msg))
	cmds = append(cmds, m.messageList.Update(msg))

	return tea.Batch(cmds...)
}

// SetSize updates the dimensions. This is a flexible component.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateLayout()
}

// updateLayout calculates sizes for children based on current dimensions.
func (m *Model) updateLayout() {
	// CommandBar is fixed height
	m.commandBar.SetWidth(m.width)
	commandBarHeight := m.commandBar.Height()

	// MessageList is flexible - gets remaining space
	messageListHeight := m.height - commandBarHeight
	if messageListHeight < 0 {
		messageListHeight = 0
	}
	m.messageList.SetSize(m.width, messageListHeight)
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	return m.commandBar.ShortHelp()
}

// ConversationID returns the current conversation ID.
func (m *Model) ConversationID() domain.ConversationID {
	return m.conversationID
}

// hasMessages returns true if there are messages to display.
func (m *Model) hasMessages() bool {
	return m.messageList.Len() > 0
}

// handleUserInput creates conversation if needed, then persists the user message.
func (m *Model) handleUserInput(input msgs.UserSubmittedInput) tea.Cmd {
	if len(input.Text) > 0 {
		m.scope.Info("user submitted text", "text_length", len(input.Text))
	} else {
		m.scope.Info("user submitted tool results", "count", len(input.ToolResults))
	}

	// If no conversation yet, create one first (only for text input)
	if m.conversationID == "" {
		return m.createConversation(input)
	}

	return m.persistUserMessage(input)
}

// createConversation creates a new conversation.
func (m *Model) createConversation(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		convID, err := m.db.Conversations().Create(
			ctx,
			m.account.ID.String(),
			m.workspace.ID.String(),
		)
		if err != nil {
			m.scope.Error("failed to create conversation", "error", err)
			return appmsg.Error{Message: "Failed to create conversation", Err: err}
		}

		return conversationCreated{
			conversationID: domain.ConversationID(convID),
			input:          input,
		}
	}
}

// conversationCreated is fired after conversation is created.
type conversationCreated struct {
	conversationID domain.ConversationID
	input          msgs.UserSubmittedInput
}

// persistUserMessage saves the user message and loads history for the API call.
func (m *Model) persistUserMessage(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var msgID domain.MessageID
		var err error

		if len(input.ToolResults) > 0 {
			// Convert typed results to domain format at the boundary
			domainResults := make([]domain.ToolResult, len(input.ToolResults))
			for i, r := range input.ToolResults {
				domainResults[i] = domain.ToolResult{
					ToolUseID: r.ToolUseID,
					IsError:   r.IsError(),
					Content:   r.ToMap(),
				}
				if r.Error != nil {
					domainResults[i].Error = r.Error.Message
				}
			}
			msgID, err = m.db.Messages().CreateToolResultMessage(ctx, m.account.ID, m.conversationID, domainResults)
		} else {
			msgID, err = m.db.Messages().CreateUserMessage(ctx, m.account.ID, m.conversationID, input.Text)
		}
		if err != nil {
			m.scope.Error("failed to create user message", "error", err)
			return appmsg.Error{Message: "Failed to save message", Err: err}
		}

		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.scope.Error("failed to load messages", "error", err)
			return appmsg.Error{Message: "Failed to load messages", Err: err}
		}

		return userMessagePersisted{
			conversationID: m.conversationID,
			messageID:      msgID,
			input:          input,
			messages:       messages,
		}
	}
}

// userMessagePersisted is fired after user message is saved to database.
type userMessagePersisted struct {
	conversationID domain.ConversationID
	messageID      domain.MessageID
	input          msgs.UserSubmittedInput
	messages       []domain.Message
}

// handlePersistedMessage starts the turn after the user message is persisted.
func (m *Model) handlePersistedMessage(msg userMessagePersisted) tea.Cmd {
	m.scope.Info("turn started", "conversation_id", msg.conversationID, "user_message_id", msg.messageID)

	return m.messageList.StartTurn(
		msg.conversationID,
		m.account.ID,
		msg.messageID,
		msg.input,
		msg.messages,
		nil,
	)
}

// View renders the chat. This is a flexible component - renders exactly to SetSize dimensions.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Empty state: centered prompt + command bar
	if !m.hasMessages() {
		emptyHeight := m.height - m.commandBar.Height()
		emptyView := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width).
			Height(emptyHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")

		return lipgloss.JoinVertical(lipgloss.Left, emptyView, m.commandBar.View())
	}

	// Normal state: message list + command bar
	return lipgloss.JoinVertical(lipgloss.Left, m.messageList.View(), m.commandBar.View())
}
