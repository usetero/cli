package assistant

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant/blocks"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant/blocks/tools/query"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// Model renders an assistant message and manages its content blocks.
type Model struct {
	theme        *styles.Theme
	scope        log.Scope
	id           domain.MessageID
	blocks       []blocks.Block
	width        int
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
		thinking:     thinking.New(theme, "Thinking"),
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
		m.ensureBlocks(msg.Message.Content)
	case msgs.StreamCompleted:
		m.ensureBlocks(msg.Message.Content)
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

	return strings.Join(parts, "\n")
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	for _, b := range m.blocks {
		b.SetWidth(width)
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
func (m *Model) ensureBlocks(content []domain.Block) {
	for _, b := range content {
		if m.hasBlock(b.Index) {
			continue // Block exists, handles its own updates
		}

		// Create new block
		switch b.Type {
		case domain.BlockTypeText:
			if b.Text != nil {
				m.blocks = append(m.blocks, blocks.NewTextBlock(m.theme, b.Index, b.Text.Content, m.width))
			}
		case domain.BlockTypeThinking:
			if b.Thinking != nil {
				m.blocks = append(m.blocks, blocks.NewThinkingBlock(m.theme, b.Index, b.Thinking.Content, m.width))
			}
		case domain.BlockTypeToolUse:
			if b.ToolUse != nil {
				m.blocks = append(m.blocks, m.newToolBlock(b.Index, b.ToolUse))
			}
		case domain.BlockTypeToolResult:
			// Tool results are handled separately, not rendered as blocks
		}
	}
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

// newToolBlock creates the appropriate tool model.
func (m *Model) newToolBlock(index int, toolUse *domain.ToolUse) tools.Model {
	switch toolUse.Name {
	case m.toolRegistry.Query.Name():
		return query.New(m.theme, index, toolUse.ID, m.width, m.toolRegistry.Query, m.scope)
	default:
		return query.New(m.theme, index, toolUse.ID, m.width, nil, m.scope)
	}
}
