package chat

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/commandbar"
	"github.com/usetero/cli/internal/tui/app/chat/messages"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	"github.com/usetero/cli/internal/tui/app/chat/messages/user"
	"github.com/usetero/cli/internal/tui/app/chat/sidebar"
	"github.com/usetero/cli/internal/tui/app/chat/turn"
	apptools "github.com/usetero/cli/internal/tui/app/tools"
	"github.com/usetero/cli/internal/tui/layouts/base"
	"github.com/usetero/cli/internal/tui/layouts/header"
	"github.com/usetero/cli/internal/version"
)

// State represents the chat state machine.
type State int

const (
	StateIdle State = iota
	StateStreaming
	StateAwaitingTools
)

// messagesLoadedMsg is sent when messages are loaded from SQLite.
type messagesLoadedMsg struct {
	conversationID domain.ConversationID
	messages       []domain.Message
	err            error
}

// Model is the chat page.
type Model struct {
	ctx    context.Context
	theme  *styles.Theme
	db     sqlite.DB
	turn   turn.Model
	logger log.Logger

	// Identity
	accountID   domain.AccountID
	workspaceID domain.WorkspaceID

	// Layouts
	headerLayout header.Model
	baseLayout   base.Model
	sidebar      sidebar.Model

	// Components
	commandBar commandbar.Model
	list       messages.Model

	// Tools for execution
	tools apptools.Tools

	// Conversation state
	conversationID domain.ConversationID
	rawMessages    []domain.Message

	// State machine
	state            State
	currentAssistant *assistant.Model // The assistant message being built
	pendingToolIDs   map[string]bool  // Tools waiting for results

	// Focus
	focusedComponent focusTarget

	width  int
	height int
}

// focusTarget indicates which component has keyboard focus.
type focusTarget int

const (
	focusCommandBar focusTarget = iota
	focusMessageList
)

// New creates a new chat model.
func New(ctx context.Context, theme *styles.Theme, db sqlite.DB, chatClient chat.Client, accountID domain.AccountID, workspaceID domain.WorkspaceID, tools apptools.Tools, logger log.Logger) Model {
	return Model{
		ctx:              ctx,
		theme:            theme,
		db:               db,
		turn:             turn.New(theme, chatClient, logger),
		logger:           logger,
		accountID:        accountID,
		workspaceID:      workspaceID,
		headerLayout:     header.New(theme, logger),
		baseLayout:       base.New(theme, logger),
		sidebar:          sidebar.New(theme, logger).SetVersion(version.Version),
		commandBar:       commandbar.New(theme, logger),
		list:             messages.New(theme, logger),
		focusedComponent: focusCommandBar,
		tools:            tools,
		state:            StateIdle,
		pendingToolIDs:   make(map[string]bool),
	}
}

// Init initializes the chat model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandBar.Init(),
		m.list.Init(),
	)
}

// keyBindings returns the current key bindings for the footer.
func (m Model) keyBindings() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
		key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}

	if m.focusedComponent == focusCommandBar {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
			key.NewBinding(key.WithKeys("shift+enter"), key.WithHelp("shift+enter", "newline")),
		)
	} else {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "scroll")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
		)
	}

	return bindings
}

// Update handles messages.
// Rule: return early ONLY if this model is the sole consumer of the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle messages we care about
	switch msg := msg.(type) {
	case commandbar.SubmitMsg:
		return m.handleSubmit(msg.Text) // sole consumer
	case tools.ResultMsg:
		return m.handleToolResult(msg) // sole consumer
	case messagesLoadedMsg:
		return m.handleMessagesLoaded(msg) // sole consumer
	case turn.StreamDoneMsg:
		m, cmd = m.handleStreamDone(msg) // turn also needs this
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateLayout() // children also need this
	case tea.KeyPressMsg:
		if msg.String() == "tab" {
			m = m.switchFocus()
			return m, nil // sole consumer
		}
	}

	// Forward to children - they decide what to handle

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Keypresses only go to focused component
		if m.focusedComponent == focusMessageList {
			m.list, cmd = m.list.Update(keyMsg)
		} else {
			m.commandBar, cmd = m.commandBar.Update(keyMsg)
		}
		cmds = append(cmds, cmd)
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		m.commandBar, cmd = m.commandBar.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.turn, cmd = m.turn.Update(msg)
	cmds = append(cmds, cmd)

	m.baseLayout, cmd = m.baseLayout.Update(msg)
	cmds = append(cmds, cmd)

	m.headerLayout, cmd = m.headerLayout.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// switchFocus toggles between command bar and message list.
