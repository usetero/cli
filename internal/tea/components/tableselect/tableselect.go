// Package tableselect wraps bubbles/table with theme-aware selection styling.
package tableselect

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
)

// Re-export types for convenience.
type (
	Column = table.Column
	Row    = table.Row
)

// Model wraps bubbles/table with themed selection highlight.
type Model struct {
	theme styles.Theme
	table table.Model
}

// New creates a new selectable table with themed styles.
func New(theme styles.Theme, cols []Column, rows []Row) *Model {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(theme.Text).
		Bold(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Padding(0, 1)
	// Auto-size width from column definitions + cell padding.
	w := 0
	for _, col := range cols {
		w += col.Width + 2 // +2 for Padding(0, 1)
	}

	s.Selected = lipgloss.NewStyle().
		Background(theme.SelectionBg).
		Foreground(theme.SelectionFg).
		Bold(true).
		Width(w)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithStyles(s),
		table.WithFocused(true),
		table.WithWidth(w),
		table.WithHeight(len(rows)+1),
	)

	return &Model{
		theme: theme,
		table: t,
	}
}

// Update handles key events for table navigation.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

// View renders the table with a themed border.
func (m *Model) View() string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border)
	return border.Render(m.table.View())
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.table.Cursor()
}

// SelectedRow returns the selected row data.
func (m *Model) SelectedRow() Row {
	return m.table.SelectedRow()
}

// SetRows replaces the table rows.
func (m *Model) SetRows(rows []Row) {
	m.table.SetRows(rows)
}

// SetHeight sets the table height.
func (m *Model) SetHeight(h int) {
	m.table.SetHeight(h)
}

// SetWidth sets the table width.
func (m *Model) SetWidth(w int) {
	m.table.SetWidth(w)
}

// ShortHelp returns key bindings for table navigation.
func (m *Model) ShortHelp() []key.Binding {
	return m.table.KeyMap.ShortHelp()
}
