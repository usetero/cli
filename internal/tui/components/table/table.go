package table

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components"
)

// Re-export types from bubbles for convenience
type Column = table.Column
type Row = table.Row

// Table is a wrapper around bubbles table with consistent theming
type Table struct {
	theme *styles.Theme
	table table.Model
}

// Compile-time check that Table implements components.Component
var _ components.Component = (*Table)(nil)

// New creates a new table with consistent theming
func New(theme *styles.Theme, columns []Column) *Table {
	colors := theme.Colors

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false), // No cursor by default
	)

	// Apply theme-aware styles
	t.SetStyles(table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Accent).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.BorderDefault).
			BorderBottom(true),
		Selected: lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true),
		Cell: lipgloss.NewStyle().
			Foreground(colors.Page.Text),
	})

	return &Table{
		theme: theme,
		table: t,
	}
}

// Init initializes the component
func (c *Table) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c *Table) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.table, cmd = c.table.Update(msg)
	return cmd
}

// View renders the table
func (c *Table) View() string {
	return c.table.View()
}

// SetRows replaces all rows in the table
func (c *Table) SetRows(rows []Row) {
	c.table.SetRows(rows)
}

// AppendRow adds a single row to the table
func (c *Table) AppendRow(row Row) {
	currentRows := c.table.Rows()
	c.table.SetRows(append(currentRows, row))
}

// Rows returns the current rows
func (c *Table) Rows() []Row {
	return c.table.Rows()
}

// SetWidth sets the table width
func (c *Table) SetWidth(width int) {
	c.table.SetWidth(width)
}

// SetHeight sets the table height
func (c *Table) SetHeight(height int) {
	c.table.SetHeight(height)
}

// SetFocused sets whether the table is focused (shows cursor)
func (c *Table) SetFocused(focused bool) {
	if focused {
		c.table.Focus()
	} else {
		c.table.Blur()
	}
}

// IsBusy returns false - tables are never busy
func (c *Table) IsBusy() bool {
	return false
}

// HasError returns false - tables have no error state
func (c *Table) HasError() bool {
	return false
}

// Error returns nil - tables have no error state
func (c *Table) Error() error {
	return nil
}
