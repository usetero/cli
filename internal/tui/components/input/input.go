package input

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/styles"
)

// Component wraps textinput.Model with sensible defaults
type Component struct {
	model  textinput.Model
	logger log.Logger
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new input component with themed defaults
func New(logger log.Logger) *Component {
	theme := styles.CurrentTheme()

	ti := textinput.New()
	ti.VirtualCursor = false
	ti.Prompt = "> "
	ti.CharLimit = 256
	ti.Focus() // Focus immediately like Crush does

	ti.Styles = textinput.Styles{
		Focused: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(theme.Text),
			Placeholder: lipgloss.NewStyle().Foreground(theme.TextSubtle),
			Prompt:      lipgloss.NewStyle().Foreground(theme.Primary),
		},
		Blurred: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(theme.TextMuted),
			Placeholder: lipgloss.NewStyle().Foreground(theme.TextSubtle),
			Prompt:      lipgloss.NewStyle().Foreground(theme.TextMuted),
		},
		Cursor: textinput.CursorStyle{
			Color: theme.Primary,
			Shape: tea.CursorBar,
			Blink: true,
		},
	}

	return &Component{model: ti, logger: logger}
}

// Init initializes the component
func (c *Component) Init() tea.Cmd {
	return nil
}

// SetPlaceholder sets the placeholder text
func (c *Component) SetPlaceholder(placeholder string) {
	c.model.Placeholder = placeholder
}

// SetCharLimit sets the character limit
func (c *Component) SetCharLimit(limit int) {
	c.model.CharLimit = limit
}

// SetWidth sets the input width
func (c *Component) SetWidth(width int) {
	c.model.SetWidth(width)
}

// SetEchoMode sets the echo mode (e.g., for password fields)
func (c *Component) SetEchoMode(mode textinput.EchoMode) {
	c.model.EchoMode = mode
}

// SetEchoCharacter sets the character to display when in echo mode
func (c *Component) SetEchoCharacter(char rune) {
	c.model.EchoCharacter = char
}

// Focus focuses the input
func (c *Component) Focus() tea.Cmd {
	c.logger.Debug("input focused")
	return c.model.Focus()
}

// Value returns the current input value
func (c *Component) Value() string {
	return c.model.Value()
}

// Cursor returns the cursor position
func (c *Component) Cursor() *tea.Cursor {
	cursor := c.model.Cursor()
	if cursor != nil {
		c.logger.Debug("input cursor position", "x", cursor.X, "y", cursor.Y)
	} else {
		c.logger.Debug("input cursor is nil")
	}
	return cursor
}

// View renders the input with cursor marker inserted
func (c *Component) View() string {
	view := c.model.View()
	cursor := c.model.Cursor()

	// Insert cursor marker at cursor position
	if cursor != nil && cursor.X >= 0 && cursor.X <= len(view) {
		view = view[:cursor.X] + "\x00CURSOR\x00" + view[cursor.X:]
	}

	return view
}

// Update handles messages
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.model, cmd = c.model.Update(msg)

	// Log cursor blink commands to debug cursor issues
	if cmd != nil {
		c.logger.Debug("input update returned command", "msgType", fmt.Sprintf("%T", msg))
	}

	return cmd
}

// IsBusy returns false - input components are never busy
func (c *Component) IsBusy() bool {
	return false
}

// HasError returns false - input components don't have error states
func (c *Component) HasError() bool {
	return false
}

// Error returns nil - input components don't have errors
func (c *Component) Error() error {
	return nil
}
