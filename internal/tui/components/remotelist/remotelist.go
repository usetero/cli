package remotelist

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/loader"
)

type state int

const (
	stateLoading state = iota
	stateLoaded
	stateError
)

// LoadResultMsg is the standard message for load completion
type LoadResultMsg struct {
	Items []list.Item
	Err   error
}

// Component is a list that loads items asynchronously
type Component struct {
	list       *list.List
	loader     *loader.Component
	state      state
	err        error
	logger     log.Logger
	loaderFunc tea.Cmd // Store the loader function for retry
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new remote list component
func New(delegate list.ItemDelegate, loadingMessage string, logger log.Logger) *Component {
	l := list.New([]list.Item{}, delegate)
	// Don't set height here - let it be set when items load

	return &Component{
		list:   l,
		loader: loader.New(loadingMessage),
		state:  stateLoading,
		logger: logger,
	}
}

// Init starts loading items
func (c *Component) Init() tea.Cmd {
	return c.loader.Init()
}

// InitWithLoader starts loading items with a custom load function
func (c *Component) InitWithLoader(loadFunc tea.Cmd) tea.Cmd {
	c.loaderFunc = loadFunc // Store for retry
	return tea.Batch(
		c.loader.Init(),
		loadFunc,
	)
}

// Retry resets the error state and retries loading
func (c *Component) Retry() tea.Cmd {
	if c.loaderFunc == nil {
		return nil
	}
	c.logger.Debug("retrying load")
	c.state = stateLoading
	c.err = nil
	return tea.Batch(
		c.loader.Init(),
		c.loaderFunc,
	)
}

// Update handles messages
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case LoadResultMsg:
		if msg.Err != nil {
			c.state = stateError
			c.err = msg.Err
			c.logger.Error("failed to load list items", "error", msg.Err)
			return nil
		}
		c.state = stateLoaded

		// Calculate appropriate height based on item count
		itemCount := len(msg.Items)
		listHeight := itemCount
		// Add 1 for filter bar if filtering is enabled
		if c.list.FilteringEnabled() {
			listHeight += 1
		}
		if listHeight > list.MaxListHeight {
			listHeight = list.MaxListHeight
		}

		if listHeight > 0 {
			c.list.SetHeight(listHeight)
		}

		// Enable pagination UI only if we exceed max height
		c.list.SetShowPagination(itemCount > list.MaxListHeight)

		c.list.SetItems(msg.Items)

		return nil
	}

	// Delegate to appropriate component based on state
	switch c.state {
	case stateLoading:
		return c.loader.Update(msg)
	case stateLoaded:
		return c.list.Update(msg)
	case stateError:
		// No interaction while in error state
		return nil
	}

	return nil
}

// View renders the component
func (c *Component) View() string {
	switch c.state {
	case stateLoading:
		return c.loader.View()
	case stateError:
		// Error is shown in footer, just show empty view
		return ""
	case stateLoaded:
		return c.list.View()
	}
	return ""
}

// SetWidth sets the list width
func (c *Component) SetWidth(width int) {
	c.list.SetWidth(width)
}

// SetHeight sets the list height
func (c *Component) SetHeight(height int) {
	c.list.SetHeight(height)
}

// SelectedItem returns the currently selected item
func (c Component) SelectedItem() list.Item {
	return c.list.SelectedItem()
}

// IsLoaded returns true if loading is complete (successfully or with error)
func (c Component) IsLoaded() bool {
	return c.state == stateLoaded || c.state == stateError
}

// IsBusy returns true if the component is performing background work
func (c Component) IsBusy() bool {
	return c.state == stateLoading
}

// HasError returns true if loading failed with an error
func (c Component) HasError() bool {
	return c.state == stateError
}

// Error returns the current error, or nil if no error
func (c Component) Error() error {
	if c.state == stateError {
		return c.err
	}
	return nil
}

// KeyMap returns the key bindings for list navigation
func (c Component) KeyMap() list.KeyMap {
	return c.list.KeyMap()
}
