package chat

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain/tool"
	"github.com/usetero/cli/internal/styles"
)

// ToolBlock displays a tool use and its result.
type ToolBlock struct {
	theme  *styles.Theme
	use    *tool.Use
	result *tool.Result
	width  int
}

// NewToolBlock creates a new tool block component.
func NewToolBlock(theme *styles.Theme, use *tool.Use) ToolBlock {
	return ToolBlock{
		theme: theme,
		use:   use,
	}
}

// Init initializes the component.
func (m ToolBlock) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m ToolBlock) Update(msg tea.Msg) (ToolBlock, tea.Cmd) {
	return m, nil
}

// View renders the tool block.
func (m ToolBlock) View() string {
	colors := m.theme.Colors

	// Tool name header
	name := string(m.use.Name)
	header := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Render("[" + name + "]")

	// Show result if available
	if m.result != nil {
		if m.result.IsError {
			errorText := lipgloss.NewStyle().
				Foreground(colors.Error.Fg).
				Width(m.width).
				Render(m.result.Error)
			return lipgloss.JoinVertical(lipgloss.Left, header, errorText)
		}
		// For now just show a checkmark for success
		success := lipgloss.NewStyle().
			Foreground(colors.Success.Fg).
			Render("Done")
		return lipgloss.JoinVertical(lipgloss.Left, header, success)
	}

	// Still running
	running := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Italic(true).
		Render("Running...")

	return lipgloss.JoinVertical(lipgloss.Left, header, running)
}

// SetWidth returns a new ToolBlock with the given width.
func (m ToolBlock) SetWidth(width int) ToolBlock {
	m.width = width
	return m
}

// SetResult returns a new ToolBlock with the given result.
func (m ToolBlock) SetResult(result *tool.Result) ToolBlock {
	m.result = result
	return m
}

// ID returns the tool use ID.
func (m ToolBlock) ID() string {
	return m.use.ID
}

// IsComplete returns true if the tool has a result.
func (m ToolBlock) IsComplete() bool {
	return m.result != nil
}
