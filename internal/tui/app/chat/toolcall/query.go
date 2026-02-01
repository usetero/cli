package toolcall

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderQuery(theme *styles.Theme, input *block.QueryInput, result *block.QueryResult, width int) string {
	if result == nil {
		return successText(theme, "Query executed")
	}

	colors := theme.Colors
	var parts []string

	// Show the SQL query
	if input != nil && input.SQL != "" {
		sql := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Italic(true).
			PaddingLeft(2).
			Width(width - 4).
			Render(input.SQL)
		parts = append(parts, sql)
	}

	// Show result summary
	rowCount := len(result.Rows)
	colCount := len(result.Columns)
	summary := lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		PaddingLeft(2).
		Render(fmt.Sprintf("%d rows, %d columns", rowCount, colCount))
	parts = append(parts, summary)

	// Show table preview (first few rows)
	if rowCount > 0 && colCount > 0 {
		table := renderQueryTable(theme, result, width-4, 5) // max 5 rows
		if table != "" {
			parts = append(parts, table)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func renderQueryTable(theme *styles.Theme, result *block.QueryResult, width int, maxRows int) string {
	colors := theme.Colors

	// Calculate column widths
	colWidths := make([]int, len(result.Columns))
	for i, col := range result.Columns {
		colWidths[i] = len(col)
	}
	for _, row := range result.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellStr := fmt.Sprintf("%v", cell)
				if len(cellStr) > colWidths[i] {
					colWidths[i] = len(cellStr)
				}
			}
		}
	}

	// Cap column widths
	maxColWidth := 20
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	var lines []string

	// Header
	var headerCells []string
	for i, col := range result.Columns {
		cell := truncate(col, colWidths[i])
		cell = padRight(cell, colWidths[i])
		headerCells = append(headerCells, cell)
	}
	header := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		PaddingLeft(2).
		Render(strings.Join(headerCells, " | "))
	lines = append(lines, header)

	// Rows
	rowsToShow := len(result.Rows)
	if rowsToShow > maxRows {
		rowsToShow = maxRows
	}
	for i := 0; i < rowsToShow; i++ {
		row := result.Rows[i]
		var rowCells []string
		for j, cell := range row {
			if j < len(colWidths) {
				cellStr := truncate(fmt.Sprintf("%v", cell), colWidths[j])
				cellStr = padRight(cellStr, colWidths[j])
				rowCells = append(rowCells, cellStr)
			}
		}
		rowLine := lipgloss.NewStyle().
			Foreground(colors.Page.Text).
			PaddingLeft(2).
			Render(strings.Join(rowCells, " | "))
		lines = append(lines, rowLine)
	}

	// Show truncation indicator
	if len(result.Rows) > maxRows {
		more := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			PaddingLeft(2).
			Render(fmt.Sprintf("... and %d more rows", len(result.Rows)-maxRows))
		lines = append(lines, more)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
