// Package table provides a themed table component.
package table

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/usetero/cli/internal/styles"
)

// Model wraps lipgloss/table with theme-aware styling.
type Model struct {
	theme   *styles.Theme
	headers []string
	rows    [][]string
	width   int
}

// New creates a new table.
func New(theme *styles.Theme) *Model {
	return &Model{theme: theme}
}

// Headers sets the column headers.
func (m *Model) Headers(headers ...string) {
	m.headers = headers
}

// Row adds a single row to the table.
func (m *Model) Row(cells ...string) {
	m.rows = append(m.rows, cells)
}

// Rows adds multiple rows to the table.
func (m *Model) Rows(rows [][]string) {
	m.rows = append(m.rows, rows...)
}

// SetWidth sets the maximum width for the table.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// Clear removes all rows (keeps headers).
func (m *Model) Clear() {
	m.rows = nil
}

// View renders the table.
func (m *Model) View() string {
	colors := m.theme.Colors

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
		Headers(m.headers...)

	for _, row := range m.rows {
		tbl.Row(row...)
	}

	if m.width > 0 {
		tbl.Width(m.width)
	}

	return tbl.Render()
}
