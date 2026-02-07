package inputbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/cursor"
	"github.com/usetero/cli/internal/tea/keymap"
)

const (
	textareaHeight  = 3                                // visible input lines
	verticalPadding = 2                                // 1 top + 1 bottom
	inputBarHeight  = textareaHeight + verticalPadding // total height
)

// Model handles user input via a textarea.
type Model struct {
	theme    *styles.Theme
	textarea textarea.Model
	width    int
	scope    log.Scope
}

// New creates a new input bar.
func New(theme *styles.Theme, scope log.Scope) *Model {
	scope = scope.Child("inputbar")
	colors := theme.Colors

	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.SetHeight(textareaHeight)
	ta.CharLimit = -1
	ta.SetVirtualCursor(false)
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
		scope:    scope,
	}
}

// Init initializes the input bar.
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
		// DEBUG: log all keys
		m.scope.Debug("key received", "key", msg.String())

		// Newline - consume, don't forward to textarea
		if key.Matches(msg, keymap.Newline) {
			m.textarea.InsertRune('\n')
			return nil
		}
		// Enter to submit - consume, don't forward to textarea
		if key.Matches(msg, keymap.Send) {
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

// View renders the input bar.
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

	// Pad textarea output to exactly textareaHeight lines
	lines := strings.Split(view, "\n")
	for len(lines) < textareaHeight {
		lines = append(lines, "")
	}
	view = strings.Join(lines[:textareaHeight], "\n")

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 0).
		Render(view)
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.textarea.SetWidth(width)
}

// Height returns the height of the input bar.
func (m *Model) Height() int {
	return inputBarHeight
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

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keymap.Send, keymap.Newline}
}
