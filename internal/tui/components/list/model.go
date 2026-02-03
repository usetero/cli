package list

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

const (
	MaxHeight = 11
)

// Re-export types from bubbles/list
type (
	Item         = list.Item
	ItemDelegate = list.ItemDelegate
	KeyMap       = list.KeyMap
)

// Model is a wrapper around bubbles list with consistent theming.
type Model struct {
	theme *styles.Theme
	list  list.Model
}

// New creates a new list model.
func New(theme *styles.Theme, items []Item, delegate ItemDelegate) Model {
	colors := theme.Colors

	l := list.New(items, delegate, 0, 0)
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

	return Model{
		theme: theme,
		list:  l,
	}
}

// Init initializes the list.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list.
func (m Model) View() string {
	return m.list.View()
}

// SetItems returns a new Model with the given items.
func (m Model) SetItems(items []Item) (Model, tea.Cmd) {
	cmd := m.list.SetItems(items)
	return m, cmd
}

// SelectedItem returns the currently selected item.
func (m Model) SelectedItem() Item {
	return m.list.SelectedItem()
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.list.SetWidth(width)
	m.list.SetHeight(height)
	return m
}

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.list.SetWidth(width)
	return m
}

// SetHeight returns a new Model with the given height.
func (m Model) SetHeight(height int) Model {
	m.list.SetHeight(height)
	return m
}

// SetShowPagination returns a new Model with pagination visibility set.
func (m Model) SetShowPagination(show bool) Model {
	m.list.SetShowPagination(show)
	return m
}

// SetFilteringEnabled returns a new Model with filtering enabled/disabled.
func (m Model) SetFilteringEnabled(enabled bool) Model {
	m.list.SetFilteringEnabled(enabled)
	return m
}

// FilteringEnabled returns whether filtering is enabled.
func (m Model) FilteringEnabled() bool {
	return m.list.FilteringEnabled()
}

// KeyMap returns the key bindings for list navigation.
func (m Model) KeyMap() KeyMap {
	return m.list.KeyMap
}

// Index returns the index of the currently selected item.
func (m Model) Index() int {
	return m.list.Index()
}
