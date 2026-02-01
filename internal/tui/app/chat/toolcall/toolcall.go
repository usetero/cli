// Package toolcall provides the ToolCall component for rendering tool invocations.
package toolcall

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/spinner"
)

// ToolCall displays a tool invocation with loading state and result.
// Rendering of the result is delegated to tool-specific renderers.
type ToolCall struct {
	theme *styles.Theme
	width int

	// Tool data
	toolUse *block.ToolUse
	result  *block.ToolResult

	// Loading state
	loading bool
	spinner *spinner.Spinner
}

// New creates a new ToolCall component.
func New(theme *styles.Theme, toolUse *block.ToolUse) *ToolCall {
	return &ToolCall{
		theme:   theme,
		toolUse: toolUse,
		loading: true,
		spinner: spinner.New(spinner.Settings{
			Size:        10,
			Label:       "Working",
			GradColorA:  theme.Colors.Brand.GradientStart,
			GradColorB:  theme.Colors.Brand.GradientEnd,
			LabelColor:  theme.Colors.Page.TextMuted,
			CycleColors: true,
		}),
	}
}

// ID returns the tool call ID.
func (t *ToolCall) ID() string {
	if t.toolUse == nil {
		return ""
	}
	return t.toolUse.ID
}

// Init initializes the component.
func (t *ToolCall) Init() tea.Cmd {
	if t.spinner != nil && t.loading {
		return t.spinner.Init()
	}
	return nil
}

// Update handles messages.
func (t *ToolCall) Update(msg tea.Msg) tea.Cmd {
	if t.spinner != nil && t.loading {
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return cmd
	}
	return nil
}

// View renders the tool call.
func (t *ToolCall) View() string {
	if t.toolUse == nil {
		return ""
	}

	header := t.viewHeader()

	// Show spinner while loading
	if t.loading && t.spinner != nil {
		spinnerView := lipgloss.NewStyle().
			PaddingLeft(2).
			Render(t.spinner.View())
		return lipgloss.JoinVertical(lipgloss.Left, header, spinnerView)
	}

	// Render result
	resultView := t.viewResult()
	if resultView != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, resultView)
	}

	// Completed with no displayable result
	colors := t.theme.Colors
	done := lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		PaddingLeft(2).
		Render("Done")
	return lipgloss.JoinVertical(lipgloss.Left, header, done)
}

// SetWidth sets the available width for rendering.
func (t *ToolCall) SetWidth(width int) {
	t.width = width
}

// SetResult sets the tool result and stops the loading animation.
func (t *ToolCall) SetResult(result *block.ToolResult) {
	t.result = result
	t.loading = false
}

// Spinning returns true if showing a loading animation.
func (t *ToolCall) Spinning() bool {
	return t.loading && t.spinner != nil
}

// viewHeader renders the tool name.
func (t *ToolCall) viewHeader() string {
	colors := t.theme.Colors
	name := PrettifyName(t.toolUse.Name)

	return lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Render("Tool: " + name)
}

// viewResult renders the tool result, delegating to tool-specific renderers.
func (t *ToolCall) viewResult() string {
	if t.result == nil {
		return ""
	}

	// Error case
	if t.result.IsError {
		return t.viewError()
	}

	// Delegate to tool-specific renderer
	return RenderResult(t.theme, t.toolUse, t.result, t.width)
}

// viewError renders an error result.
func (t *ToolCall) viewError() string {
	colors := t.theme.Colors

	return lipgloss.NewStyle().
		Foreground(colors.Error.Fg).
		PaddingLeft(2).
		Width(t.width - 4).
		Render("Error: " + t.result.Error)
}

// PrettifyName converts a tool name to a display-friendly format.
func PrettifyName(name block.ToolName) string {
	switch name {
	case block.ToolQuery:
		return "Query"
	case block.ToolShowMetric:
		return "Show Metric"
	case block.ToolShowSeries:
		return "Show Series"
	case block.ToolShowTimeSeries:
		return "Show Time Series"
	case block.ToolShowTable:
		return "Show Table"
	case block.ToolAddContext:
		return "Add Context"
	case block.ToolRemoveContext:
		return "Remove Context"
	case block.ToolStartJourney:
		return "Start Journey"
	case block.ToolEndJourney:
		return "End Journey"
	case block.ToolApprovePolicy:
		return "Approve Policy"
	case block.ToolDismissPolicy:
		return "Dismiss Policy"
	default:
		return string(name)
	}
}
