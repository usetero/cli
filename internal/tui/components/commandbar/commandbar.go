package commandbar

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/cursor"
)

const (
	// InputHeight is the height of the input area in lines
	InputHeight = 3
)

// SubmitMsg is sent when the user submits input.
type SubmitMsg struct {
	Text string
}

// CommandBar is the input component for chat.
// It renders a multi-line textarea with a prompt.
type CommandBar struct {
	theme  *styles.Theme
	logger log.Logger

	// Input
	textarea textarea.Model

	// Dimensions
	width int
}

// New creates a new command bar
func New(theme *styles.Theme, logger log.Logger) *CommandBar {
	colors := theme.Colors

	ta := textarea.New()
	ta.Placeholder = "Type a message or / for commands..."
	ta.ShowLineNumbers = false
	ta.SetHeight(InputHeight)
	ta.CharLimit = -1
	ta.SetVirtualCursor(false)
	ta.Focus()

	// Style the textarea
	base := lipgloss.NewStyle().Background(colors.Page.Bg)
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:        base,
			Text:        base.Foreground(colors.Page.Text),
			Placeholder: base.Foreground(colors.Page.TextMuted),
			Prompt:      base.Foreground(colors.Accent),
		},
		Blurred: textarea.StyleState{
			Base:        base,
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

	// Set prompt function for multi-line support
	ta.SetPromptFunc(4, func(info textarea.PromptInfo) string {
		if info.LineNumber == 0 {
			return "> "
		}
		return "::: "
	})

	return &CommandBar{
		theme:    theme,
		logger:   logger,
		textarea: ta,
	}
}

// Init initializes the command bar
func (c *CommandBar) Init() tea.Cmd {
	return tea.Batch(
		c.textarea.Focus(),
		textarea.Blink,
	)
}

// Update handles messages
func (c *CommandBar) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Shift+enter or ctrl+j for newline
		if msg.String() == "shift+enter" || msg.String() == "ctrl+j" {
			c.textarea.InsertRune('\n')
			return nil
		}
		// Enter to submit
		if msg.String() == "enter" {
			text := strings.TrimSpace(c.textarea.Value())
			if text != "" {
				c.textarea.Reset()
				return func() tea.Msg { return SubmitMsg{Text: text} }
			}
			return nil
		}
	}

	// Update textarea
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	return cmd
}

// View renders the input area with cursor marker
func (c *CommandBar) View() string {
	if c.width == 0 {
		return ""
	}

	// Get RAW textarea view BEFORE any styling
	textareaView := c.textarea.View()

	// Insert marker in RAW view BEFORE applying padding
	cur := c.textarea.Cursor()
	if cur != nil {
		textareaView = cursor.Insert(textareaView, cur.X, cur.Y)
	}

	// NOW apply padding (lipgloss will preserve the marker position)
	paddedView := lipgloss.NewStyle().
		Padding(1, 0).
		Render(textareaView)

	return paddedView
}

// SetSize sets the width of the command bar
func (c *CommandBar) SetSize(width int) {
	c.width = width
	c.textarea.SetWidth(width)
}

// SetPlaceholder sets the placeholder text
func (c *CommandBar) SetPlaceholder(placeholder string) {
	c.textarea.Placeholder = placeholder
}

// Value returns the current input value
func (c *CommandBar) Value() string {
	return c.textarea.Value()
}

// Reset clears the input
func (c *CommandBar) Reset() {
	c.textarea.Reset()
}

// Focus focuses the input
func (c *CommandBar) Focus() tea.Cmd {
	return c.textarea.Focus()
}

// Blur removes focus from the input
func (c *CommandBar) Blur() {
	c.textarea.Blur()
}

// Height returns the height of the component
func (c *CommandBar) Height() int {
	return InputHeight + 2 // +2 for padding (1 top, 1 bottom)
}
