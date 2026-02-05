package chat

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/commandbar"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/app/chat/sidebar"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/header"
	"github.com/usetero/cli/internal/tea/keymap"
)

const (
	sidebarWidth          = 30
	sidebarMinWindowWidth = 100 // minimum window width to show sidebar
)

// Model is the main chat model.
type Model struct {
	scope log.Scope

	header      *header.Model
	commandBar  *commandbar.Model
	sidebar     *sidebar.Model
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
	width, height int,
	theme *styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("chat")

	h := header.New(theme)
	h.SetTitle("Chat")
	h.SetOrgName(workspace.Name)

	sidebarModel := sidebar.New(theme)
	sidebarModel.SetWorkspace(workspace.Name)

	m := &Model{
		scope:        scope,
		header:       h,
		commandBar:   commandbar.New(theme, width),
		sidebar:      sidebarModel,
		messageList:  messagelist.New(theme, width, height, db, chatClient, toolRegistry, scope),
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

// SetSize updates the dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateLayout()
}

// useSidebarLayout returns true when we should show sidebar instead of header.
func (m *Model) useSidebarLayout() bool {
	return m.hasMessages() && m.width >= sidebarMinWindowWidth
}

// updateLayout updates all component sizes based on current dimensions and state.
func (m *Model) updateLayout() {
	// Always update all components so they're ready when layout switches
	m.header.SetWidth(m.width)
	m.sidebar.SetSize(sidebarWidth, m.height)

	if m.useSidebarLayout() {
		mainWidth := m.width - sidebarWidth - 1 // -1 for gap
		listHeight := m.height - m.commandBar.Height()

		m.commandBar.SetWidth(mainWidth)
		m.messageList.SetSize(mainWidth, listHeight)
	} else {
		listHeight := m.height - m.header.Height() - m.commandBar.Height()

		m.commandBar.SetWidth(m.width)
		m.messageList.SetSize(m.width, listHeight)
	}
}

// KeyBindings returns the key bindings for display in footer.
func (m *Model) KeyBindings() []key.Binding {
	bindings := m.commandBar.KeyBindings()
	bindings = append(bindings, keymap.Global...)
	return bindings
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
			m.scope.Error("failed to create user message", "error", err)
			return nil
		}

		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.scope.Error("failed to load messages", "error", err)
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

	m.scope.Info("turn started", "conversation_id", msg.conversationID, "user_message_id", msg.messageID)

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

// View renders the chat content (without chrome - app handles that).
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Empty state: header + centered prompt
	if !m.hasMessages() {
		headerView := m.header.View()
		headerHeight := m.header.Height()
		contentHeight := m.height - headerHeight - m.commandBar.Height()

		emptyView := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width).
			Height(contentHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")

		commandBarView := m.commandBar.View()

		return lipgloss.JoinVertical(lipgloss.Left, headerView, emptyView, commandBarView)
	}

	// Sidebar layout: messages + command bar | sidebar
	if m.useSidebarLayout() {
		mainContent := lipgloss.JoinVertical(lipgloss.Left, m.messageList.View(), m.commandBar.View())
		return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, " ", m.sidebar.View())
	}

	// Header layout: header + messages + command bar
	return lipgloss.JoinVertical(lipgloss.Left, m.header.View(), m.messageList.View(), m.commandBar.View())
}
