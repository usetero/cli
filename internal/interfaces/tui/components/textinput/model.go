package textinput

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var (
	backspaceBinding = key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete"),
	)
	typeHelpBinding = key.NewBinding(
		key.WithKeys("type"),
		key.WithHelp("type", "edit"),
	)
)

// Model owns single-line text editing state.
type Model struct {
	theme       theme.Theme
	value       string
	placeholder string
}

// New constructs a text input model.
func New(theme theme.Theme) *Model {
	return &Model{
		theme:       theme,
		placeholder: "Type here...",
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of shared contracts. Text input currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// Value returns current field text.
func (m *Model) Value() string { return m.value }

// SetValue sets current field text.
func (m *Model) SetValue(v string) { m.value = v }

// SetPlaceholder changes placeholder text shown when input is empty.
func (m *Model) SetPlaceholder(placeholder string) { m.placeholder = placeholder }

// Reset clears the input.
func (m *Model) Reset() { m.value = "" }

// ShortHelp returns text editing key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{typeHelpBinding, backspaceBinding}
}

// Update handles text editing keys.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, backspaceBinding) {
		runes := []rune(m.value)
		if len(runes) > 0 {
			m.value = string(runes[:len(runes)-1])
		}
		return m, nil
	}
	if keyMsg.Text != "" {
		m.value += keyMsg.Text
	}
	return m, nil
}

// View satisfies tea.Model; wrappers usually render their own labeled lines.
func (m *Model) View() tea.View {
	if m.value == "" {
		return tea.NewView(m.theme.Input.Placeholder.Render(m.placeholder))
	}
	return tea.NewView(m.theme.Input.Value.Render(m.value))
}
