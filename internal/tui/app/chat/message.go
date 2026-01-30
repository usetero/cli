package chat

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// Message renders a single chat message.
// Stateless - just takes data and renders it.
type Message struct {
	theme *styles.Theme
	width int
}

// NewMessage creates a new message renderer.
func NewMessage(theme *styles.Theme) *Message {
	return &Message{
		theme: theme,
	}
}

// SetWidth sets the available width for rendering.
func (m *Message) SetWidth(width int) {
	m.width = width
}

// Render renders a message from the database.
func (m *Message) Render(msg sqlite.Message) string {
	if msg.Role == nil {
		return ""
	}

	switch *msg.Role {
	case "user":
		return m.renderUser(msg)
	case "assistant":
		return m.renderAssistant(msg)
	default:
		return ""
	}
}

// renderUser renders a user message.
func (m *Message) renderUser(msg sqlite.Message) string {
	colors := m.theme.Colors

	// Parse content blocks
	content := ""
	if msg.Content != nil {
		blocks, err := ParseContentBlocks(*msg.Content)
		if err == nil {
			content = m.extractText(blocks)
		} else {
			// Fallback: treat as plain text
			content = *msg.Content
		}
	}

	// User label
	label := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Render("You")

	// Message content
	text := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Width(m.width - 4). // Padding
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, label, text)
}

// renderAssistant renders an assistant message.
func (m *Message) renderAssistant(msg sqlite.Message) string {
	colors := m.theme.Colors

	// Parse content blocks
	var parts []string

	if msg.Content != nil {
		blocks, err := ParseContentBlocks(*msg.Content)
		if err == nil {
			for _, block := range blocks {
				switch block.Type {
				case BlockTypeText:
					if block.Text != nil && block.Text.Content != "" {
						text := lipgloss.NewStyle().
							Foreground(colors.Page.Text).
							Width(m.width - 4).
							Render(block.Text.Content)
						parts = append(parts, text)
					}

				case BlockTypeThinking:
					if block.Thinking != nil && block.Thinking.Content != "" {
						parts = append(parts, m.renderThinking(block.Thinking.Content))
					}

				case BlockTypeToolUse:
					if block.ToolUse != nil {
						parts = append(parts, m.renderToolUse(block.ToolUse))
					}

				case BlockTypeToolResult:
					if block.ToolResult != nil {
						parts = append(parts, m.renderToolResult(block.ToolResult))
					}
				}
			}
		}
	}

	// Assistant label
	label := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientEnd).
		Bold(true).
		Render("Tero")

	// If no content yet (streaming), show placeholder
	if len(parts) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render("...")
		parts = append(parts, placeholder)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.JoinVertical(lipgloss.Left, label, content)
}

// renderThinking renders a thinking block (collapsible).
func (m *Message) renderThinking(content string) string {
	colors := m.theme.Colors

	// Truncate long thinking for display
	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	// Simple collapsed view for now
	// TODO: Make expandable with key binding
	return lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Italic(true).
		Width(m.width - 4).
		Render(fmt.Sprintf("Thinking: %s", preview))
}

// renderToolUse renders a tool call.
func (m *Message) renderToolUse(tool *ToolUse) string {
	colors := m.theme.Colors

	// Tool name with icon
	name := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Render(fmt.Sprintf("Tool: %s", tool.Name))

	return name
}

// renderToolResult renders a tool result.
func (m *Message) renderToolResult(result *ToolResult) string {
	colors := m.theme.Colors

	if result.IsError {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Render(fmt.Sprintf("Error: %s", result.Error))
	}

	// Success indicator
	return lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		Render("Done")
}

// extractText extracts plain text from content blocks.
func (m *Message) extractText(blocks []ContentBlock) string {
	var texts []string
	for _, block := range blocks {
		if block.Type == BlockTypeText && block.Text != nil {
			texts = append(texts, block.Text.Content)
		}
	}
	return strings.Join(texts, "\n")
}
