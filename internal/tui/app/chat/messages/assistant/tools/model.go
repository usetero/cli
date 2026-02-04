package tools

import (
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages"
)

// Status represents the current state of a tool execution.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusError
)

// Body is the interface for tool-specific models.
// Each tool model implements this and handles its own execution.
type Body interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	Result() *domain.ToolResult
	Render(width int) string
	Params() []string
}

// Model displays a tool execution in the chat.
// It provides chrome (header, icon, container) around a Body.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	use      *domain.ToolUse
	body     Body
	focused  bool
	expanded bool
}

// Compile-time interface checks
var (
	_ messages.Item       = (*Model)(nil)
	_ messages.Expandable = (*Model)(nil)
)

// New creates a new tool model with the given body.
// The body is created by the caller with its specific executor.
func New(theme *styles.Theme, logger log.Logger, use *domain.ToolUse, body Body) *Model {
	return &Model{
		theme:  theme,
		logger: logger,
		use:    use,
		body:   body,
	}
}

// ID returns the tool use ID.
func (m *Model) ID() string {
	return m.use.ID
}

// Init starts tool execution by delegating to the body.
func (m *Model) Init() tea.Cmd {
	return m.body.Init()
}

// Status returns the current tool status.
func (m *Model) Status() Status {
	result := m.body.Result()
	if result == nil {
		if m.use.Name != "" {
			return StatusRunning
		}
		return StatusPending
	}
	if result.IsError {
		return StatusError
	}
	return StatusSuccess
}

// Update handles messages, delegates to body.
func (m *Model) Update(msg tea.Msg) (messages.Item, tea.Cmd) {
	cmd := m.body.Update(msg)
	return m, cmd
}

// Result returns the tool result from the body.
func (m *Model) Result() *domain.ToolResult {
	return m.body.Result()
}

// Render returns the rendered tool with chrome.
func (m *Model) Render(width int) string {
	body := m.body.Render(width - 4)
	return m.renderChrome(body)
}

func (m *Model) renderChrome(body string) string {
	header := m.renderHeader()
	content := header
	if body != "" {
		content = header + "\n" + body
	}
	return m.renderContainer(content)
}

func (m *Model) renderHeader() string {
	colors := m.theme.Colors

	icon := m.renderIcon()
	toolName := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Bold(true).
		Render(m.use.Name)

	header := fmt.Sprintf("%s %s", icon, toolName)

	params := m.body.Params()
	if len(params) > 0 {
		paramStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
		for _, p := range params {
			header += " " + paramStyle.Render(p)
		}
	}

	return header
}

func (m *Model) renderIcon() string {
	colors := m.theme.Colors
	switch m.Status() {
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(colors.Success.Fg).Render("●")
	case StatusError:
		return lipgloss.NewStyle().Foreground(colors.Error.Fg).Render("●")
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(colors.Warning.Fg).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render("○")
	}
}

func (m *Model) renderContainer(content string) string {
	colors := m.theme.Colors

	style := lipgloss.NewStyle().PaddingLeft(2)

	if m.focused {
		style = style.
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "▌"}).
			BorderForeground(colors.Brand.GradientStart).
			PaddingLeft(1)
	}

	return style.Render(content)
}

// Height returns the rendered height.
func (m *Model) Height(width int) int {
	rendered := m.Render(width)
	return strings.Count(rendered, "\n") + 1
}

// SetFocused updates the focus state.
func (m *Model) SetFocused(focused bool) messages.Item {
	m.focused = focused
	return m
}

// ToggleExpanded toggles the expanded state.
func (m *Model) ToggleExpanded() messages.Item {
	m.expanded = !m.expanded
	return m
}

// IsExpanded returns whether the model is expanded.
func (m *Model) IsExpanded() bool {
	return m.expanded
}