func (m Model) switchFocus() Model {
	if m.focusedComponent == focusCommandBar {
		m.focusedComponent = focusMessageList
		m.list = m.list.Focus()
	} else {
		m.focusedComponent = focusCommandBar
		m.list = m.list.Blur()
	}
	m.headerLayout = m.headerLayout.SetKeyBindings(m.keyBindings())
	m.baseLayout = m.baseLayout.SetKeyBindings(m.keyBindings())
	return m
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
		m.logger.Info("conversation created", "id", convID)
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

	// Append to state and list
	wasEmpty := !m.hasMessages()
	m.rawMessages = append(m.rawMessages, userMsg)
	m.list = m.list.AppendItem(user.New(m.theme, m.logger, userMsg))
	m.list = m.list.ScrollToBottom()

	if wasEmpty {
		m = m.updateLayout()
	}

	// Start streaming with thinking indicator
	m.state = StateStreaming
	var thinkingCmd tea.Cmd
	m.list, thinkingCmd = m.list.StartThinking()
	m.logger.Info("message sent")

	return m, tea.Batch(thinkingCmd, m.turn.Send(ctx, convID.String(), m.rawMessages, m.tools.Definitions()))
}

// handleStreamDone processes when streaming completes.
func (m Model) handleStreamDone(msg turn.StreamDoneMsg) (Model, tea.Cmd) {
	// Stop thinking indicator
	m.list = m.list.StopThinking()

	if msg.Err != nil {
		m.logger.Error("stream error", "error", msg.Err)
		m.state = StateIdle
		return m, nil
	}

	m.logger.Info("response received", "stop_reason", msg.StopReason)

	// Create assistant model with the message
	assistantModel := assistant.New(m.theme, m.logger, msg.Message, m.tools)
	m.currentAssistant = assistantModel

	// Add to list
	m.list = m.list.AppendItem(assistantModel)
	m.list = m.list.ScrollToBottom()

	// Append to raw messages
	m.rawMessages = append(m.rawMessages, msg.Message)

	// Check if we need to execute tools
	if msg.StopReason == "tool_use" && assistantModel.HasTools() {
		m.state = StateAwaitingTools
		m.pendingToolIDs = make(map[string]bool)

		// Track pending tools
		for _, tool := range assistantModel.Tools() {
			m.pendingToolIDs[tool.ID()] = true
		}

		m.logger.Info("executing tools", "count", len(m.pendingToolIDs))

		// Init triggers tool execution
		return m, assistantModel.Init()
	}

	// Done
	m.state = StateIdle
	m.currentAssistant = nil
	return m, nil
}

// handleToolResult processes when a tool completes.
func (m Model) handleToolResult(msg tools.ResultMsg) (Model, tea.Cmd) {
	m.logger.Info("tool completed", "id", msg.ToolUseID, "error", msg.Result.IsError)

	// Forward to current assistant to update tool state
	if m.currentAssistant != nil {
		updated, cmd := m.currentAssistant.Update(msg)
		if a, ok := updated.(*assistant.Model); ok {
			m.currentAssistant = a
			// Update in list
			m.list = m.list.UpdateItem(a.ID(), a)
		}

		// Mark tool as complete
		delete(m.pendingToolIDs, msg.ToolUseID)

		m.logger.Debug("tools remaining", "remaining", len(m.pendingToolIDs))

		// Check if all tools are done
		if len(m.pendingToolIDs) == 0 {
			return m.continueAfterTools()
		}

		if cmd != nil {
			return m, cmd
		}
	}

	return m, nil
}

// handleMessagesLoaded processes loaded messages from SQLite.
func (m Model) handleMessagesLoaded(msg messagesLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.logger.Error("failed to load messages", "error", msg.err)
		return m, nil
	}
	if msg.conversationID == m.conversationID {
		m.rawMessages = msg.messages
		m = m.rebuildList()
		m.logger.Info("conversation loaded", "id", msg.conversationID, "messages", len(msg.messages))
	}
	return m, nil
}

