package toolcall

import (
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

// RenderResult renders a tool result based on the tool type.
// Each tool type has its own rendering logic.
func RenderResult(theme *styles.Theme, toolUse *block.ToolUse, result *block.ToolResult, width int) string {
	switch toolUse.Name {
	case block.ToolQuery:
		return renderQuery(theme, toolUse.Query, result.Query, width)
	case block.ToolShowMetric:
		return renderMetric(theme, toolUse.ShowMetric, width)
	case block.ToolShowSeries:
		return renderSeries(theme, toolUse.ShowSeries, width)
	case block.ToolShowTimeSeries:
		return renderTimeSeries(theme, toolUse.ShowTimeSeries, width)
	case block.ToolShowTable:
		return renderTable(theme, toolUse.ShowTable, width)
	case block.ToolAddContext:
		return renderAddContext(theme, toolUse.AddContext, result.AddContext, width)
	case block.ToolRemoveContext:
		return renderRemoveContext(theme, toolUse.RemoveContext, result.RemoveContext, width)
	case block.ToolStartJourney:
		return renderStartJourney(theme, toolUse.StartJourney, width)
	case block.ToolEndJourney:
		return renderEndJourney(theme, toolUse.EndJourney, width)
	case block.ToolApprovePolicy:
		return renderApprovePolicy(theme, toolUse.ApprovePolicy, result.ApprovePolicy, width)
	case block.ToolDismissPolicy:
		return renderDismissPolicy(theme, toolUse.DismissPolicy, result.DismissPolicy, width)
	default:
		return ""
	}
}

// successText renders a simple success message.
func successText(theme *styles.Theme, text string) string {
	return lipgloss.NewStyle().
		Foreground(theme.Colors.Success.Fg).
		PaddingLeft(2).
		Render(text)
}

// mutedText renders muted informational text.
func mutedText(theme *styles.Theme, text string, width int) string {
	return lipgloss.NewStyle().
		Foreground(theme.Colors.Page.TextMuted).
		PaddingLeft(2).
		Width(width - 4).
		Render(text)
}
