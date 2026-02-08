package blocks

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// ThinkingAnimBlock wraps a thinking animation as a block.
// This is the streaming indicator shown while the assistant is generating.
type ThinkingAnimBlock struct {
	thinking *thinking.Model
	focused  bool
}

// NewThinkingAnimBlock creates a new thinking animation block.
func NewThinkingAnimBlock(t *thinking.Model) *ThinkingAnimBlock {
	return &ThinkingAnimBlock{thinking: t}
}

// View implements block.Block.
func (m *ThinkingAnimBlock) View() string {
	return lipgloss.NewStyle().
		Padding(0, block.PaddingX).
		Render(m.thinking.View())
}

// Height implements block.Block.
func (m *ThinkingAnimBlock) Height() int {
	return lipgloss.Height(m.View())
}

// Update implements block.Block.
func (m *ThinkingAnimBlock) Update(msg tea.Msg) tea.Cmd {
	return m.thinking.Update(msg)
}

// SetWidth implements block.Block.
func (m *ThinkingAnimBlock) SetWidth(_ int) {}

// SetFocused implements block.Block.
func (m *ThinkingAnimBlock) SetFocused(focused bool) {
	m.focused = focused
}

// Focused implements block.Block.
func (m *ThinkingAnimBlock) Focused() bool {
	return m.focused
}

// Kind implements block.Block.
func (m *ThinkingAnimBlock) Kind() block.Kind {
	return block.KindThinkingAnimation
}
