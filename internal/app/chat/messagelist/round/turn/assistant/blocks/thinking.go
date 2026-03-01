package blocks

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// Body padding matches the tool block.
const (
	thinkingBodyPaddingLeft  = 2
	thinkingBodyPaddingRight = 1
	thinkingBodyPaddingH     = thinkingBodyPaddingLeft + thinkingBodyPaddingRight
)

// ThinkingBlock renders a thinking content block.
// Can be expanded or collapsed. It is a fixed-height component.
// Implements block.Block.
type ThinkingBlock struct {
	theme    styles.Theme
	index    int
	text     string
	expanded bool
	width    int
	focused  bool
}

// NewThinkingBlock creates a new thinking block with the given content.
func NewThinkingBlock(theme styles.Theme, index int, text string, width int) *ThinkingBlock {
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
	colors := m.theme
	mutedStyle := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Bg)

	chevron := mutedStyle.Render("▶")
	if m.expanded {
		chevron = mutedStyle.Render("▼")
	}

	header := fmt.Sprintf("%s %s", chevron, nameStyle.Render("Thinking"))

	var content string
	if !m.expanded {
		content = header
	} else {
		// Render body with markdown styling, wrapped to available width
		bodyWidth := m.width - thinkingBodyPaddingH
		if bodyWidth < 1 {
			bodyWidth = 1
		}
		rendered := styles.RenderMarkdown(m.theme, m.text, bodyWidth)
		rendered = strings.TrimRight(rendered, "\n")

		body := lipgloss.NewStyle().
			Padding(1, thinkingBodyPaddingRight, 1, thinkingBodyPaddingLeft).
			Render(rendered)

		content = header + "\n\n" + body
	}

	return lipgloss.NewStyle().
		Background(colors.Bg).
		Padding(0, block.PaddingX).
		Width(m.width).
		Render(content)
}

// Height returns the number of lines this block renders.
func (m *ThinkingBlock) Height() int {
	return lipgloss.Height(m.View())
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
// Only toggles when clicking the header line (y == 0, no top padding).
func (m *ThinkingBlock) Toggle(y int) {
	if m.expanded && y != 0 {
		return
	}
	m.expanded = !m.expanded
}

// Kind implements block.Block.
func (m *ThinkingBlock) Kind() block.Kind {
	return block.KindThinking
}

// SetFocused implements block.Block.
func (m *ThinkingBlock) SetFocused(focused bool) {
	m.focused = focused
}

// Focused implements block.Block.
func (m *ThinkingBlock) Focused() bool {
	return m.focused
}
