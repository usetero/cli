package selectlist

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
		key.WithHelp("enter", "confirm"),
	)
	moveHelpBinding = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "select"),
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
	theme     theme.Theme
	items     []Item
	cursor    int
	width     int
	emptyText string
}

// New constructs a select list model.
func New(theme theme.Theme) *Model {
	return &Model{
		theme:     theme,
		emptyText: "No options available.",
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the shared screen contract.
func (m *Model) SetSize(width, _ int) {
	if width > 0 {
		m.width = width
	}
}

// SetEmptyText replaces the default empty-state copy.
func (m *Model) SetEmptyText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "No options available."
	}
	m.emptyText = text
}

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
		return tea.NewView(m.theme.List.Empty.Render(m.emptyText))
	}
	lines := make([]string, 0, len(m.items)*2)
	for i := range m.items {
		border := m.theme.List.CursorInactive.Render("│")
		prefix := m.theme.List.CursorInactive.Render("  ")
		titleStyle := m.theme.List.Item
		subtitleStyle := m.theme.List.Subtitle
		if i == m.cursor {
			border = m.theme.List.Cursor.Render("┃")
			prefix = m.theme.List.Cursor.Render("> ")
			titleStyle = m.theme.List.ItemActive
			subtitleStyle = m.theme.List.SubtitleActive
		}

		contentWidth := m.width - 4
		if contentWidth > 0 {
			titleStyle = titleStyle.MaxWidth(contentWidth)
			subtitleStyle = subtitleStyle.MaxWidth(contentWidth)
		}
		lines = append(lines, border+" "+prefix+titleStyle.Render(m.items[i].Title))
		if m.items[i].Subtitle != "" {
			lines = append(lines, border+"   "+subtitleStyle.Render(m.items[i].Subtitle))
		}
	}
	return tea.NewView(strings.Join(lines, "\n"))
}
