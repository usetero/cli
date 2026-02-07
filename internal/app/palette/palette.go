// Package palette provides a command palette overlay for the TUI.
// Triggered by "/" in the input bar, it shows a fuzzy-filtered list of commands.
package palette

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// Command is a single palette entry.
type Command struct {
	Name    string         // Display name (e.g. "New Conversation")
	Handler func() tea.Cmd // What to do when selected
}

// OpenMsg requests the palette to open.
type OpenMsg struct{}

// CloseMsg is sent when the palette closes itself.
type CloseMsg struct{}

// Key bindings.
var (
	selectKey = key.NewBinding(key.WithKeys("enter"))
	closeKey  = key.NewBinding(key.WithKeys("esc"))
	nextKey   = key.NewBinding(key.WithKeys("down", "ctrl+n"))
	prevKey   = key.NewBinding(key.WithKeys("up", "ctrl+p"))
)

const (
	maxVisible   = 8 // max items shown
	headerHeight = 1
	inputHeight  = 1
	paddingH     = 1
	itemPaddingH = 1
	diag         = "╱"
)

// Model is the command palette.
type Model struct {
	theme    *styles.Theme
	input    *input.Model
	commands []Command // all commands
	matches  []match   // filtered results
	selected int       // index into matches
	width    int       // available width (set by caller)
}

// match pairs a command with its fuzzy match positions.
type match struct {
	command      Command
	matchIndexes []int
}

// New creates a new command palette.
func New(theme *styles.Theme, commands []Command) *Model {
	ti := input.New(theme)
	ti.SetPlaceholder("Type a command...")
	ti.SetCharLimit(128)

	m := &Model{
		theme:    theme,
		input:    ti,
		commands: commands,
	}
	m.filter()
	return m
}

// Init returns the initial command (focus + blink).
func (m *Model) Init() tea.Cmd {
	return m.input.Focus()
}

// Update handles key events. The caller must only forward events when the palette is open.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, closeKey):
			return func() tea.Msg { return CloseMsg{} }

		case key.Matches(msg, selectKey):
			if len(m.matches) > 0 {
				handler := m.matches[m.selected].command.Handler
				return tea.Sequence(
					func() tea.Msg { return CloseMsg{} },
					handler(),
				)
			}
			return nil

		case key.Matches(msg, nextKey):
			if m.selected < len(m.matches)-1 {
				m.selected++
			}
			return nil

		case key.Matches(msg, prevKey):
			if m.selected > 0 {
				m.selected--
			}
			return nil
		}
	}

	// Forward to text input for typing
	prevValue := m.input.Value()
	cmd := m.input.Update(msg)
	if m.input.Value() != prevValue {
		m.filter()
	}
	return cmd
}

// filter updates the match list based on current input.
func (m *Model) filter() {
	query := m.input.Value()

	if query == "" {
		// Show all commands
		m.matches = make([]match, len(m.commands))
		for i, cmd := range m.commands {
			m.matches[i] = match{command: cmd}
		}
	} else {
		// Fuzzy match
		source := commandSource(m.commands)
		results := fuzzy.FindFrom(query, source)
		m.matches = make([]match, len(results))
		for i, r := range results {
			m.matches[i] = match{
				command:      m.commands[r.Index],
				matchIndexes: r.MatchedIndexes,
			}
		}
	}

	// Clamp selection
	if m.selected >= len(m.matches) {
		m.selected = max(0, len(m.matches)-1)
	}
}

// SetWidth sets the palette width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.SetWidth(width - 2*paddingH - 2) // 2 for "> " prompt
}

// Height returns the rendered height of the palette.
func (m *Model) Height() int {
	visible := min(len(m.matches), maxVisible)
	// border(2) + header + gap + input + separator + items
	return 2 + headerHeight + 1 + inputHeight + 1 + visible
}

// View renders the palette.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	colors := m.theme.Colors
	innerWidth := m.width - 2 // borders
	contentWidth := innerWidth - 2*paddingH

	// Header: "Commands ╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱"
	header := m.renderHeader(contentWidth)

	// Input line (cursor marker inserted by input component)
	inputView := m.input.View()

	// Items
	visible := min(len(m.matches), maxVisible)
	var items []string
	for i := range visible {
		items = append(items, m.renderItem(i, contentWidth))
	}

	// Separator between input and items
	sep := lipgloss.NewStyle().
		Foreground(colors.BorderDefault).
		Render(strings.Repeat("─", contentWidth))

	// Stack: header + gap + input + separator + items
	var sections []string
	sections = append(sections, header, "", inputView)
	if len(items) > 0 {
		sections = append(sections, sep)
		sections = append(sections, items...)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(colors.Page.Bg).
		Foreground(colors.Page.Text).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Accent).
		BorderBackground(colors.Page.Bg).
		Padding(0, paddingH).
		Render(content)
}

// renderHeader renders "Commands ╱╱╱╱╱╱╱╱" with gradient slashes.
func (m *Model) renderHeader(width int) string {
	colors := m.theme.Colors
	title := lipgloss.NewStyle().Foreground(colors.Accent).Bold(true).Render("Commands")
	titleWidth := lipgloss.Width(title)

	remaining := width - titleWidth - 1 // 1 for space
	if remaining <= 0 {
		return title
	}

	slashes := strings.Repeat(diag, remaining)
	slashes = styles.ApplyForegroundGrad(slashes, colors.Accent, colors.AccentAlt)

	return title + " " + slashes
}

// renderItem renders a single list item.
func (m *Model) renderItem(index, width int) string {
	colors := m.theme.Colors
	item := m.matches[index]
	name := item.command.Name
	isSelected := index == m.selected

	// Truncate name to fit
	name = ansi.Truncate(name, width, "…")

	if isSelected {
		// Highlight matched characters in selected item
		styled := m.highlightMatches(name, item.matchIndexes,
			lipgloss.NewStyle().Foreground(colors.Page.Bg).Background(colors.Accent).Bold(true),
			lipgloss.NewStyle().Foreground(colors.Page.Bg).Background(colors.Accent),
		)
		return lipgloss.NewStyle().
			Width(width).
			Background(colors.Accent).
			Foreground(colors.Page.Bg).
			Render(styled)
	}

	// Highlight matched characters in normal item
	styled := m.highlightMatches(name, item.matchIndexes,
		lipgloss.NewStyle().Foreground(colors.Accent).Background(colors.Page.Bg).Bold(true),
		lipgloss.NewStyle().Foreground(colors.Page.Text).Background(colors.Page.Bg),
	)
	return lipgloss.NewStyle().
		Width(width).
		Background(colors.Page.Bg).
		Foreground(colors.Page.Text).
		Render(styled)
}

// highlightMatches applies matchStyle to matched rune positions and baseStyle to the rest.
func (m *Model) highlightMatches(text string, indexes []int, matchStyle, baseStyle lipgloss.Style) string {
	if len(indexes) == 0 {
		return baseStyle.Render(text)
	}

	indexSet := make(map[int]bool, len(indexes))
	for _, idx := range indexes {
		indexSet[idx] = true
	}

	var b strings.Builder
	for i, r := range text {
		if indexSet[i] {
			b.WriteString(matchStyle.Render(string(r)))
		} else {
			b.WriteString(baseStyle.Render(string(r)))
		}
	}
	return b.String()
}

// commandSource implements fuzzy.Source for a slice of commands.
type commandSource []Command

func (s commandSource) String(i int) string { return s[i].Name }
func (s commandSource) Len() int            { return len(s) }
