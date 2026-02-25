package inputbar

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/app/palette"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/cursor"
	"github.com/usetero/cli/internal/tea/keymap"
)

const (
	textareaHeight = 3 // visible input lines

	// Layout: border(1) + innerPadL(2) + text + innerPadR(2)
	borderWidth   = 1
	innerPadX     = 2
	innerPadY     = 1
	chrome        = borderWidth + innerPadX*2

	inputBarHeight = textareaHeight + innerPadY*2 // textarea + top/bottom inner padding
)

// Model handles user input via a textarea.
type Model struct {
	theme       styles.Theme
	textarea    textarea.Model
	width       int
	scope       log.Scope
	pendingText string // saved input text, restored on stream failure
	placeholder string // rendered outside textarea to avoid bg issues
}

// placeholder returns a random placeholder for the session.
func placeholder(user *auth.User) string {
	name := ""
	if user != nil && user.FirstName != "" {
		name = user.FirstName
	}

	pool := []string{
		"What should we get into?",
		"What's on your mind?",
	}
	if name != "" {
		pool = append(pool,
			"Ready when you are, "+name,
			"Let's get to work, "+name,
		)
	}
	return pool[rand.IntN(len(pool))]
}

// New creates a new input bar.
func New(user *auth.User, theme styles.Theme, scope log.Scope) *Model {
	scope = scope.Child("inputbar")

	// Input bar uses elevated background to match user message blocks.
	elevated := theme.WithBg(theme.BgElevated)

	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.SetHeight(textareaHeight)
	ta.CharLimit = -1
	ta.SetVirtualCursor(false)
	ta.Focus()

	base := lipgloss.NewStyle().Foreground(elevated.Text).Background(elevated.Bg)
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:   base,
			Text:   base,
			Prompt: base,
		},
		Blurred: textarea.StyleState{
			Base:   base.Foreground(elevated.TextMuted),
			Text:   base.Foreground(elevated.TextMuted),
			Prompt: base.Foreground(elevated.TextMuted),
		},
		Cursor: textarea.CursorStyle{
			Color: elevated.Accent,
			Shape: tea.CursorBar,
			Blink: true,
		},
	})

	ta.SetPromptFunc(0, func(_ textarea.PromptInfo) string {
		return ""
	})

	return &Model{
		theme:       elevated,
		textarea:    ta,
		scope:       scope,
		placeholder: placeholder(user),
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

		// "/" on empty input opens the command palette
		if key.Matches(msg, keymap.Palette) && m.textarea.Value() == "" {
			return func() tea.Msg { return palette.OpenMsg{} }
		}

		// Newline - consume, don't forward to textarea
		if key.Matches(msg, keymap.Newline) {
			m.textarea.InsertRune('\n')
			return nil
		}
		// Enter to submit - consume, don't forward to textarea
		if key.Matches(msg, keymap.Send) {
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				m.pendingText = text
				m.textarea.Reset()
				return func() tea.Msg { return msgs.UserSubmittedInput{Text: text} }
			}
			return nil
		}

	case msgs.StreamFailed:
		if m.pendingText != "" {
			m.textarea.SetValue(m.pendingText)
			m.pendingText = ""
		}
		return nil

	case msgs.StreamCompleted:
		m.pendingText = ""
		return nil
	}

	// Forward to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return cmd
}

// View renders the input bar styled like a user message block:
// grey background, green left border, matching padding.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	var view string
	if m.textarea.Value() == "" {
		// Render placeholder ourselves so background is correct.
		view = lipgloss.NewStyle().
			Foreground(m.theme.TextMuted).
			Background(m.theme.Bg).
			Render(m.placeholder)
	} else {
		view = m.textarea.View()

		// The textarea emits SGR resets (\033[0m) that kill our background.
		// Re-establish the theme background after every reset.
		r, g, b, _ := m.theme.Bg.RGBA()
		bgSeq := fmt.Sprintf("\033[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
		view = strings.ReplaceAll(view, "\033[0m", "\033[0m"+bgSeq)
	}

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

	// Inner box: elevated bg, padding, matches user message block styling
	contentWidth := m.width - borderWidth
	inner := lipgloss.NewStyle().
		Background(m.theme.Bg).
		Foreground(m.theme.Text).
		Padding(innerPadY, innerPadX).
		Width(contentWidth).
		Render(view)

	// Border: accent left border, matches renderBlock for user messages
	bordered := lipgloss.NewStyle().
		Width(contentWidth + borderWidth).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Accent).
		Render(inner)

	return bordered
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	// Textarea gets the width inside all chrome layers
	m.textarea.SetWidth(width - chrome)
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
	return []key.Binding{keymap.Send, keymap.Newline, keymap.Palette}
}
