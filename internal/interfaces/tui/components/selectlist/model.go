package selectlist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var (
	upBinding = key.NewBinding(
		key.WithKeys("up", "k"),
	)
	downBinding = key.NewBinding(
		key.WithKeys("down", "j"),
	)
	selectBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "continue"),
	)
	moveHelpBinding = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("up/down", "move"),
	)
)

// Item is one selectable row in the list.
type Item struct {
	Title    string
	Subtitle string
}

// SelectedMsg indicates the current cursor index was confirmed.
type SelectedMsg struct {
	Index int
}

// Model owns list cursor state and selection key handling.
type Model struct {
	theme  theme.Theme
	items  []Item
	cursor int
}

// New constructs a select list model.
func New(theme theme.Theme) *Model {
	return &Model{theme: theme}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the shared screen contract. Select list currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// SetItems replaces options and applies selected index if valid.
func (m *Model) SetItems(items []Item, selected int) {
	m.items = append([]Item(nil), items...)
	m.cursor = 0
	if selected >= 0 && selected < len(m.items) {
		m.cursor = selected
	}
}

// SelectedIndex returns the cursor index, or -1 if no options exist.
func (m *Model) SelectedIndex() int {
	if len(m.items) == 0 {
		return -1
	}
	return m.cursor
}

// ShortHelp returns list key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{moveHelpBinding, selectBinding}
}

// Update handles cursor movement and selection.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, upBinding):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, downBinding):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, selectBinding):
		index := m.SelectedIndex()
		if index < 0 {
			return m, nil
		}
		return m, func() tea.Msg { return SelectedMsg{Index: index} }
	}
	return m, nil
}

// View renders only the rows for this list.
func (m *Model) View() tea.View {
	if len(m.items) == 0 {
		return tea.NewView(m.theme.List.Empty.Render("No options available."))
	}
	lines := make([]string, 0, len(m.items)*2)
	for i := range m.items {
		prefix := m.theme.List.CursorInactive.Render("  ")
		title := m.theme.List.Item.Render(m.items[i].Title)
		subtitle := m.theme.List.Subtitle.Render(m.items[i].Subtitle)
		if i == m.cursor {
			prefix = m.theme.List.Cursor.Render("> ")
			title = m.theme.List.ItemActive.Render(m.items[i].Title)
			subtitle = m.theme.List.SubtitleActive.Render(m.items[i].Subtitle)
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, prefix, title))
		if m.items[i].Subtitle != "" {
			lines = append(lines, "  "+subtitle)
		}
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
