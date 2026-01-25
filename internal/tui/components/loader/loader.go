package loader

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components"
)

// Component is a loading indicator with an animated spinner
type Component struct {
	theme   *styles.Theme
	spinner spinner.Model
	message string
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new loading component
func New(theme *styles.Theme, message string) *Component {
	colors := theme.Colors

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return &Component{
		theme:   theme,
		spinner: s,
		message: message,
	}
}

// Init starts the spinner animation
func (c *Component) Init() tea.Cmd {
	return c.spinner.Tick
}

// Update handles spinner tick messages
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.spinner, cmd = c.spinner.Update(msg)
	return cmd
}

// View renders the loading indicator
func (c *Component) View() string {
	colors := c.theme.Colors

	style := lipgloss.NewStyle().
		Foreground(colors.Accent)

	return c.spinner.View() + " " + style.Render(c.message+"...")
}

// IsBusy returns true - loader is always busy when visible
func (c *Component) IsBusy() bool {
	return true
}

// HasError returns false - loader components don't have error states
func (c *Component) HasError() bool {
	return false
}

// Error returns nil - loader components don't have errors
func (c *Component) Error() error {
	return nil
}
