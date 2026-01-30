package chat

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// messagesLoadedMsg is sent when messages are loaded from SQLite.
type messagesLoadedMsg struct {
	messages []sqlite.Message
	err      error
}

// tablesChangedMsg is sent when SQLite tables change.
type tablesChangedMsg struct {
	tables []string
}

// MessageList displays a scrollable list of chat messages.
// Reads from SQLite and re-renders when data changes.
type MessageList struct {
	theme *styles.Theme
	db    sqlite.Database

	// Current conversation
	conversationID string

	// Loaded messages
	messages []sqlite.Message

	// Message renderer
	renderer *Message

	// Change subscription
	subscription *sqlite.Subscription

	// Viewport state
	width    int
	height   int
	offset   int // Scroll offset (lines from bottom)
	selected int // Selected message index for keyboard nav

	// Focus state
	focused bool

	// Error state
	err error
}

// NewMessageList creates a new message list component.
func NewMessageList(theme *styles.Theme, db sqlite.Database) *MessageList {
	return &MessageList{
		theme:    theme,
		db:       db,
		renderer: NewMessage(theme),
		selected: -1, // No selection
	}
}

// Init initializes the component and starts listening for changes.
func (m *MessageList) Init() tea.Cmd {
	m.subscription = m.db.Subscribe()
	return m.listenForChanges()
}

// listenForChanges returns a command that waits for database changes.
func (m *MessageList) listenForChanges() tea.Cmd {
	if m.subscription == nil {
		return nil
	}
	return func() tea.Msg {
		tables, ok := <-m.subscription.Changes()
		if !ok {
			return nil // Channel closed
		}
		return tablesChangedMsg{tables: tables}
	}
}

// Update handles messages.
func (m *MessageList) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case messagesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return nil
		}
		m.messages = msg.messages
		m.err = nil
		// Auto-scroll to bottom on new messages
		m.offset = 0
		return nil

	case tablesChangedMsg:
		// Check if messages table changed
		for _, table := range msg.tables {
			if table == "messages" {
				// Refresh and keep listening
				return tea.Batch(m.Refresh(), m.listenForChanges())
			}
		}
		// Keep listening even if messages didn't change
		return m.listenForChanges()

	case tea.KeyPressMsg:
		if !m.focused {
			return nil
		}
		switch msg.String() {
		case "up", "k":
			m.scrollUp()
		case "down", "j":
			m.scrollDown()
		case "home", "g":
			m.scrollToTop()
		case "end", "G":
			m.scrollToBottom()
		case "pgup":
			m.pageUp()
		case "pgdown":
			m.pageDown()
		}
	}

	return nil
}

// View renders the message list.
func (m *MessageList) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors

	// Error state
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.err.Error())
	}

	// Empty state
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")
	}

	// Render messages
	m.renderer.SetWidth(m.width - 2) // Account for padding

	var rendered []string
	for _, msg := range m.messages {
		rendered = append(rendered, m.renderer.Render(msg))
	}

	// Join with spacing
	content := lipgloss.JoinVertical(lipgloss.Left, rendered...)

	// Apply viewport scrolling
	// For now, simple overflow - TODO: proper virtual scrolling
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(0, 1)

	// Add focus indicator
	if m.focused {
		style = style.BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(colors.Accent).
			Width(m.width - 1) // Account for border
	}

	return style.Render(content)
}

// SetSize sets the dimensions.
func (m *MessageList) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetConversation sets the current conversation and loads messages.
func (m *MessageList) SetConversation(conversationID string) tea.Cmd {
	m.conversationID = conversationID
	m.messages = nil
	m.offset = 0
	m.selected = -1

	if conversationID == "" || m.db == nil {
		return nil
	}

	return m.loadMessages()
}

// Refresh reloads messages from SQLite.
func (m *MessageList) Refresh() tea.Cmd {
	if m.conversationID == "" || m.db == nil {
		return nil
	}
	return m.loadMessages()
}

// loadMessages loads messages from SQLite.
func (m *MessageList) loadMessages() tea.Cmd {
	return func() tea.Msg {
		convID := m.conversationID
		messages, err := m.db.Queries().ListMessagesByConversation(context.Background(), &convID)
		return messagesLoadedMsg{messages: messages, err: err}
	}
}

// Focus sets focus state.
func (m *MessageList) Focus() {
	m.focused = true
}

// Blur removes focus.
func (m *MessageList) Blur() {
	m.focused = false
}

// IsFocused returns focus state.
func (m *MessageList) IsFocused() bool {
	return m.focused
}

// Scrolling methods

func (m *MessageList) scrollUp() {
	m.offset++
}

func (m *MessageList) scrollDown() {
	if m.offset > 0 {
		m.offset--
	}
}

func (m *MessageList) scrollToTop() {
	// Calculate max offset based on content height
	// For now, just set a large number - will be clamped in render
	m.offset = len(m.messages) * 10
}

func (m *MessageList) scrollToBottom() {
	m.offset = 0
}

func (m *MessageList) pageUp() {
	m.offset += m.height / 2
}

func (m *MessageList) pageDown() {
	m.offset -= m.height / 2
	if m.offset < 0 {
		m.offset = 0
	}
}

// IsBusy returns false - message list is never busy.
func (m *MessageList) IsBusy() bool {
	return false
}

// HasError returns true if there's an error.
func (m *MessageList) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m *MessageList) Error() error {
	return m.err
}
