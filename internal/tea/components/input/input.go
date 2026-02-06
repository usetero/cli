// Package input provides a themed text input component.
package input

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/cursor"
)

// Model wraps textinput.Model with themed defaults.
type Model struct {
	theme *styles.Theme
	input textinput.Model
}

// New creates a new themed text input.
func New(theme *styles.Theme) *Model {
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

	return &Model{theme: theme, input: ti}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

// View renders the input with cursor marker.
func (m *Model) View() string {
	view := m.input.View()
	cur := m.input.Cursor()

	if cur != nil && cur.X >= 0 && cur.X <= len(view) {
		view = view[:cur.X] + cursor.Marker + view[cur.X:]
	}

	return view
}

// Value returns the current input value.
func (m *Model) Value() string {
	return m.input.Value()
}

// SetValue sets the input value.
func (m *Model) SetValue(s string) {
	m.input.SetValue(s)
}

// SetPlaceholder sets the placeholder text.
func (m *Model) SetPlaceholder(placeholder string) {
	m.input.Placeholder = placeholder
}

// SetCharLimit sets the maximum character limit.
func (m *Model) SetCharLimit(limit int) {
	m.input.CharLimit = limit
}

// SetWidth sets the input width.
func (m *Model) SetWidth(width int) {
	m.input.SetWidth(width)
}

// SetEchoMode sets the echo mode (normal, password, none).
func (m *Model) SetEchoMode(mode textinput.EchoMode) {
	m.input.EchoMode = mode
}

// SetEchoCharacter sets the character used for password mode.
func (m *Model) SetEchoCharacter(char rune) {
	m.input.EchoCharacter = char
}

// Focus focuses the input.
func (m *Model) Focus() tea.Cmd {
	return m.input.Focus()
}

// Blur removes focus from the input.
func (m *Model) Blur() {
	m.input.Blur()
}

// Focused returns whether the input is focused.
func (m *Model) Focused() bool {
	return m.input.Focused()
}

// Reset clears the input value.
func (m *Model) Reset() {
	m.input.Reset()
}

// ShortHelp returns the key bindings for the short help view.
// Text inputs don't expose editing keybindings in help; parent defines submit.
func (m *Model) ShortHelp() []key.Binding {
	return nil
}
