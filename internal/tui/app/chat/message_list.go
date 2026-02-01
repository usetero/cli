package chat

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

// MessageList displays a scrollable list of chat items.
// Purely presentation - delegates data management to Messages.
type MessageList struct {
	theme    *styles.Theme
	messages Messages

	// Viewport state
	width    int
	height   int
	offset   int // Scroll offset (lines from bottom)
	selected int // Selected item index for keyboard nav

	// Focus state
	focused bool
}

// NewMessageList creates a new message list component.
func NewMessageList(theme *styles.Theme, messages Messages) *MessageList {
	return &MessageList{
		theme:    theme,
		messages: messages,
		selected: -1,
	}
}

// Init initializes the component.
func (m *MessageList) Init() tea.Cmd {
	return m.messages.Init()
}

// Update handles messages.
func (m *MessageList) Update(msg tea.Msg) tea.Cmd {
	// Let messages handle data-related updates
	if cmd := m.messages.Update(msg); cmd != nil {
		return cmd
	}

	// Handle presentation updates
	switch msg := msg.(type) {
	case messagesLoadedMsg:
		m.offset = 0 // Auto-scroll to bottom on load

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
	if m.messages.HasError() {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.messages.Error().Error())
	}

	items := m.messages.Items()

	// Empty state
	if len(items) == 0 {
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Start a conversation...")
	}

	// Render items
	var rendered []string
	for _, item := range items {
		rendered = append(rendered, item.View())
	}

	// Join with spacing
	content := lipgloss.JoinVertical(lipgloss.Left, rendered...)

	// Apply viewport scrolling
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(0, 1)

	// Add focus indicator
	if m.focused {
		style = style.BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(colors.Accent).
			Width(m.width - 1)
	}

	return style.Render(content)
}

// SetSize sets the dimensions.
func (m *MessageList) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.messages.SetWidth(width - 2)
}

// SetConversation sets the current conversation.
func (m *MessageList) SetConversation(conversationID string) tea.Cmd {
	m.offset = 0
	m.selected = -1
	return m.messages.SetConversation(conversationID)
}

// Refresh reloads messages.
func (m *MessageList) Refresh() tea.Cmd {
	return m.messages.Refresh()
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

// IsBusy returns true if any item is spinning.
func (m *MessageList) IsBusy() bool {
	return m.messages.IsBusy()
}

// HasError returns true if there's an error.
func (m *MessageList) HasError() bool {
	return m.messages.HasError()
}

// Error returns the current error.
func (m *MessageList) Error() error {
	return m.messages.Error()
}

// Close releases resources.
func (m *MessageList) Close() error {
	return m.messages.Close()
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
	m.offset = len(m.messages.Items()) * 10
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