// continueAfterTools sends tool results back to API and continues.
func (m Model) continueAfterTools() (Model, tea.Cmd) {
	if m.currentAssistant == nil {
		m.state = StateIdle
		return m, nil
	}

	// Collect tool results
	toolResults := m.currentAssistant.ToolResults()
	if len(toolResults) == 0 {
		m.state = StateIdle
		m.currentAssistant = nil
		return m, nil
	}

	// Create tool result message
	toolResultMsg := domain.Message{
		ID:             domain.NewMessageID(),
		ConversationID: m.conversationID,
		Role:           domain.RoleUser,
		Content:        toolResults,
	}
	m.rawMessages = append(m.rawMessages, toolResultMsg)

	// Continue streaming with thinking indicator
	m.state = StateStreaming
	m.currentAssistant = nil
	var thinkingCmd tea.Cmd
	m.list, thinkingCmd = m.list.StartThinking()

	m.logger.Info("continuing with tool results", "count", len(toolResults))

	return m, tea.Batch(thinkingCmd, m.turn.Send(m.ctx, m.conversationID.String(), m.rawMessages, m.tools.Definitions()))
}

// rebuildList rebuilds the message list from raw messages.
func (m Model) rebuildList() Model {
	var items []messages.Item
	for _, msg := range m.rawMessages {
		items = append(items, m.messageToItems(msg)...)
	}

	m.list = m.list.SetItems(items)
	m.list = m.list.ScrollToBottom()
	return m
}

// messageToItems converts a domain.Message to list items.
func (m Model) messageToItems(msg domain.Message) []messages.Item {
	var items []messages.Item

	switch msg.Role {
	case domain.RoleUser:
		// Skip user messages that only contain tool_result blocks
		hasNonToolResult := false
		for _, block := range msg.Content {
			if block.Type != domain.BlockTypeToolResult {
				hasNonToolResult = true
				break
			}
		}
		if hasNonToolResult {
			items = append(items, user.New(m.theme, m.logger, msg))
		}

	case domain.RoleAssistant:
		items = append(items, assistant.New(m.theme, m.logger, msg, m.tools))
	}

	return items
}

// hasMessages returns true if there are messages to display.
func (m Model) hasMessages() bool {
	return m.list.Len() > 0 || m.state == StateStreaming
}

// View renders the chat.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Empty state
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

	// Has messages
	commandBarView := m.commandBar.View()
	mainContent := lipgloss.JoinVertical(lipgloss.Left, m.list.View(), commandBarView)

	sidebarView := m.sidebar.View()
	composedContent := lipgloss.JoinHorizontal(lipgloss.Top, mainContent, " ", sidebarView)

	return m.baseLayout.Render(composedContent)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m.updateLayout()
}

const sidebarWidth = 30

// updateLayout updates all component sizes.
func (m Model) updateLayout() Model {
	m.headerLayout = m.headerLayout.SetSize(m.width, m.height)
	m.headerLayout = m.headerLayout.SetKeyBindings(m.keyBindings())

	m.baseLayout = m.baseLayout.SetSize(m.width, m.height)
	m.baseLayout = m.baseLayout.SetKeyBindings(m.keyBindings())

	var contentWidth, contentHeight int
	if m.hasMessages() {
		baseWidth, baseHeight := m.baseLayout.ContentSize()
		contentWidth = baseWidth - sidebarWidth - 1
		contentHeight = baseHeight
		m.sidebar = m.sidebar.SetSize(sidebarWidth, contentHeight)
	} else {
		contentWidth, contentHeight = m.headerLayout.ContentSize()
	}

	commandBarHeight := m.commandBar.Height()
	listHeight := contentHeight - commandBarHeight

	m.commandBar = m.commandBar.SetWidth(contentWidth)
	m.list = m.list.SetSize(contentWidth, listHeight)

	return m
}

// SetConversation loads a conversation from SQLite.
func (m Model) SetConversation(conversationID domain.ConversationID) (Model, tea.Cmd) {
	m.conversationID = conversationID
	m.rawMessages = nil
	m.state = StateIdle
	m.currentAssistant = nil
	m.pendingToolIDs = make(map[string]bool)
	m.list = m.list.SetItems(nil)

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

// IsStreaming returns true if currently receiving a response.
func (m Model) IsStreaming() bool {
	return m.state == StateStreaming
}
