package user

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// Model renders a user message.
// It is a fixed-height component - height is determined by content.
// Implements block.Block.
type Model struct {
	theme   styles.Theme
	id      domain.MessageID
	input   msgs.UserSubmittedInput
	width   int
	focused bool
}

// New creates a new user message view.
func New(theme styles.Theme, id domain.MessageID, input msgs.UserSubmittedInput, width int) *Model {
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

// View renders the user message content without border decoration.
// The border is applied by renderBlock in the message list.
func (m *Model) View() string {
	// Tool result messages are not rendered visually
	if len(m.input.ToolResults) > 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Panel.Text).
		Background(m.theme.Colors.Panel.Bg).
		Width(m.width)

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

// Kind implements block.Block.
func (m *Model) Kind() block.Kind {
	return block.KindUser
}

// SetFocused implements block.Block.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// Focused implements block.Block.
func (m *Model) Focused() bool {
	return m.focused
}

// IsVisible returns false for tool result messages (they have no visual representation).
func (m *Model) IsVisible() bool {
	return len(m.input.ToolResults) == 0
}
