package footer

import (
	"github.com/charmbracelet/bubbles/v2/help"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/styles"
)

// Component is a footer component that displays help text and error state
type Component struct {
	help   help.Model
	keyMap help.KeyMap
	err    error // Current error state (persistent until cleared)
	width  int
	logger log.Logger
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new footer component
func New(logger log.Logger) *Component {
	theme := styles.CurrentTheme()
	helpModel := help.New()
	helpModel.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(theme.TextMuted),
		ShortDesc:      lipgloss.NewStyle().Foreground(theme.TextSubtle),
		ShortSeparator: lipgloss.NewStyle().Foreground(theme.Border),
		Ellipsis:       lipgloss.NewStyle().Foreground(theme.Border),
		FullKey:        lipgloss.NewStyle().Foreground(theme.TextMuted),
		FullDesc:       lipgloss.NewStyle().Foreground(theme.TextSubtle),
		FullSeparator:  lipgloss.NewStyle().Foreground(theme.Border),
	}

	return &Component{
		help:   helpModel,
		logger: logger,
	}
}

// Init initializes the component
func (c *Component) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.logger.Debug("footer received window size", "width", msg.Width, "height", msg.Height)
		c.width = msg.Width
		c.help.Width = msg.Width
	}

	return nil
}

// SetKeyMap sets the key bindings to display in help
func (c *Component) SetKeyMap(keyMap help.KeyMap) {
	c.keyMap = keyMap
}

// SetError sets the error to display in the footer
func (c *Component) SetError(err error) {
	c.err = err
}

// View renders the footer - error above help text
func (c *Component) View() string {
	var help string
	if c.keyMap != nil {
		help = c.help.ShortHelpView(c.keyMap.ShortHelp())
	}

	// If there's an error, show it above the help with spacing
	if c.err != nil {
		c.logger.Debug("rendering error", "width", c.width, "error", c.err.Error())
		if help != "" {
			return lipgloss.JoinVertical(
				lipgloss.Left,
				c.renderError(),
				"", // Empty line for spacing
				help,
			)
		}
		return c.renderError()
	}

	// Default: just show help
	return help
}

// renderError renders an error banner
func (c *Component) renderError() string {
	theme := styles.CurrentTheme()

	labelStyle := lipgloss.NewStyle().
		Background(theme.ErrorBackground).
		Foreground(theme.Text).
		Padding(0, 1).
		Bold(true)

	labelText := labelStyle.Render("ERROR")

	// Calculate width left for message
	widthLeft := c.width - lipgloss.Width(labelText) - 2

	// Truncate message if needed
	message := ansi.Truncate(c.err.Error(), widthLeft, "…")

	messageStyle := lipgloss.NewStyle().
		Background(theme.ErrorBackground).
		Foreground(theme.Text).
		Width(widthLeft+2).
		Padding(0, 1)

	messageText := messageStyle.Render(message)

	return ansi.Truncate(labelText+messageText, c.width, "…")
}

// IsBusy returns false - footer components are never busy
func (c *Component) IsBusy() bool {
	return false
}

// HasError returns false - footer displays errors but doesn't have its own error state
func (c *Component) HasError() bool {
	return false
}

// Error returns nil - footer displays errors but doesn't have its own error state
func (c *Component) Error() error {
	return nil
}
