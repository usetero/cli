package selectlist

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	baselist "github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type optionItem struct {
	option core.Option
}

func (i optionItem) FilterValue() string { return i.option.Label + " " + i.option.Subtitle }
func (i optionItem) Title() string       { return i.option.Label }
func (i optionItem) Subtitle() string    { return i.option.Subtitle }

type Model struct {
	list *baselist.Model
}

var _ core.Model = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{list: baselist.New(appTheme, false, "")}
}

func (m *Model) Init() tea.Cmd { return m.list.Init() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.list.Update(msg)
	if typed, ok := msg.(baselist.SelectedMsg); ok {
		selected, ok := typed.Item.(optionItem)
		if !ok {
			return m, cmd
		}
		return m, func() tea.Msg { return SelectedMsg{Option: selected.option} }
	}
	if updated, ok := next.(*baselist.Model); ok {
		m.list = updated
	}
	return m, cmd
}

func (m *Model) ConsumesKey(msg tea.KeyPressMsg) bool {
	return m.list.ConsumesKey(msg)
}

func (m *Model) SetOptions(options []core.Option) {
	items := make([]list.Item, 0, len(options))
	for _, option := range options {
		items = append(items, optionItem{option: option})
	}
	m.list.SetItems(items)
}

func (m *Model) ApplyInput(input core.Input) {
	m.SetOptions(input.Options)
}

func (m *Model) PreferredHeight(width int) int {
	provider, ok := any(m.list).(core.HeightProvider)
	if !ok {
		return 1
	}
	return provider.PreferredHeight(width)
}
