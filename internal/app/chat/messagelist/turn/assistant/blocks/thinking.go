package blocks

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// ThinkingBlock renders a thinking content block.
// Can be expanded or collapsed.
type ThinkingBlock struct {
	theme    *styles.Theme
	index    int
	text     string
	expanded bool
	width    int
}

// NewThinkingBlock creates a new thinking block with the given content.
func NewThinkingBlock(theme *styles.Theme, index int, text string, width int) *ThinkingBlock {
	return &ThinkingBlock{
		theme:    theme,
		index:    index,
		text:     text,
		expanded: false,
		width:    width,
	}
}

// Update handles messages.
func (m *ThinkingBlock) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.AssistantContentUpdated:
		m.updateFromContent(msg.Message.Content)
	case msgs.StreamCompleted:
		m.updateFromContent(msg.Message.Content)
	}
	return nil
}

// updateFromContent finds this block's content by index and updates.
func (m *ThinkingBlock) updateFromContent(content []domain.Block) {
	for _, b := range content {
		if b.Index == m.index && b.Type == domain.BlockTypeThinking && b.Thinking != nil {
			m.SetText(b.Thinking.Content)
			return
		}
	}
}

// Index returns the block index.
func (m *ThinkingBlock) Index() int {
	return m.index
}

// View renders the thinking block.
func (m *ThinkingBlock) View() string {
	if m.expanded {
		return "▼ Thinking\n" + m.text
	}
	return "▶ Thinking (collapsed)"
}

// SetText sets the text.
func (m *ThinkingBlock) SetText(text string) {
	m.text = text
}

// SetWidth sets the width.
func (m *ThinkingBlock) SetWidth(width int) {
	m.width = width
}

// SetExpanded sets the expanded state.
func (m *ThinkingBlock) SetExpanded(expanded bool) {
	m.expanded = expanded
}

// Toggle toggles the expanded state.
func (m *ThinkingBlock) Toggle() {
	m.expanded = !m.expanded
}
