package selectlist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

var (
	upBinding = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move"),
	)
	downBinding = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move"),
	)
	selectBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	)
)

// Model owns local command-option selection state.
type Model struct {
	theme   theme.Theme
	width   int
	options []core.Option
	index   int
}

var _ core.Model = (*Model)(nil)

// New constructs a select-list model.
func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update handles local list navigation and selection.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, upBinding):
		if m.index > 0 {
			m.index--
		}
	case key.Matches(keyMsg, downBinding):
		if m.index < len(m.options)-1 {
			m.index++
		}
	case key.Matches(keyMsg, selectBinding):
		if len(m.options) == 0 {
			return m, nil
		}
		option := m.options[m.index]
		return m, func() tea.Msg { return SelectedMsg{Option: option} }
	}

	return m, nil
}

// SetOptions updates the current option list.
func (m *Model) SetOptions(options []core.Option) {
	m.options = append([]core.Option(nil), options...)
	if len(m.options) == 0 {
		m.index = 0
		return
	}
	if m.index >= len(m.options) {
		m.index = len(m.options) - 1
	}
}

// ApplySpec updates the current option list from a select-list spec.
func (m *Model) ApplyInput(input core.Input) {
	m.SetOptions(input.Options)
}
