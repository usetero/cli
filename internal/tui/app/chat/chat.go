package chat

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
)

// Chat is the main chat page - the "canvas" everything builds on.
// It displays the message list for the current conversation.
// Input handling is owned by App (command bar is app-level).
type Chat struct {
	theme  *styles.Theme
	logger log.Logger

	// Current conversation (empty until first message sent)
	conversationID string

	// Components
	messages *MessageList

	// State
	width  int
	height int
	ready  bool
}

// New creates a new chat page.
func New(theme *styles.Theme, db sqlite.Database, logger log.Logger) *Chat {
	messages := NewMessages(theme, db)
	return &Chat{
		theme:    theme,
		logger:   logger,
		messages: NewMessageList(theme, messages),
	}
}

// Init initializes the chat page.
func (c *Chat) Init() tea.Cmd {
	return c.messages.Init()
}

// Update handles messages.
func (c *Chat) Update(msg tea.Msg) tea.Cmd {
	// Forward to message list
	return c.messages.Update(msg)
}

// View renders the chat page (messages only, command bar is app-level).
func (c *Chat) View() string {
	if !c.ready {
		return ""
	}
	return c.messages.View()
}

// SetSize sets the dimensions available for content.
func (c *Chat) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.ready = true
	c.messages.SetSize(width, height)
}

// Title returns the page title.
func (c *Chat) Title() string {
	return "Chat"
}

// Metadata returns context to display in sidebar/header.
func (c *Chat) Metadata() []page.Metadata {
	// Chat doesn't have metadata of its own - context entities will go here
	return nil
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
	return nil
}

// SetConversation sets the conversation to display.
func (c *Chat) SetConversation(conversationID string) tea.Cmd {
	c.conversationID = conversationID
	return c.messages.SetConversation(conversationID)
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

// Close releases any resources held by the chat.
func (c *Chat) Close() error {
	return c.messages.Close()
}
