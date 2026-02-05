package chat

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/commandbar"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/app/chat/sidebar"
	"github.com/usetero/cli/internal/app/layouts/base"
	"github.com/usetero/cli/internal/app/layouts/header"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const sidebarWidth = 30

// Model is the main chat model.
type Model struct {
	logger log.Logger

	headerLayout *header.Model
	baseLayout   *base.Model
	commandBar   *commandbar.Model
	sidebar      *sidebar.Model
	messageList  *messagelist.Model

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
	width, height int,
	theme *styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	logger log.Logger,
) *Model {
	chatLogger := logger.With("component", "chat")

	headerLayout := header.New(theme)
	headerLayout.SetTitle("Chat")
	headerLayout.SetOrgName(workspace.Name)

	sidebarModel := sidebar.New(theme)
	sidebarModel.SetWorkspace(workspace.Name)

	m := &Model{
		logger:       chatLogger,
		headerLayout: headerLayout,
		baseLayout:   base.New(theme),
		commandBar:   commandbar.New(theme, width),
		sidebar:      sidebarModel,
		messageList:  messagelist.New(theme, width, height, db, chatClient, toolRegistry, chatLogger),
		account:      account,
		workspace:    workspace,
		theme:        theme,
		width:        width,
		height:       height,
		db:           db,
		chatClient:   chatClient,
		toolRegistry: toolRegistry,
	}

	m.updateLayout()
	return m
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case msgs.UserSubmittedInput:
		cmds = append(cmds, m.handleUserInput(msg))

	case conversationCreated:
		m.logger.Info("conversation created", "id", msg.conversationID)
		m.conversationID = msg.conversationID
		cmds = append(cmds, m.persistUserMessage(msg.input))

	case userMessagePersisted:
		m.logger.Debug("received userMessagePersisted", "message_id", msg.messageID)
		cmds = append(cmds, m.handlePersistedMessage(msg))
	}

	// Always forward to children
	cmds = append(cmds, m.commandBar.Update(msg))
	cmds = append(cmds, m.messageList.Update(msg))

	return tea.Batch(cmds...)
}

// updateLayout updates all component sizes based on current dimensions and state.
func (m *Model) updateLayout() {
	m.headerLayout.SetSize(m.width, m.height)
	m.baseLayout.SetSize(m.width, m.height)

	if m.hasMessages() {
		// With messages: base layout + sidebar
		contentWidth, contentHeight := m.baseLayout.ContentSize()
		mainWidth := contentWidth - sidebarWidth - 1 // -1 for gap

		m.sidebar.SetSize(sidebarWidth, contentHeight)
		m.commandBar.SetWidth(mainWidth)

		listHeight := contentHeight - m.commandBar.Height()
		m.messageList.SetSize(mainWidth, listHeight)
	} else {
		// Empty state: header layout, centered
		contentWidth, contentHeight := m.headerLayout.ContentSize()
		m.commandBar.SetWidth(contentWidth)

		listHeight := contentHeight - m.commandBar.Height()
		m.messageList.SetSize(contentWidth, listHeight)
	}
}

// hasMessages returns true if there are messages to display.
func (m *Model) hasMessages() bool {
	return m.messageList.Len() > 0
}

// handleUserInput creates conversation if needed, then persists the user message.
func (m *Model) handleUserInput(input msgs.UserSubmittedInput) tea.Cmd {
	if len(input.Text) > 0 {
		m.logger.Info("user submitted text", "text_length", len(input.Text))
	} else {
		m.logger.Info("user submitted tool results", "count", len(input.ToolResults))
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
			m.logger.Error("failed to create conversation", "error", err)
			return nil
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
			m.logger.Error("failed to create user message", "error", err)
			return nil
		}

		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.logger.Error("failed to load messages", "error", err)
			return nil
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
	wasEmpty := !m.hasMessages()

	m.logger.Info("turn started", "conversation_id", msg.conversationID, "user_message_id", msg.messageID)

	cmd := m.messageList.StartTurn(
		msg.conversationID,
		m.account.ID,
		msg.messageID,
		msg.input,
		msg.messages,
		nil,
	)

	if wasEmpty {
		m.updateLayout()
	}

	return cmd
}

// View renders the chat.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Empty state: header layout with centered prompt
	if !m.hasMessages() {
		contentWidth, contentHeight := m.headerLayout.ContentSize()

		emptyView := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(contentWidth).
			Height(contentHeight-m.commandBar.Height()).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")

		commandBarView := m.commandBar.View()
		content := lipgloss.JoinVertical(lipgloss.Left, emptyView, commandBarView)

		return m.headerLayout.Render(content)
	}

	// Has messages: base layout with sidebar on right
	commandBarView := m.commandBar.View()
	mainContent := lipgloss.JoinVertical(lipgloss.Left, m.messageList.View(), commandBarView)

	sidebarView := m.sidebar.View()
	composedContent := lipgloss.JoinHorizontal(lipgloss.Top, mainContent, " ", sidebarView)

	return m.baseLayout.Render(composedContent)
}
