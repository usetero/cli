package selectlist

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

const (
	filterThreshold = 4
	maxVisibleRows  = 5
)

var selectBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select"),
)

type Row interface {
	list.Item
	Title() string
	Subtitle() string
}

type Model struct {
	baseFilter bool
	list       list.Model
}

var _ core.Model = (*Model)(nil)

func New(appTheme theme.Theme, filterable bool, placeholder string) *Model {
	l := newList(appTheme)
	return &Model{
		baseFilter: filterable,
		list:       l,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, selectBinding) {
		if selected := m.list.SelectedItem(); selected != nil {
			return m, func() tea.Msg { return SelectedMsg{Item: selected} }
		}
	}

	return m, cmd
}

func (m *Model) SetItems(items []list.Item) {
	_ = m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
	}

	filtering := m.baseFilter && len(items) > filterThreshold
	m.list.SetFilteringEnabled(filtering)
	m.list.SetShowFilter(filtering)
	m.list.SetHeight(m.preferredHeight())
}

func (m *Model) SetSize(width, height int) {
	if width < 1 {
		width = 1
	}
	m.list.SetWidth(width)
	if height > 0 {
		m.list.SetHeight(height)
		return
	}
	m.list.SetHeight(m.preferredHeight())
}

func (m *Model) View() tea.View {
	return tea.NewView(m.list.View())
}

func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding(nil), m.list.ShortHelp()...)
	for _, binding := range bindings {
		for _, name := range binding.Keys() {
			if name == "enter" {
				return bindings
			}
		}
	}
	return append(bindings, selectBinding)
}

func (m *Model) ConsumesKey(msg tea.KeyPressMsg) bool {
	if m.list.SettingFilter() {
		return true
	}
	if key.Matches(msg, selectBinding) {
		return true
	}
	for _, binding := range m.list.ShortHelp() {
		if key.Matches(msg, binding) {
			return true
		}
	}
	return false
}

func (m *Model) preferredHeight() int {
	rows := len(m.list.Items())
	if rows < 1 {
		rows = 1
	}
	if rows > maxVisibleRows {
		rows = maxVisibleRows
	}
	if m.list.ShowFilter() {
		rows++
	}
	return rows
}

func SelectBinding() key.Binding {
	return selectBinding
}

func (m *Model) PreferredHeight(int) int {
	return m.preferredHeight()
}
