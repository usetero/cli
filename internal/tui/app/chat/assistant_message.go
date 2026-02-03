package chat

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// AssistantMessage displays an assistant's chat message.
type AssistantMessage struct {
	theme      *styles.Theme
	message    domain.Message
	toolBlocks []ToolBlock
	width      int
}

// NewAssistantMessage creates a new assistant message component.
func NewAssistantMessage(theme *styles.Theme, message domain.Message) AssistantMessage {
	m := AssistantMessage{
		theme:   theme,
		message: message,
	}

	// Build tool blocks from content
	toolResults := make(map[string]*domain.Block)
	for i := range message.Content {
		block := &message.Content[i]
		if block.Type == domain.BlockTypeToolResult && block.ToolResult != nil {
			toolResults[block.ToolResult.ToolUseID] = block
		}
	}

	for _, block := range message.Content {
		if block.Type == domain.BlockTypeToolUse && block.ToolUse != nil {
			tb := NewToolBlock(theme, block.ToolUse)
			if result, ok := toolResults[block.ToolUse.ID]; ok {
				tb = tb.SetResult(result.ToolResult)
			}
			m.toolBlocks = append(m.toolBlocks, tb)
		}
	}

	return m
}

// Init initializes the component.
func (m AssistantMessage) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m AssistantMessage) Update(msg tea.Msg) (AssistantMessage, tea.Cmd) {
	var cmds []tea.Cmd
	for i, tb := range m.toolBlocks {
		var cmd tea.Cmd
		m.toolBlocks[i], cmd = tb.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

// View renders the message.
func (m AssistantMessage) View() string {
	colors := m.theme.Colors

	label := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientEnd).
		Bold(true).
		Render("Tero")

	var parts []string

	// Extract text content
	var texts []string
	for _, block := range m.message.Content {
		if block.Type == domain.BlockTypeText && block.Text != nil {
			texts = append(texts, block.Text.Content)
		}
	}

	if len(texts) > 0 {
		text := lipgloss.NewStyle().
			Foreground(colors.Page.Text).
			Width(m.width).
			Render(strings.Join(texts, "\n"))
		parts = append(parts, text)
	}

	// Render tool blocks
	for _, tb := range m.toolBlocks {
		parts = append(parts, tb.View())
	}

	if len(parts) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render("...")
		parts = append(parts, placeholder)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.JoinVertical(lipgloss.Left, label, content)
}

// SetWidth returns a new AssistantMessage with the given width.
func (m AssistantMessage) SetWidth(width int) AssistantMessage {
	m.width = width
	for i, tb := range m.toolBlocks {
		m.toolBlocks[i] = tb.SetWidth(width)
	}
	return m
}

// ID returns the message ID.
func (m AssistantMessage) ID() string {
	return m.message.ID.String()
}
