package assistant

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/query"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// paddingWidth is the width consumed by the left padding (aligns with user border).
const paddingWidth = 2

// gapBetweenBlocks is the number of blank lines between sibling blocks.
const gapBetweenBlocks = 1

// Model renders an assistant message and manages its content blocks.
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme        *styles.Theme
	scope        log.Scope
	id           domain.MessageID
	blocks       []blocks.Block
	width        int
	focused      bool
	toolRegistry *chattools.Registry
	thinking     *thinking.Model
	streaming    bool
}

// New creates a new assistant message view.
func New(theme *styles.Theme, id domain.MessageID, width int, toolRegistry *chattools.Registry, scope log.Scope) *Model {
	scope = scope.Child("assistant")
	return &Model{
		theme:        theme,
		scope:        scope,
		id:           id,
		width:        width,
		toolRegistry: toolRegistry,
		thinking:     thinking.New(theme, thinking.Settings{Label: "Thinking"}),
		streaming:    true,
	}
}

// Init starts the thinking animation.
func (m *Model) Init() tea.Cmd {
	return m.thinking.Init()
}

// Update handles messages. Turn filters by TurnID before forwarding.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.AssistantContentUpdated:
		cmds = append(cmds, m.ensureBlocks(msg.Message.Content))
	case msgs.StreamCompleted:
		cmds = append(cmds, m.ensureBlocks(msg.Message.Content))
		m.streaming = false
	}

	// Update thinking animation while streaming
	if m.streaming {
		cmds = append(cmds, m.thinking.Update(msg))
	}

	for _, b := range m.blocks {
		cmds = append(cmds, b.Update(msg))
	}

	return tea.Batch(cmds...)
}

// View renders the assistant message.
func (m *Model) View() string {
	var parts []string
	for _, b := range m.blocks {
		parts = append(parts, b.View())
	}

	// Show thinking indicator while streaming
	if m.streaming {
		parts = append(parts, m.thinking.View())
	}

	// Parent (assistant) adds gaps between children (blocks)
	var contentParts []string
	for i, p := range parts {
		if i > 0 {
			// Add blank lines between blocks
			for j := 0; j < gapBetweenBlocks; j++ {
				contentParts = append(contentParts, "")
			}
		}
		contentParts = append(contentParts, p)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	// Assistant message is indented to align with user message border
	// When focused, it gets a visible left border
	if m.focused {
		colors := m.theme.Colors
		style := lipgloss.NewStyle().
			Width(m.width - paddingWidth).
			PaddingLeft(1).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(colors.Success.Fg)
		return style.Render(content)
	}

	// Not focused: just padding to align with user message (border + padding = 2)
	return lipgloss.NewStyle().
		Width(m.width - paddingWidth).
		PaddingLeft(paddingWidth).
		Render(content)
}

// Height returns the number of lines this component renders.
func (m *Model) Height() int {
	return lipgloss.Height(m.View())
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	// Blocks get the content width (minus padding)
	contentWidth := width - paddingWidth
	for _, b := range m.blocks {
		b.SetWidth(contentWidth)
	}
}

// SetContent populates blocks from content. Used for initial population to avoid empty render.
func (m *Model) SetContent(content []domain.Block) {
	m.ensureBlocks(content)
}

// ID returns the message ID.
func (m *Model) ID() domain.MessageID {
	return m.id
}

// SetID sets the message ID.
func (m *Model) SetID(id domain.MessageID) {
	m.id = id
}

// ensureBlocks creates block models as needed. Blocks handle their own updates via messages.
// Returns a command to initialize any new tool animations.
func (m *Model) ensureBlocks(content []domain.Block) tea.Cmd {
	var cmds []tea.Cmd
	contentWidth := m.width - paddingWidth
	for _, b := range content {
		if m.hasBlock(b.Index) {
			continue // Block exists, handles its own updates
		}

		// Create new block with content width (minus padding)
		switch b.Type {
		case domain.BlockTypeText:
			if b.Text != nil {
				m.blocks = append(m.blocks, blocks.NewTextBlock(m.theme, b.Index, b.Text.Content, contentWidth))
			}
		case domain.BlockTypeThinking:
			if b.Thinking != nil {
				m.blocks = append(m.blocks, blocks.NewThinkingBlock(m.theme, b.Index, b.Thinking.Content, contentWidth))
			}
		case domain.BlockTypeToolUse:
			if b.ToolUse != nil {
				tool := m.newToolBlock(b.Index, b.ToolUse, contentWidth)
				m.blocks = append(m.blocks, tool)
				cmds = append(cmds, tool.Init())
			}
		case domain.BlockTypeToolResult:
			// Tool results are handled separately, not rendered as blocks
		}
	}
	return tea.Batch(cmds...)
}

// hasBlock checks if a block with the given index already exists.
func (m *Model) hasBlock(index int) bool {
	for _, b := range m.blocks {
		if b.Index() == index {
			return true
		}
	}
	return false
}

// newToolBlock creates the appropriate tool model wrapped in chrome.
func (m *Model) newToolBlock(index int, toolUse *domain.ToolUse, width int) *tools.Model {
	var child tools.Child
	switch toolUse.Name {
	case m.toolRegistry.Query.Name():
		child = query.New(m.theme, index, toolUse.ID, width, m.toolRegistry.Query, m.scope)
	default:
		child = query.New(m.theme, index, toolUse.ID, width, nil, m.scope)
	}
	return tools.New(m.theme, index, toolUse.ID, width, child)
}

// SetFocused sets the focused state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}
