package chat

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// UserMessage displays a user's chat message.
type UserMessage struct {
	theme   *styles.Theme
	message domain.Message
	width   int
}

// NewUserMessage creates a new user message component.
func NewUserMessage(theme *styles.Theme, message domain.Message) UserMessage {
	return UserMessage{
		theme:   theme,
		message: message,
	}
}

// Init initializes the component.
func (m UserMessage) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m UserMessage) Update(msg tea.Msg) (UserMessage, tea.Cmd) {
	return m, nil
}

// View renders the message.
func (m UserMessage) View() string {
	colors := m.theme.Colors

	label := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Render("You")

	// Extract text from content blocks
	var text string
	for _, block := range m.message.Content {
		if block.Type == domain.BlockTypeText && block.Text != nil {
			text = block.Text.Content
			break
		}
	}

	content := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Width(m.width).
		Render(text)

	return lipgloss.JoinVertical(lipgloss.Left, label, content)
}

// SetWidth returns a new UserMessage with the given width.
func (m UserMessage) SetWidth(width int) UserMessage {
	m.width = width
	return m
}

// ID returns the message ID.
func (m UserMessage) ID() string {
	return m.message.ID.String()
}
