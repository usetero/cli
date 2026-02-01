package chat

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/toolcall"
	"github.com/usetero/cli/internal/tui/components/spinner"
)

var _ Item = (*AssistantMessage)(nil)

// LoadingState represents the current loading phase.
type LoadingState int

const (
	// StateSending means we're waiting for the API request to start.
	StateSending LoadingState = iota
	// StateThinking means the API is processing, waiting for content.
	StateThinking
	// StateReady means content has arrived, no longer loading.
	StateReady
)

// AssistantMessage displays an assistant's chat message.
// Handles loading states, streaming content, thinking blocks, and tool calls.
type AssistantMessage struct {
	theme *styles.Theme
	width int

	id    string
	state LoadingState

	// Content
	textContent string
	thinking    []string
	toolCalls   []*toolcall.ToolCall

	// Loading animation
	spinner *spinner.Spinner
}

// NewAssistantMessage creates a new assistant message in sending state.
// Call SetMessageID once the actual message is created in SQLite.
func NewAssistantMessage(theme *styles.Theme) *AssistantMessage {
	m := &AssistantMessage{
		theme: theme,
		state: StateSending,
	}
	m.createSpinner("Sending")
	return m
}

// NewAssistantMessageWithID creates an assistant message with a known ID.
// Used when loading existing messages from SQLite.
func NewAssistantMessageWithID(theme *styles.Theme, id string) *AssistantMessage {
	m := &AssistantMessage{
		theme: theme,
		id:    id,
		state: StateThinking,
	}
	m.createSpinner("Thinking")
	return m
}

func (m *AssistantMessage) createSpinner(label string) {
	m.spinner = spinner.New(spinner.Settings{
		Size:        12,
		Label:       label,
		GradColorA:  m.theme.Colors.Brand.GradientStart,
		GradColorB:  m.theme.Colors.Brand.GradientEnd,
		LabelColor:  m.theme.Colors.Page.TextMuted,
		CycleColors: true,
	})
}

// ID returns the message ID.
func (m *AssistantMessage) ID() string {
	return m.id
}

// SetMessageID sets the ID once the message is created in SQLite.
// Transitions from Sending to Thinking state.
func (m *AssistantMessage) SetMessageID(id string) {
	m.id = id
	if m.state == StateSending {
		m.state = StateThinking
		m.createSpinner("Thinking")
	}
}

// Init initializes the component.
func (m *AssistantMessage) Init() tea.Cmd {
	if m.spinner != nil && m.state != StateReady {
		return m.spinner.Init()
	}
	return nil
}

// Update handles messages.
func (m *AssistantMessage) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update spinner
	if m.spinner != nil && m.state != StateReady {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update tool calls
	for _, tc := range m.toolCalls {
		if tc.Spinning() {
			if cmd := tc.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

// View renders the message.
func (m *AssistantMessage) View() string {
	colors := m.theme.Colors

	label := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientEnd).
		Bold(true).
		Render("Tero")

	// Show spinner while loading
	if m.state != StateReady {
		content := lipgloss.NewStyle().
			PaddingLeft(1).
			Render(m.spinner.View())
		return lipgloss.JoinVertical(lipgloss.Left, label, content)
	}

	// Render content
	var parts []string

	// Text content
	if m.textContent != "" {
		text := lipgloss.NewStyle().
			Foreground(colors.Page.Text).
			Width(m.width).
			Render(m.textContent)
		parts = append(parts, text)
	}

	// Thinking blocks (collapsed)
	for _, thinking := range m.thinking {
		parts = append(parts, m.viewThinking(thinking))
	}

	// Tool calls
	for _, tc := range m.toolCalls {
		parts = append(parts, tc.View())
	}

	// Placeholder if nothing to show
	if len(parts) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render("...")
		parts = append(parts, placeholder)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.JoinVertical(lipgloss.Left, label, content)
}

func (m *AssistantMessage) viewThinking(content string) string {
	colors := m.theme.Colors

	// Truncate for display
	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	return lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Italic(true).
		Width(m.width).
		Render("Thinking: " + preview)
}

// SetWidth sets the available width for rendering.
func (m *AssistantMessage) SetWidth(width int) {
	m.width = width
	for _, tc := range m.toolCalls {
		tc.SetWidth(width)
	}
}

// SetContent updates the message content from parsed blocks.
// Transitions to Ready state once text content arrives.
func (m *AssistantMessage) SetContent(blocks []block.Block) {
	var texts []string
	m.thinking = nil

	for _, b := range blocks {
		switch b.Type {
		case block.TypeText:
			if b.Text != nil && b.Text.Content != "" {
				texts = append(texts, b.Text.Content)
				m.state = StateReady
			}
		case block.TypeThinking:
			if b.Thinking != nil && b.Thinking.Content != "" {
				m.thinking = append(m.thinking, b.Thinking.Content)
			}
		case block.TypeToolUse:
			if b.ToolUse != nil {
				m.addOrUpdateToolCall(b.ToolUse, nil)
			}
		case block.TypeToolResult:
			if b.ToolResult != nil {
				m.updateToolResult(b.ToolResult)
			}
		}
	}

	m.textContent = strings.Join(texts, "\n")
}

func (m *AssistantMessage) addOrUpdateToolCall(toolUse *block.ToolUse, result *block.ToolResult) {
	// Check if we already have this tool call
	for _, tc := range m.toolCalls {
		if tc.ID() == toolUse.ID {
			if result != nil {
				tc.SetResult(result)
			}
			return
		}
	}

	// Create new tool call
	tc := toolcall.New(m.theme, toolUse)
	tc.SetWidth(m.width)
	if result != nil {
		tc.SetResult(result)
	}
	m.toolCalls = append(m.toolCalls, tc)
}

func (m *AssistantMessage) updateToolResult(result *block.ToolResult) {
	for _, tc := range m.toolCalls {
		if tc.ID() == result.ToolUseID {
			tc.SetResult(result)
			return
		}
	}
}

// Spinning returns true if showing a loading animation.
func (m *AssistantMessage) Spinning() bool {
	if m.state != StateReady {
		return true
	}
	for _, tc := range m.toolCalls {
		if tc.Spinning() {
			return true
		}
	}
	return false
}

// State returns the current loading state.
func (m *AssistantMessage) State() LoadingState {
	return m.state
}
