package input

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

const visibleLines = 3

var (
	sendBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	)
	newlineBinding = key.NewBinding(
		key.WithKeys("ctrl+j"),
		key.WithHelp("ctrl+j", "newline"),
	)
)

// Model owns the command bar textarea state.
type Model struct {
	theme       theme.Theme
	textarea    textarea.Model
	width       int
	placeholder string
}

var _ core.Model = (*Model)(nil)

// New constructs a command bar input model.
func New(appTheme theme.Theme) *Model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.SetHeight(visibleLines)
	ta.CharLimit = -1
	ta.SetVirtualCursor(false)
	ta.Focus()
	ta.Placeholder = ""
	ta.SetPromptFunc(0, func(textarea.PromptInfo) string { return "" })

	base := lipgloss.NewStyle().
		Foreground(appTheme.Palette.Text).
		Background(appTheme.Background)
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:        base,
			Text:        base,
			Placeholder: appTheme.Input.Placeholder,
			Prompt:      appTheme.Input.Active,
		},
		Blurred: textarea.StyleState{
			Base:        appTheme.Text.Muted,
			Text:        appTheme.Text.Muted,
			Placeholder: appTheme.Input.Placeholder,
			Prompt:      appTheme.Input.Inactive,
		},
		Cursor: textarea.CursorStyle{
			Color: appTheme.Palette.AccentAlt,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})

	return &Model{
		theme:       appTheme,
		textarea:    ta,
		placeholder: "What's on your mind?",
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.textarea.Focus(), textarea.Blink)
}

// Update handles textarea interaction.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, newlineBinding):
			m.textarea.InsertRune('\n')
			return m, nil
		case key.Matches(keyMsg, sendBinding):
			text := strings.TrimSpace(m.textarea.Value())
			if text == "" {
				return m, nil
			}
			m.textarea.Reset()
			return m, func() tea.Msg { return SubmittedMsg{Text: text} }
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// SetPlaceholder updates the visible placeholder copy.
func (m *Model) SetPlaceholder(placeholder string) {
	if strings.TrimSpace(placeholder) == "" {
		placeholder = "What's on your mind?"
	}
	m.placeholder = placeholder
	m.textarea.Placeholder = placeholder
}

// ApplySpec updates the input state from an input spec.
func (m *Model) ApplyInput(input core.Input) {
	m.SetPlaceholder(input.Placeholder)
}
