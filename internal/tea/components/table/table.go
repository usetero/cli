// Package table provides a themed table component.
package table

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/styles"
)

const (
	defaultMinHeaderWidth = 8
	defaultMaxValueWidth  = 20
)

// Model wraps lipgloss/table with theme-aware styling.
type Model struct {
	theme          styles.Theme
	headers        []string
	rows           [][]string
	width          int
	fitHeaders     bool        // truncate headers to fit widest value in column
	minHeaderWidth int         // minimum header width when fitHeaders is on
	maxValueWidth  int         // truncate cell values wider than this
	bg             color.Color // background for cells (needed because lipgloss table resets styles between cells)
}

// Option configures a table.
type Option func(*Model)

// WithFitHeaders truncates column headers to the widest value in each column,
// with a minimum width of 5. Use WithMinHeaderWidth to override the minimum.
func WithFitHeaders() Option {
	return func(m *Model) {
		m.fitHeaders = true
		if m.minHeaderWidth == 0 {
			m.minHeaderWidth = defaultMinHeaderWidth
		}
	}
}

// WithMinHeaderWidth sets the minimum header width when FitHeaders is enabled.
func WithMinHeaderWidth(n int) Option {
	return func(m *Model) {
		m.minHeaderWidth = n
	}
}

// WithMaxValueWidth sets the maximum width for cell values.
// Values wider than this are truncated with an ellipsis. Default is 20.
func WithMaxValueWidth(n int) Option {
	return func(m *Model) {
		m.maxValueWidth = n
	}
}

// WithBackground sets the background color for cells and borders.
// Needed because lipgloss table resets ANSI styles between cells,
// so a parent container's background doesn't propagate through.
func WithBackground(c color.Color) Option {
	return func(m *Model) {
		m.bg = c
	}
}

// New creates a new table.
func New(theme styles.Theme, opts ...Option) *Model {
	m := &Model{
		theme:         theme,
		maxValueWidth: defaultMaxValueWidth,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
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

	if m.bg != nil {
		headerStyle = headerStyle.Background(m.bg)
		cellStyle = cellStyle.Background(m.bg)
		borderStyle = borderStyle.Background(m.bg)
	}

	headers := m.headers
	if m.fitHeaders && len(m.rows) > 0 {
		headers = m.truncatedHeaders()
	}

	tbl := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(true).
		BorderColumn(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		}).
		Headers(headers...)

	for _, row := range m.rows {
		if m.maxValueWidth > 0 {
			truncated := make([]string, len(row))
			for i, cell := range row {
				if ansi.StringWidth(cell) > m.maxValueWidth {
					truncated[i] = ansi.Truncate(cell, m.maxValueWidth-1, "") + "…"
				} else {
					truncated[i] = cell
				}
			}
			tbl.Row(truncated...)
		} else {
			tbl.Row(row...)
		}
	}

	if m.width > 0 {
		tbl.Width(m.width)
	}

	return tbl.Render()
}

// truncatedHeaders returns headers truncated to the widest value in each column.
func (m *Model) truncatedHeaders() []string {
	result := make([]string, len(m.headers))
	for col, header := range m.headers {
		maxVal := 0
		for _, row := range m.rows {
			if col < len(row) {
				w := ansi.StringWidth(row[col])
				if m.maxValueWidth > 0 && w > m.maxValueWidth {
					w = m.maxValueWidth
				}
				if w > maxVal {
					maxVal = w
				}
			}
		}
		targetWidth := max(maxVal, m.minHeaderWidth)
		headerWidth := ansi.StringWidth(header)
		if targetWidth > 0 && headerWidth > targetWidth {
			result[col] = ansi.Truncate(header, targetWidth-1, "") + "…"
		} else {
			result[col] = header
		}
	}
	return result
}
