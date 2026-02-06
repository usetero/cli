package blocks

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// TextBlock renders a text content block.
// It is a fixed-height component - height is determined by content.
type TextBlock struct {
	theme    *styles.Theme
	index    int
	text     string
	width    int
	rendered string // cached rendered output
}

// NewTextBlock creates a new text block with the given content.
func NewTextBlock(theme *styles.Theme, index int, text string, width int) *TextBlock {
	b := &TextBlock{
		theme: theme,
		index: index,
		text:  text,
		width: width,
	}
	b.render()
	return b
}

// Update handles messages.
func (m *TextBlock) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.AssistantContentUpdated:
		m.updateFromContent(msg.Message.Content)
	case msgs.StreamCompleted:
		m.updateFromContent(msg.Message.Content)
	}
	return nil
}

// updateFromContent finds this block's content by index and updates.
func (m *TextBlock) updateFromContent(content []domain.Block) {
	for _, b := range content {
		if b.Index == m.index && b.Type == domain.BlockTypeText && b.Text != nil {
			m.SetText(b.Text.Content)
			return
		}
	}
}

// Index returns the block index.
func (m *TextBlock) Index() int {
	return m.index
}

// View renders the text block.
func (m *TextBlock) View() string {
	return m.rendered
}

// Height returns the number of lines this block renders.
func (m *TextBlock) Height() int {
	if m.rendered == "" {
		return 0
	}
	return lipgloss.Height(m.rendered)
}

// SetText sets the text and re-renders.
func (m *TextBlock) SetText(text string) {
	if m.text == text {
		return
	}
	m.text = text
	m.render()
}

// SetWidth sets the width and re-renders.
func (m *TextBlock) SetWidth(width int) {
	if m.width == width {
		return
	}
	m.width = width
	m.render()
}

func (m *TextBlock) render() {
	if m.text == "" {
		m.rendered = ""
		return
	}
	rendered := styles.RenderMarkdown(m.theme, m.text, m.width)
	m.rendered = strings.TrimRight(rendered, "\n")
}
