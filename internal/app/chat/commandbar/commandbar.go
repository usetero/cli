package commandbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/cursor"
)

// Model handles user input via a textarea.
type Model struct {
	theme    *styles.Theme
	textarea textarea.Model
	width    int
}

// New creates a new command bar.
func New(theme *styles.Theme, width int) *Model {
	colors := theme.Colors

	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.CharLimit = -1
	ta.SetVirtualCursor(false)
	ta.SetWidth(width)
	ta.Focus()

	base := lipgloss.NewStyle().Foreground(colors.Page.Text)
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:        base,
			Text:        base,
			Placeholder: base.Foreground(colors.Page.TextMuted),
			Prompt:      base.Foreground(colors.Accent),
		},
		Blurred: textarea.StyleState{
			Base:        base.Foreground(colors.Page.TextMuted),
			Text:        base.Foreground(colors.Page.TextMuted),
			Placeholder: base.Foreground(colors.Page.TextMuted),
			Prompt:      base.Foreground(colors.Page.TextMuted),
		},
		Cursor: textarea.CursorStyle{
			Color: colors.Accent,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})

	ta.SetPromptFunc(4, func(info textarea.PromptInfo) string {
		if info.LineNumber == 0 {
			if info.Focused {
				return "  > "
			}
			return "::: "
		}
		if info.Focused {
			return lipgloss.NewStyle().Foreground(colors.Accent).Render("::: ")
		}
		return lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render("::: ")
	})

	return &Model{
		theme:    theme,
		textarea: ta,
		width:    width,
	}
}

// Init initializes the command bar.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.textarea.Focus(),
		textarea.Blink,
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Shift+enter for newline - consume, don't forward to textarea
		if msg.String() == "shift+enter" || msg.String() == "ctrl+j" {
			m.textarea.InsertRune('\n')
			return nil
		}
		// Enter to submit - consume, don't forward to textarea
		if msg.String() == "enter" {
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				m.textarea.Reset()
				return func() tea.Msg { return msgs.UserSubmittedInput{Text: text} }
			}
			return nil
		}
	}

	// Forward to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return cmd
}

// View renders the command bar.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	view := m.textarea.View()

	// Insert cursor marker
	cur := m.textarea.Cursor()
	if cur != nil {
		view = cursor.Insert(view, cur.X, cur.Y)
	}

	return lipgloss.NewStyle().
		Padding(1, 0).
		Render(view)
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.textarea.SetWidth(width)
}

// Height returns the height of the command bar.
func (m *Model) Height() int {
	return 5 // 3 lines + 2 padding
}

// Focus returns a command to focus the textarea.
func (m *Model) Focus() tea.Cmd {
	return m.textarea.Focus()
}

// Blur removes focus from the textarea.
func (m *Model) Blur() {
	m.textarea.Blur()
}

// Focused returns whether the textarea is focused.
func (m *Model) Focused() bool {
	return m.textarea.Focused()
}

// KeyBindings returns the key bindings for the command bar.
func (m *Model) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
		key.NewBinding(key.WithKeys("shift+enter"), key.WithHelp("shift+enter", "newline")),
	}
}
