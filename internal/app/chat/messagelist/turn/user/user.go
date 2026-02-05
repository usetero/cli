package user

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// Model renders a user message.
type Model struct {
	theme *styles.Theme
	id    domain.MessageID
	input msgs.UserSubmittedInput
	width int
}

// New creates a new user message view.
func New(theme *styles.Theme, id domain.MessageID, input msgs.UserSubmittedInput, width int) *Model {
	return &Model{
		theme: theme,
		id:    id,
		input: input,
		width: width,
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the user message.
func (m *Model) View() string {
	// Tool result messages are not rendered visually
	if len(m.input.ToolResults) > 0 {
		return ""
	}
	return "> " + m.input.Text
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// ID returns the message ID.
func (m *Model) ID() domain.MessageID {
	return m.id
}
