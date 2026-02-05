// Package list provides a themed list component.
package list

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
)

// MaxHeight is the default maximum height for lists.
const MaxHeight = 11

// Re-export types from bubbles/list for convenience.
type (
	Item   = list.Item
	KeyMap = list.KeyMap
)

// delegate is the default item delegate using theme colors.
type delegate struct {
	theme *styles.Theme
}

func (d *delegate) Height() int                             { return 1 }
func (d *delegate) Spacing() int                            { return 0 }
func (d *delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d *delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	name := item.FilterValue()
	colors := d.theme.Colors
	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Accent).Bold(true).Render("> "+name))
	} else {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Page.Text).Render("  "+name))
	}
}

// Model wraps bubbles list with consistent theming.
type Model struct {
	theme *styles.Theme
	list  list.Model
}

// New creates a new themed list.
func New(theme *styles.Theme, items []Item) *Model {
	colors := theme.Colors
	d := &delegate{theme: theme}

	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	l.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)

	l.Styles.TitleBar = lipgloss.NewStyle().
		Foreground(colors.Accent)

	itemCount := len(items)
	listHeight := itemCount
	if listHeight > MaxHeight {
		listHeight = MaxHeight
	}
	if listHeight > 0 {
		l.SetHeight(listHeight)
	}

	return &Model{
		theme: theme,
		list:  l,
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

// View renders the list.
func (m *Model) View() string {
	return m.list.View()
}

// SetItems updates the list items.
func (m *Model) SetItems(items []Item) tea.Cmd {
	return m.list.SetItems(items)
}

// SelectedItem returns the currently selected item.
func (m *Model) SelectedItem() Item {
	return m.list.SelectedItem()
}

// SetSize sets both width and height.
func (m *Model) SetSize(width, height int) {
	m.list.SetWidth(width)
	m.list.SetHeight(height)
}

// SetWidth sets the list width.
func (m *Model) SetWidth(width int) {
	m.list.SetWidth(width)
}

// SetHeight sets the list height.
func (m *Model) SetHeight(height int) {
	m.list.SetHeight(height)
}

// SetShowPagination controls pagination visibility.
func (m *Model) SetShowPagination(show bool) {
	m.list.SetShowPagination(show)
}

// SetFilteringEnabled controls whether filtering is enabled.
func (m *Model) SetFilteringEnabled(enabled bool) {
	m.list.SetFilteringEnabled(enabled)
}

// FilteringEnabled returns whether filtering is enabled.
func (m *Model) FilteringEnabled() bool {
	return m.list.FilteringEnabled()
}

// KeyMap returns the key bindings for list navigation.
func (m *Model) KeyMap() KeyMap {
	return m.list.KeyMap
}

// Index returns the index of the currently selected item.
func (m *Model) Index() int {
	return m.list.Index()
}
