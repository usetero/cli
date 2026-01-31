package chat

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/commandbar"
)

// ConversationSelectedMsg is sent when a conversation is selected.
type ConversationSelectedMsg struct {
	ConversationID string
}

// messageSentMsg is sent when a message has been written to SQLite.
type messageSentMsg struct {
	conversationID string
	err            error
}

// Chat is the main chat page - the "canvas" everything builds on.
// It displays the message list and owns its input.
type Chat struct {
	theme   *styles.Theme
	logger  log.Logger
	service *chat.Service

	// Identity
	orgID     string
	accountID string

	// Current conversation
	conversationID string

	// Components
	messages *MessageList
	input    *commandbar.CommandBar

	// State
	width  int
	height int
	ready  bool
}

// New creates a new chat page.
func New(theme *styles.Theme, db sqlite.Database, service *chat.Service, orgID string, accountID string, logger log.Logger) *Chat {
	c := &Chat{
		theme:     theme,
		service:   service,
		orgID:     orgID,
		accountID: accountID,
		logger:    logger,
		messages:  NewMessageList(theme, db),
		input:     commandbar.New(theme, logger),
	}
	return c
}

// Init initializes the chat page.
func (c *Chat) Init() tea.Cmd {
	return tea.Batch(
		c.messages.Init(),
		c.input.Init(),
	)
}

// Update handles messages.
func (c *Chat) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ConversationSelectedMsg:
		c.conversationID = msg.ConversationID
		cmds = append(cmds, c.messages.SetConversation(msg.ConversationID))
		return tea.Batch(cmds...)

	case commandbar.SubmitMsg:
		// User submitted a message - write to SQLite
		return c.sendMessage(msg.Text)

	case messageSentMsg:
		if msg.err != nil {
			c.logger.Error("failed to send message", "error", msg.err)
			return nil
		}
		// Update conversation ID if a new one was created
		if c.conversationID == "" && msg.conversationID != "" {
			c.conversationID = msg.conversationID
			return c.messages.SetConversation(msg.conversationID)
		}
		return nil

	case tea.KeyPressMsg:
		// Route to input
		cmd := c.input.Update(msg)
		cmds = append(cmds, cmd)
		return tea.Batch(cmds...)
	}

	// Forward other messages to input and message list
	cmd := c.input.Update(msg)
	cmds = append(cmds, cmd)

	cmd = c.messages.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// View renders the chat page (messages + input).
func (c *Chat) View() string {
	if !c.ready {
		return ""
	}

	// Render messages
	messagesView := c.messages.View()

	// Render input
	inputView := c.input.View()

	// Compose vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		messagesView,
		inputView,
	)
}

// SetSize sets the dimensions available for content.
func (c *Chat) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.ready = true

	// Input takes fixed height, messages get the rest
	inputHeight := c.input.Height()
	messagesHeight := height - inputHeight
	if messagesHeight < 1 {
		messagesHeight = 1
	}

	c.messages.SetSize(width, messagesHeight)
	c.input.SetSize(width)
}

// Title returns the page title.
func (c *Chat) Title() string {
	return "Chat"
}

// Metadata returns context to display in sidebar/header.
func (c *Chat) Metadata() []page.Metadata {
	return []page.Metadata{
		{Label: "Organization", Value: c.orgID, Priority: 1},
		{Label: "Account", Value: c.accountID, Priority: 2},
	}
}

// AcceptsNaturalLanguage returns true - chat accepts free-form input.
func (c *Chat) AcceptsNaturalLanguage() bool {
	return true
}

// Commands returns available slash commands.
func (c *Chat) Commands() []page.Command {
	return []page.Command{
		{Name: "services", Description: "View and manage services"},
		{Name: "policies", Description: "View and manage policies"},
		{Name: "help", Description: "Show available commands"},
	}
}

// KeyBindings returns keyboard shortcuts for the footer.
func (c *Chat) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear"),
		),
	}
}

// sendMessage writes a user message to SQLite via the chat service.
func (c *Chat) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		convID, err := c.service.SendMessage(context.Background(), c.accountID, c.conversationID, text)
		return messageSentMsg{conversationID: convID, err: err}
	}
}

// IsBusy returns true if chat is loading or streaming.
func (c *Chat) IsBusy() bool {
	return c.messages.IsBusy()
}

// HasError returns true if chat is in an error state.
func (c *Chat) HasError() bool {
	return c.messages.HasError()
}

// Error returns the current error.
func (c *Chat) Error() error {
	return c.messages.Error()
}
