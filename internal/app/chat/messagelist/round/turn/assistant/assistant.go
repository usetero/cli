package assistant

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/policyapprove"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/query"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/startpolicyapproval"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model renders an assistant message and manages its content blocks.
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme        styles.Theme
	blockTheme   styles.Theme // theme with elevated bg for blocks
	scope        log.Scope
	id           domain.MessageID
	blocks       []blocks.Block
	width        int
	toolRegistry *chattools.Registry
}

// New creates a new assistant message view.
func New(theme styles.Theme, id domain.MessageID, width int, toolRegistry *chattools.Registry, scope log.Scope) *Model {
	scope = scope.Child("assistant")
	return &Model{
		theme:        theme,
		blockTheme:   theme.WithBg(theme.BgElevated),
		scope:        scope,
		id:           id,
		width:        width,
		toolRegistry: toolRegistry,
	}
}

// Update handles messages. Turn filters by TurnID before forwarding.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.AssistantContentUpdated:
		cmds = append(cmds, m.ensureBlocks(msg.Message.Content))
	case msgs.StreamCompleted:
		cmds = append(cmds, m.ensureBlocks(msg.Message.Content))
	}

	for _, b := range m.blocks {
		cmds = append(cmds, b.Update(msg))
	}

	return tea.Batch(cmds...)
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	// Blocks get the width inside the border; they handle their own internal padding.
	contentWidth := width - block.BorderWidth
	for _, b := range m.blocks {
		b.SetWidth(contentWidth)
	}
}

// AddBlock adds a block directly (for testing).
func (m *Model) AddBlock(b blocks.Block) {
	m.blocks = append(m.blocks, b)
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

// Cancel stops all in-progress tool animations.
func (m *Model) Cancel() {
	for _, b := range m.blocks {
		if t, ok := b.(*tools.Model); ok {
			t.Cancel()
		}
	}
}

// Blocks returns all visual blocks for the viewport.
func (m *Model) Blocks() []block.Block {
	var result []block.Block
	for _, b := range m.blocks {
		result = append(result, b)
	}
	return result
}

// ensureBlocks creates block models as needed. Blocks handle their own updates via messages.
// Returns a command to initialize any new tool animations.
func (m *Model) ensureBlocks(content []domain.Block) tea.Cmd {
	var cmds []tea.Cmd
	contentWidth := m.width - block.BorderWidth
	for _, b := range content {
		if m.hasBlock(b.Index) {
			continue // Block exists, handles its own updates
		}

		// Create new block with content width (minus padding)
		switch b.Type {
		case domain.BlockTypeText:
			if b.Text != nil {
				m.blocks = append(m.blocks, blocks.NewTextBlock(m.blockTheme, b.Index, b.Text.Content, contentWidth))
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
	switch {
	case m.toolRegistry.Query != nil && toolUse.Name == m.toolRegistry.Query.Name():
		child = query.New(m.blockTheme, index, toolUse.ID, width, m.toolRegistry.Query, m.scope)
	case m.toolRegistry.PolicyApprove.Name():
		child = policyapprove.New(m.blockTheme, index, toolUse.ID, width, m.toolRegistry.PolicyApprove, m.scope)
	case m.toolRegistry.StartPolicyApproval.Name():
		child = startpolicyapproval.New(m.blockTheme, index, toolUse.ID, width, m.toolRegistry.StartPolicyApproval, m.scope)
	default:
		entry, ok := m.toolRegistry.Lookup(toolUse.Name)
		if !ok {
			m.scope.Warn("unknown tool, using generic action", "name", toolUse.Name)
			entry = chattools.UnknownTool(toolUse.Name)
		}
		child = action.New(index, toolUse.ID, width, entry.Config, entry.Exec, m.scope)
	}
	return tools.New(m.blockTheme, index, toolUse.ID, width, child)
}
