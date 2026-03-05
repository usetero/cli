// Package remotelist provides a list that loads items asynchronously.
package remotelist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/list"
	"github.com/usetero/cli/internal/tea/components/loader"
)

type state int

const (
	stateLoading state = iota
	stateLoaded
	stateError
)

// Item is re-exported from list for convenience.
type Item = list.Item

// LoadResult is sent when async loading completes.
type LoadResult struct {
	Items []Item
	Err   error
}

// Model is a list that loads items asynchronously.
type Model struct {
	theme    styles.Theme
	list     *list.Model
	loader   *loader.Model
	loadFunc tea.Cmd
	state    state
	err      error
}

// New creates a new remotelist.
func New(theme styles.Theme, loadingMessage string) *Model {
	return &Model{
		theme:  theme,
		list:   list.New(theme, []list.Item{}),
		loader: loader.New(theme, loadingMessage),
		state:  stateLoading,
	}
}

// Init starts the loader animation.
func (m *Model) Init() tea.Cmd {
	return m.loader.Init()
}

// InitWithLoader starts loading with the given load function.
func (m *Model) InitWithLoader(loadFunc tea.Cmd) tea.Cmd {
	m.loadFunc = loadFunc
	return tea.Batch(m.loader.Init(), loadFunc)
}

// Retry resets error state and retries loading.
func (m *Model) Retry() tea.Cmd {
	if m.loadFunc == nil {
		return nil
	}
	m.state = stateLoading
	m.err = nil
	return tea.Batch(m.loader.Init(), m.loadFunc)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case LoadResult:
		if msg.Err != nil {
			m.state = stateError
			m.err = msg.Err
			return nil
		}
		m.state = stateLoaded

		itemCount := len(msg.Items)
		listHeight := itemCount
		if m.list.FilteringEnabled() {
			listHeight++
		}
		if listHeight > list.MaxHeight {
			listHeight = list.MaxHeight
		}
		if listHeight > 0 {
			m.list.SetHeight(listHeight)
		}

		m.list.SetShowPagination(itemCount > list.MaxHeight)
		return m.list.SetItems(msg.Items)
	}

	switch m.state {
	case stateLoading:
		return m.loader.Update(msg)
	case stateLoaded:
		return m.list.Update(msg)
	case stateError:
		return nil
	}

	return nil
}

// View renders the component.
func (m *Model) View() string {
	switch m.state {
	case stateLoading:
		return m.loader.View()
	case stateError:
		return ""
	case stateLoaded:
		return m.list.View()
	}
	return ""
}

// SetWidth sets the list width.
func (m *Model) SetWidth(width int) {
	m.list.SetWidth(width)
}

// SetHeight sets the list height.
func (m *Model) SetHeight(height int) {
	m.list.SetHeight(height)
}

// SelectedItem returns the currently selected item.
func (m *Model) SelectedItem() list.Item {
	return m.list.SelectedItem()
}

// IsLoaded returns true if loading completed (success or error).
func (m *Model) IsLoaded() bool {
	return m.state == stateLoaded || m.state == stateError
}

// IsLoading returns true while loading is in progress.
func (m *Model) IsLoading() bool {
	return m.state == stateLoading
}

// HasError returns true if loading failed.
func (m *Model) HasError() bool {
	return m.state == stateError
}

// Error returns the loading error, if any.
func (m *Model) Error() error {
	if m.state == stateError {
		return m.err
	}
	return nil
}

// KeyMap returns the list key bindings.
func (m *Model) KeyMap() list.KeyMap {
	return m.list.KeyMap()
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.state != stateLoaded {
		return nil
	}
	return m.list.ShortHelp()
}
