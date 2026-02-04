package commandbar

import (
	"strings"

	"github.com/usetero/cli/internal/log"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/cursor"
)

// SubmitMsg is sent when the user submits text input.
type SubmitMsg struct {
	Text string
}

// Model is the input component for chat.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	textarea textarea.Model
	width    int
}

// New creates a new command bar model.
func New(theme *styles.Theme, logger log.Logger) Model {
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

	return Model{
		theme:    theme,
		logger:   logger,
		textarea: ta,
	}
}

// Init initializes the command bar.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.textarea.Focus(),
		textarea.Blink,
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
				return m, func() tea.Msg { return SubmitMsg{Text: text} }
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the command bar.
func (m Model) View() string {
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

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	m.textarea.SetWidth(width)
	return m
}

// Height returns the height of the command bar.
func (m Model) Height() int {
	return 5 // 3 lines + 2 padding
}

// Focus focuses the command bar.
func (m Model) Focus() tea.Cmd {
	return m.textarea.Focus()
}
