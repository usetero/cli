package generic

import (
	"strings"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
)

// Model renders unknown tools. Cannot execute - returns error.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	use      *domain.ToolUse
	result   *domain.ToolResult
	expanded bool
}

// Compile-time interface check
var _ tools.Body = (*Model)(nil)

// New creates a new generic model for unknown tools.
func New(theme *styles.Theme, logger log.Logger, use *domain.ToolUse) *Model {
	return &Model{
		theme:  theme,
		logger: logger,
		use:    use,
	}
}

// Init returns error - unknown tools can't be executed locally.
func (m *Model) Init() tea.Cmd {
	use := m.use
	logger := m.logger

	return func() tea.Msg {
		logger.Warn("unknown tool", "name", use.Name)
		return tools.ResultMsg{
			ToolUseID: use.ID,
			Result: &domain.ToolResult{
				ToolUseID: use.ID,
				IsError:   true,
				Error:     "unknown tool: " + use.Name,
			},
		}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tools.ResultMsg:
		if msg.ToolUseID == m.use.ID {
			m.result = msg.Result
		}
	}
	return nil
}

// Result returns the tool result.
func (m *Model) Result() *domain.ToolResult {
	return m.result
}

// Render returns the rendered body.
func (m *Model) Render(width int) string {
	colors := m.theme.Colors

	if m.result == nil {
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Italic(true).
			Render("Running...")
	}

	if m.result.IsError {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Render(m.result.Error)
	}

	if len(m.result.Content) > 0 {
		contentStr := string(m.result.Content)
		if m.expanded {
			return lipgloss.NewStyle().
				Foreground(colors.Page.TextMuted).
				Width(width).
				Render(contentStr)
		}
		preview := truncateLines(contentStr, 3, width)
		if strings.Count(contentStr, "\n") > 3 {
			preview += "\n" + lipgloss.NewStyle().
				Foreground(colors.Page.TextMuted).
				Italic(true).
				Render("[space to expand]")
		}
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Render(preview)
	}

	return lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		Render("Done")
}

// Params returns header params.
func (m *Model) Params() []string {
	return nil
}

// SetExpanded sets the expanded state.
func (m *Model) SetExpanded(expanded bool) {
	m.expanded = expanded
}

func truncateLines(content string, maxLines, maxWidth int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		if len(line) > maxWidth {
			lines[i] = line[:maxWidth-3] + "..."
		}
	}
	return strings.Join(lines, "\n")
}
