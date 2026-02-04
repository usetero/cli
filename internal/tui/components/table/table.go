// Package table provides a themed table component wrapping lipgloss/table.
package table

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/usetero/cli/internal/styles"
)

// Table wraps lipgloss/table with theme-aware styling.
type Table struct {
	theme   *styles.Theme
	headers []string
	rows    [][]string
	width   int
}

// New creates a new table.
func New(theme *styles.Theme) *Table {
	return &Table{
		theme: theme,
	}
}

// Headers sets the column headers.
func (t *Table) Headers(headers ...string) *Table {
	t.headers = headers
	return t
}

// Row adds a row to the table.
func (t *Table) Row(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Rows adds multiple rows to the table.
func (t *Table) Rows(rows [][]string) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

// Width sets the maximum width for the table.
func (t *Table) Width(width int) *Table {
	t.width = width
	return t
}

// Render returns the rendered table string.
func (t *Table) Render() string {
	colors := t.theme.Colors

	headerStyle := lipgloss.NewStyle().
		Foreground(colors.Page.Text).
		Bold(true).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Foreground(colors.BorderDefault)

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		}).
		Headers(t.headers...)

	for _, row := range t.rows {
		tbl.Row(row...)
	}

	if t.width > 0 {
		tbl.Width(t.width)
	}

	return tbl.Render()
}
