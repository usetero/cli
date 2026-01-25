package commandbar

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
)

const (
	// Input area height (lines)
	InputHeight = 3
)

// CommandBar is the input area and footer shown at the bottom of the app.
// It adapts based on page capabilities - showing natural language input
// for chat, or command-only input for other pages.
type CommandBar struct {
	theme  *styles.Theme
	logger log.Logger

	// Input
	textarea textarea.Model

	// Help/footer
	help help.Model

	// Page capabilities
	acceptsNaturalLanguage bool
	commands               []page.Command
	keyBindings            []key.Binding

	// Dimensions
	width int
}

// New creates a new command bar
func New(theme *styles.Theme, logger log.Logger) *CommandBar {
	ta := textarea.New()
	ta.Placeholder = "Type a message or / for commands..."
	ta.ShowLineNumbers = false
	ta.SetHeight(InputHeight)
	ta.Focus()

	h := help.New()

	return &CommandBar{
		theme:    theme,
		logger:   logger,
		textarea: ta,
		help:     h,
	}
}

// Init initializes the command bar
func (c *CommandBar) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages
func (c *CommandBar) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update textarea
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// View renders the command bar (input + footer)
func (c *CommandBar) View() string {
	if c.width == 0 {
		return ""
	}

	colors := c.theme.Colors

	// Input area
	c.textarea.SetWidth(c.width - 4) // Account for prompt and padding
	inputView := c.renderInput()

	// Footer with keybindings
	footerView := c.renderFooter()

	// Compose
	return lipgloss.NewStyle().
		Width(c.width).
		Background(colors.Page.Bg).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			inputView,
			footerView,
		))
}

// renderInput renders the input area with prompt
func (c *CommandBar) renderInput() string {
	colors := c.theme.Colors

	prompt := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Render("> ")

	// Style textarea (focused state)
	s := c.textarea.Styles()
	s.Focused.Base = lipgloss.NewStyle().
		Background(colors.Page.Bg)
	s.Focused.Placeholder = lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted)
	c.textarea.SetStyles(s)

	return lipgloss.NewStyle().
		Padding(0, 1).
		Render(prompt + c.textarea.View())
}

// renderFooter renders the keybindings footer
func (c *CommandBar) renderFooter() string {
	colors := c.theme.Colors

	// Build keybindings display
	var parts []string

	// Always show / for commands
	parts = append(parts, c.renderBinding("/", "commands"))

	// Add page-specific keybindings
	for _, kb := range c.keyBindings {
		if kb.Help().Key != "" {
			parts = append(parts, c.renderBinding(kb.Help().Key, kb.Help().Desc))
		}
	}

	// Always show quit
	parts = append(parts, c.renderBinding("ctrl+c", "quit"))

	footer := strings.Join(parts, "  ")

	return lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Padding(0, 1).
		Render(footer)
}

// renderBinding renders a single keybinding
func (c *CommandBar) renderBinding(key, desc string) string {
	colors := c.theme.Colors

	keyStyle := lipgloss.NewStyle().
		Foreground(colors.Page.Text)
	descStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted)

	return keyStyle.Render(key) + " " + descStyle.Render(desc)
}

// SetSize sets the width of the command bar
func (c *CommandBar) SetSize(width int) {
	c.width = width
	c.help.SetWidth(width - 2)
}

// SetAcceptsNaturalLanguage updates whether free-form input is accepted
func (c *CommandBar) SetAcceptsNaturalLanguage(accepts bool) {
	c.acceptsNaturalLanguage = accepts
	if accepts {
		c.textarea.Placeholder = "Type a message or / for commands..."
	} else {
		c.textarea.Placeholder = "/ for commands"
	}
}

// SetCommands updates the available slash commands
func (c *CommandBar) SetCommands(commands []page.Command) {
	c.commands = commands
}

// SetKeyBindings updates the keybindings shown in footer
func (c *CommandBar) SetKeyBindings(bindings []key.Binding) {
	c.keyBindings = bindings
}

// Value returns the current input value
func (c *CommandBar) Value() string {
	return c.textarea.Value()
}

// Clear clears the input
func (c *CommandBar) Clear() {
	c.textarea.Reset()
}

// Cursor returns the cursor position for the view
func (c *CommandBar) Cursor() *tea.Cursor {
	// TODO: Implement proper cursor positioning
	return nil
}
