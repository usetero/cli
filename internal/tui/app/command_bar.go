package app

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/messages"
	"github.com/usetero/cli/internal/tui/cursor"
)

// CommandBar is the input component for chat.
type CommandBar struct {
	theme    *styles.Theme
	textarea textarea.Model
	width    int
}

// NewCommandBar creates a new command bar.
func NewCommandBar(theme *styles.Theme) CommandBar {
	colors := theme.Colors

	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
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
			return "> "
		}
		return "  "
	})

	return CommandBar{
		theme:    theme,
		textarea: ta,
	}
}

// Init initializes the command bar.
func (m CommandBar) Init() tea.Cmd {
	return tea.Batch(
		m.textarea.Focus(),
		textarea.Blink,
	)
}

// Update handles messages.
func (m CommandBar) Update(msg tea.Msg) (CommandBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Shift+enter for newline
		if msg.String() == "shift+enter" || msg.String() == "ctrl+j" {
			m.textarea.InsertRune('\n')
			return m, nil
		}
		// Enter to submit
		if msg.String() == "enter" {
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				m.textarea.Reset()
				return m, func() tea.Msg { return messages.SubmitMsg{Text: text} }
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the command bar.
func (m CommandBar) View() string {
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

// SetWidth returns a new CommandBar with the given width.
func (m CommandBar) SetWidth(width int) CommandBar {
	m.width = width
	m.textarea.SetWidth(width)
	return m
}

// Height returns the height of the command bar.
func (m CommandBar) Height() int {
	return 5 // 3 lines + 2 padding
}

// Focus focuses the command bar.
func (m CommandBar) Focus() tea.Cmd {
	return m.textarea.Focus()
}
