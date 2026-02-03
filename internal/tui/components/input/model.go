package input

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/cursor"
)

// Model wraps textinput.Model with themed defaults.
type Model struct {
	theme  *styles.Theme
	logger log.Logger
	model  textinput.Model
}

// New creates a new input model.
func New(theme *styles.Theme, logger log.Logger) Model {
	colors := theme.Colors

	ti := textinput.New()
	ti.SetVirtualCursor(false)
	ti.Prompt = "> "
	ti.CharLimit = 256
	ti.Focus()

	ti.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(colors.Input.Text),
			Placeholder: lipgloss.NewStyle().Foreground(colors.Input.Placeholder),
			Prompt:      lipgloss.NewStyle().Foreground(colors.Accent),
		},
		Blurred: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
			Placeholder: lipgloss.NewStyle().Foreground(colors.Input.Placeholder),
			Prompt:      lipgloss.NewStyle().Foreground(colors.Page.TextMuted),
		},
		Cursor: textinput.CursorStyle{
			Color: colors.Accent,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})

	return Model{theme: theme, logger: logger, model: ti}
}

// Init initializes the input.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

// View renders the input with cursor marker.
func (m Model) View() string {
	view := m.model.View()
	cur := m.model.Cursor()

	if cur != nil && cur.X >= 0 && cur.X <= len(view) {
		view = view[:cur.X] + cursor.Marker + view[cur.X:]
	}

	return view
}

// SetPlaceholder returns a new Model with the given placeholder.
func (m Model) SetPlaceholder(placeholder string) Model {
	m.model.Placeholder = placeholder
	return m
}

// SetCharLimit returns a new Model with the given char limit.
func (m Model) SetCharLimit(limit int) Model {
	m.model.CharLimit = limit
	return m
}

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.model.SetWidth(width)
	return m
}

// SetEchoMode returns a new Model with the given echo mode.
func (m Model) SetEchoMode(mode textinput.EchoMode) Model {
	m.model.EchoMode = mode
	return m
}

// SetEchoCharacter returns a new Model with the given echo character.
func (m Model) SetEchoCharacter(char rune) Model {
	m.model.EchoCharacter = char
	return m
}

// Focus focuses the input and returns a command.
func (m Model) Focus() tea.Cmd {
	return m.model.Focus()
}

// Value returns the current input value.
func (m Model) Value() string {
	return m.model.Value()
}
