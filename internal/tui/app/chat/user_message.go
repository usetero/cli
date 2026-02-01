package chat

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

var _ Item = (*UserMessage)(nil)

// UserMessage displays a user's chat message.
// Simple component: just a label and text content.
type UserMessage struct {
	theme *styles.Theme
	width int

	id      string
	content string
}

// NewUserMessage creates a new user message component.
func NewUserMessage(theme *styles.Theme, id string) *UserMessage {
	return &UserMessage{
		theme: theme,
		id:    id,
	}
}

// ID returns the message ID.
func (m *UserMessage) ID() string {
	return m.id
}

// Init initializes the component.
func (m *UserMessage) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *UserMessage) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the message.
func (m *UserMessage) View() string {
	colors := m.theme.Colors

	label := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Render("You")

	text := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Width(m.width).
		Render(m.content)

	return lipgloss.JoinVertical(lipgloss.Left, label, text)
}

// SetWidth sets the available width for rendering.
func (m *UserMessage) SetWidth(width int) {
	m.width = width
}

// SetContent updates the message content from parsed blocks.
func (m *UserMessage) SetContent(blocks []block.Block) {
	for _, b := range blocks {
		if b.Type == block.TypeText && b.Text != nil {
			m.content = b.Text.Content
			return
		}
	}
}

// Spinning returns false - user messages never spin.
func (m *UserMessage) Spinning() bool {
	return false
}
