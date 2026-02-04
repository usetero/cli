package assistant

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools/endjourney"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools/generic"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools/query"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools/startjourney"
	apptools "github.com/usetero/cli/internal/tui/app/tools"
	appendjourney "github.com/usetero/cli/internal/tui/app/tools/endjourney"
	appquery "github.com/usetero/cli/internal/tui/app/tools/query"
	appstartjourney "github.com/usetero/cli/internal/tui/app/tools/startjourney"
	"github.com/usetero/cli/internal/tui/components/highlight"
)

// Model displays an assistant message in the chat.
// It renders text content and creates tool models for tool_use blocks.
type Model struct {
	theme     *styles.Theme
	logger    log.Logger
	message   domain.Message
	executors apptools.Tools
	focused   bool

	// Highlighting support
	highlight highlight.Item

	// Child tool models - created in New()
	toolModels []*tools.Model

	cachedRender string
	cachedWidth  int
}

// Compile-time interface checks
var (
	_ messages.Item          = (*Model)(nil)
	_ messages.Copyable      = (*Model)(nil)
	_ messages.Highlightable = (*Model)(nil)
)

// New creates a new assistant message model.
// Tool models are created immediately from message content.
func New(theme *styles.Theme, logger log.Logger, message domain.Message, executors apptools.Tools) *Model {
	h := highlight.NewItem()
	h.SetHighlighter(highlight.SelectionHighlighter(theme.Colors.SelectionBg, theme.Colors.SelectionFg))
	m := &Model{
		theme:     theme,
		logger:    logger,
		message:   message,
		executors: executors,
		highlight: h,
	}
	m.toolModels = m.createToolModels()
	return m
}

// Init starts execution of all tool models.
// Returns Cmds for all tools to execute.
func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, tool := range m.toolModels {
		if cmd := tool.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// createToolModels builds tool models from the message content.
// Each tool gets its specific executor.
func (m *Model) createToolModels() []*tools.Model {
	var models []*tools.Model

	for _, block := range m.message.Content {
		if block.Type != domain.BlockTypeToolUse || block.ToolUse == nil {
			continue
		}

		use := block.ToolUse
		var body tools.Body

		switch use.Name {
		case appquery.Name:
			body = query.New(m.theme, m.logger, use, m.executors.Query)
		case appstartjourney.Name:
			body = startjourney.New(m.theme, m.logger, use, m.executors.StartJourney)
		case appendjourney.Name:
			body = endjourney.New(m.theme, m.logger, use, m.executors.EndJourney)
		default:
			body = generic.New(m.theme, m.logger, use)
		}

		model := tools.New(m.theme, m.logger, use, body)
		models = append(models, model)
	}

	return models
}

// ID returns the message ID.
func (m *Model) ID() string {
	return m.message.ID.String()
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (messages.Item, tea.Cmd) {
	var cmds []tea.Cmd

	// Forward to tool models
	for i, tool := range m.toolModels {
		updated, cmd := tool.Update(msg)
		if t, ok := updated.(*tools.Model); ok {
			m.toolModels[i] = t
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		m.cachedRender = ""
	}

	return m, tea.Batch(cmds...)
}

// leftPadding is the total left inset (border + padding) for coordinate adjustment.
const leftPadding = 1

// Render returns the rendered message.
func (m *Model) Render(width int) string {
	// Skip cache if highlighted (highlighting changes appearance)
	if !m.highlight.IsHighlighted() && m.cachedWidth == width && m.cachedRender != "" {
		return m.cachedRender
	}

	colors := m.theme.Colors

	label := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientEnd).
		Bold(true).
		Render("Tero")

	var parts []string

	// Text content
	text := m.textContent()
	if text != "" {
		rendered := styles.RenderMarkdown(m.theme, text, width-2)
		rendered = strings.TrimRight(rendered, "\n")
		parts = append(parts, rendered)
	}

	// Tool models
	for _, tool := range m.toolModels {
		parts = append(parts, tool.Render(width-2))
	}

	// Placeholder if empty
	if len(parts) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render("...")
		parts = append(parts, placeholder)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	containerStyle := lipgloss.NewStyle().
		PaddingLeft(1)

	if m.focused {
		containerStyle = containerStyle.
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "▌"}).
			BorderForeground(colors.Brand.GradientEnd).
			PaddingLeft(0)
	}

	result := containerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, label, content),
	)

	// Apply highlighting if set
	if m.highlight.IsHighlighted() {
		height := strings.Count(result, "\n") + 1
		result = m.highlight.RenderHighlighted(result, width, height)
	} else {
		m.cachedRender = result
		m.cachedWidth = width
	}

	return result
}

// Height returns the rendered height.
func (m *Model) Height(width int) int {
	rendered := m.Render(width)
	return strings.Count(rendered, "\n") + 1
}

// SetFocused updates the focus state.
func (m *Model) SetFocused(focused bool) messages.Item {
	if m.focused == focused {
		return m
	}
	m.focused = focused
	m.cachedRender = ""
	return m
}

// CopyableContent returns the text content for clipboard.
func (m *Model) CopyableContent() string {
	return m.textContent()
}

func (m *Model) textContent() string {
	var texts []string
	for _, block := range m.message.Content {
		if block.Type == domain.BlockTypeText && block.Text != nil {
			texts = append(texts, block.Text.Content)
		}
	}
	return strings.Join(texts, "\n")
}

// Tools returns the tool models.
func (m *Model) Tools() []*tools.Model {
	return m.toolModels
}

// AllToolsComplete returns true if all tools have results.
func (m *Model) AllToolsComplete() bool {
	for _, tool := range m.toolModels {
		if tool.Result() == nil {
			return false
		}
	}
	return true
}

// ToolResults returns all tool results for sending back to the API.
func (m *Model) ToolResults() []domain.Block {
	var blocks []domain.Block
	for _, tool := range m.toolModels {
		if result := tool.Result(); result != nil {
			blocks = append(blocks, domain.Block{
				Type:       domain.BlockTypeToolResult,
				ToolResult: result,
			})
		}
	}
	return blocks
}

// HasTools returns true if this message has tool_use blocks.
func (m *Model) HasTools() bool {
	return len(m.toolModels) > 0
}

// SetHighlight sets the highlight range, adjusting for the left padding.
func (m *Model) SetHighlight(startLine, startCol, endLine, endCol int) {
	m.highlight.SetHighlightWithOffset(startLine, startCol, endLine, endCol, leftPadding)
}

// Highlight returns the current highlight range.
func (m *Model) Highlight() (startLine, startCol, endLine, endCol int) {
	return m.highlight.Highlight()
}

// ClearHighlight removes any highlight.
func (m *Model) ClearHighlight() {
	m.highlight.ClearHighlight()
}

// IsHighlighted returns true if the item has a highlight set.
func (m *Model) IsHighlighted() bool {
	return m.highlight.IsHighlighted()
}

// HighlightedContent returns the plain text content of the highlighted region.
func (m *Model) HighlightedContent() string {
	if !m.highlight.IsHighlighted() {
		return ""
	}
	rendered := m.Render(m.cachedWidth)
	height := strings.Count(rendered, "\n") + 1
	return m.highlight.HighlightedContent(rendered, m.cachedWidth, height)
}
