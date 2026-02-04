package user

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages"
	"github.com/usetero/cli/internal/tui/components/highlight"
)

// Model displays a user message in the chat.
type Model struct {
	theme   *styles.Theme
	logger  log.Logger
	message domain.Message
	focused bool

	// Highlighting support
	highlight highlight.Item

	cachedRender string
	cachedWidth  int
}

// Compile-time interface checks
var (
	_ messages.Item          = (*Model)(nil)
	_ messages.Copyable      = (*Model)(nil)
	_ messages.Highlightable = (*Model)(nil)
)

// New creates a new user message model.
func New(theme *styles.Theme, logger log.Logger, message domain.Message) *Model {
	return &Model{
		theme:     theme,
		logger:    logger,
		message:   message,
		highlight: highlight.NewItem(),
	}
}

// ID returns the message ID.
func (m *Model) ID() string {
	return m.message.ID.String()
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (messages.Item, tea.Cmd) {
	return m, nil
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
		Foreground(colors.Accent).
		Bold(true).
		Render("You")

	text := m.textContent()

	contentStyle := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Width(width - 2)

	content := contentStyle.Render(text)

	containerStyle := lipgloss.NewStyle().
		PaddingLeft(1)

	if m.focused {
		containerStyle = containerStyle.
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "▌"}).
			BorderForeground(colors.Accent).
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
