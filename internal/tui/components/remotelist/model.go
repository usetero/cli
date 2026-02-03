package remotelist

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/loader"
)

type state int

const (
	stateLoading state = iota
	stateLoaded
	stateError
)

// LoadResultMsg is the message sent when loading completes.
type LoadResultMsg struct {
	Items []list.Item
	Err   error
}

// Model is a list that loads items asynchronously.
type Model struct {
	theme      *styles.Theme
	logger     log.Logger
	list       list.Model
	loader     loader.Model
	loaderFunc tea.Cmd
	state      state
	err        error
}

// New creates a new remotelist model.
func New(theme *styles.Theme, delegate list.ItemDelegate, loadingMessage string, logger log.Logger) Model {
	return Model{
		theme:  theme,
		logger: logger,
		list:   list.New(theme, []list.Item{}, delegate),
		loader: loader.New(theme, loadingMessage),
		state:  stateLoading,
	}
}

// Init starts the loader animation.
func (m Model) Init() tea.Cmd {
	return m.loader.Init()
}

// InitWithLoader starts loading with the given load function.
func (m Model) InitWithLoader(loadFunc tea.Cmd) (Model, tea.Cmd) {
	m.loaderFunc = loadFunc
	return m, tea.Batch(m.loader.Init(), loadFunc)
}

// Retry resets error state and retries loading.
func (m Model) Retry() (Model, tea.Cmd) {
	if m.loaderFunc == nil {
		return m, nil
	}
	m.logger.Debug("retrying load")
	m.state = stateLoading
	m.err = nil
	return m, tea.Batch(m.loader.Init(), m.loaderFunc)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadResultMsg:
		if msg.Err != nil {
			m.state = stateError
			m.err = msg.Err
			m.logger.Error("failed to load list items", "error", msg.Err)
			return m, nil
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
			m.list = m.list.SetHeight(listHeight)
		}

		m.list = m.list.SetShowPagination(itemCount > list.MaxHeight)

		var cmd tea.Cmd
		m.list, cmd = m.list.SetItems(msg.Items)
		return m, cmd
	}

	switch m.state {
	case stateLoading:
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		return m, cmd
	case stateLoaded:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case stateError:
		return m, nil
	}

	return m, nil
}

// View renders the component.
func (m Model) View() string {
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

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.list = m.list.SetWidth(width)
	return m
}

// SetHeight returns a new Model with the given height.
func (m Model) SetHeight(height int) Model {
	m.list = m.list.SetHeight(height)
	return m
}

// SelectedItem returns the currently selected item.
func (m Model) SelectedItem() list.Item {
	return m.list.SelectedItem()
}

// IsLoaded returns true if loading completed.
func (m Model) IsLoaded() bool {
	return m.state == stateLoaded || m.state == stateError
}

// IsBusy returns true while loading.
func (m Model) IsBusy() bool {
	return m.state == stateLoading
}

// HasError returns true if loading failed.
func (m Model) HasError() bool {
	return m.state == stateError
}

// Error returns the current error.
func (m Model) Error() error {
	if m.state == stateError {
		return m.err
	}
	return nil
}

// KeyMap returns the list key bindings.
func (m Model) KeyMap() list.KeyMap {
	return m.list.KeyMap()
}
