package list

import (
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/styles"
)

const (
	// MaxListHeight is the maximum height for lists
	// Set to 11 to accommodate 10 items + 1 action item (e.g. "Create new")
	MaxListHeight = 11
)

// Re-export types from bubbles/list so callers don't need to import it directly
type (
	Item         = list.Item
	ItemDelegate = list.ItemDelegate
	KeyMap       = list.KeyMap
	Model        = list.Model
)

// List is a wrapper around bubbles list with consistent theming
type List struct {
	list list.Model
}

// Compile-time check that List implements components.Component
var _ components.Component = (*List)(nil)

// New creates a new list with consistent theming and filtering enabled
func New(items []Item, delegate ItemDelegate) *List {
	theme := styles.CurrentTheme()

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false) // Disable pagination UI - we'll enable it only when needed

	// Apply theme-aware styles
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	l.Styles.TitleBar = lipgloss.NewStyle().
		Foreground(theme.Primary)

	// Set initial height based on item count (scales down for fewer items)
	// Since pagination is disabled by default, we don't need extra space
	itemCount := len(items)
	listHeight := itemCount
	if listHeight > MaxListHeight {
		listHeight = MaxListHeight
	}
	if listHeight > 0 {
		l.SetHeight(listHeight)
	}

	return &List{
		list: l,
	}
}

// Init initializes the component
func (c *List) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c *List) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return cmd
}

// View renders the list
func (c *List) View() string {
	return c.list.View()
}

// SetItems sets the list items
func (c *List) SetItems(items []Item) tea.Cmd {
	return c.list.SetItems(items)
}

// SelectedItem returns the currently selected item
func (c *List) SelectedItem() Item {
	return c.list.SelectedItem()
}

// SetSize sets both width and height
func (c *List) SetSize(width, height int) {
	c.list.SetWidth(width)
	c.list.SetHeight(height)
}

// SetWidth sets the list width
func (c *List) SetWidth(width int) {
	c.list.SetWidth(width)
}

// SetHeight sets the list height
func (c *List) SetHeight(height int) {
	c.list.SetHeight(height)
}

// SetShowPagination enables or disables the pagination UI
func (c *List) ShowPagination() bool {
	return c.list.ShowPagination()
}

// SetFilteringEnabled enables or disables filtering
func (c *List) SetFilteringEnabled(enabled bool) {
	c.list.SetFilteringEnabled(enabled)
}

// FilteringEnabled returns whether filtering is enabled
func (c *List) FilteringEnabled() bool {
	return c.list.FilteringEnabled()
}

// SetShowPagination enables or disables the pagination UI
func (c *List) SetShowPagination(show bool) {
	c.list.SetShowPagination(show)
}

// KeyMap returns the key bindings for list navigation
func (c List) KeyMap() KeyMap {
	return c.list.KeyMap
}

// Index returns the index of the currently selected item
func (c List) Index() int {
	return c.list.Index()
}

// DebugPagination returns pagination state for debugging
func (c List) DebugPagination() (perPage int, totalPages int, currentPage int) {
	return c.list.Paginator.PerPage, c.list.Paginator.TotalPages, c.list.Paginator.Page
}

// IsBusy returns false - list components are never busy (they display static data)
func (c *List) IsBusy() bool {
	return false
}

// HasError returns false - list components don't have error states
func (c *List) HasError() bool {
	return false
}

// Error returns nil - list components don't have errors
func (c *List) Error() error {
	return nil
}
