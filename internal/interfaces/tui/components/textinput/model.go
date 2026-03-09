package textinput

import (
	"charm.land/bubbles/v2/key"
	bubbleinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/cursor"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model owns single-line text editing state.
type Model struct {
	theme theme.Theme
	input bubbleinput.Model
}

// New constructs a themed text input model.
func New(appTheme theme.Theme) *Model {
	ti := bubbleinput.New()
	ti.SetVirtualCursor(false)
	ti.Prompt = ""
	ti.CharLimit = 256
	ti.Focus()
	ti.SetStyles(bubbleinput.Styles{
		Focused: bubbleinput.StyleState{
			Text: lipgloss.NewStyle().
				Foreground(appTheme.TextColor).
				Background(appTheme.Background),
			Placeholder: lipgloss.NewStyle().
				Foreground(appTheme.TextSubtle).
				Background(appTheme.Background),
			Prompt: lipgloss.NewStyle().
				Foreground(appTheme.Accent).
				Background(appTheme.Background),
		},
		Blurred: bubbleinput.StyleState{
			Text: lipgloss.NewStyle().
				Foreground(appTheme.TextMuted).
				Background(appTheme.Background),
			Placeholder: lipgloss.NewStyle().
				Foreground(appTheme.TextSubtle).
				Background(appTheme.Background),
			Prompt: lipgloss.NewStyle().
				Foreground(appTheme.TextSubtle).
				Background(appTheme.Background),
		},
		Cursor: bubbleinput.CursorStyle{
			Color: appTheme.AccentAlt,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})

	return &Model{
		theme: appTheme,
		input: ti,
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of shared contracts.
func (m *Model) SetSize(width, _ int) {
	if width > 0 {
		m.input.SetWidth(width)
	}
}

// Value returns current field text.
func (m *Model) Value() string { return m.input.Value() }

// SetValue sets current field text.
func (m *Model) SetValue(v string) { m.input.SetValue(v) }

// SetPlaceholder changes placeholder text shown when input is empty.
func (m *Model) SetPlaceholder(placeholder string) { m.input.Placeholder = placeholder }

// SetWidth updates the input width.
func (m *Model) SetWidth(width int) { m.input.SetWidth(width) }

// SetCharLimit changes the character limit.
func (m *Model) SetCharLimit(limit int) { m.input.CharLimit = limit }

// SetEchoMode changes echo behavior.
func (m *Model) SetEchoMode(mode bubbleinput.EchoMode) { m.input.EchoMode = mode }

// SetEchoCharacter changes the password echo character.
func (m *Model) SetEchoCharacter(char rune) { m.input.EchoCharacter = char }

// Focus focuses the input.
func (m *Model) Focus() tea.Cmd { return m.input.Focus() }

// Blur removes focus from the input.
func (m *Model) Blur() { m.input.Blur() }

// Focused reports whether the input has focus.
func (m *Model) Focused() bool { return m.input.Focused() }

// Reset clears the input.
func (m *Model) Reset() { m.input.Reset() }

// ShortHelp returns text editing key bindings. Parent screens own submit/help policy.
func (m *Model) ShortHelp() []key.Binding { return nil }

// Update handles text editing.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the text input.
func (m *Model) View() tea.View {
	if m.input.Value() == "" {
		view := m.theme.Input.Placeholder.Render(m.input.Placeholder)
		if m.input.Focused() {
			view = cursor.Insert(view, 0, 0)
		}
		return tea.NewView(view)
	}
	view := m.input.View()
	if cur := m.input.Cursor(); cur != nil {
		view = cursor.Insert(view, cur.X, 0)
	}
	return tea.NewView(view)
}
