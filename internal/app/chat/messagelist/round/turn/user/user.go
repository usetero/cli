package user

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// Model renders a user message.
// It is a fixed-height component - height is determined by content.
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

// borderWidth is the width consumed by the left border + padding.
const borderWidth = 2

// View renders the user message.
func (m *Model) View() string {
	// Tool result messages are not rendered visually
	if len(m.input.ToolResults) > 0 {
		return ""
	}

	colors := m.theme.Colors

	// User message has a colored left border
	style := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Width(m.width - borderWidth).
		PaddingLeft(1).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Accent)

	return style.Render(m.input.Text)
}

// Height returns the number of lines this component renders.
func (m *Model) Height() int {
	// Tool result messages have no visual representation
	if len(m.input.ToolResults) > 0 {
		return 0
	}
	return lipgloss.Height(m.View())
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// ID returns the message ID.
func (m *Model) ID() domain.MessageID {
	return m.id
}
