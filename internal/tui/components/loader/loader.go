package loader

import (
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/styles"
)

// Component is a loading indicator with an animated spinner
type Component struct {
	spinner spinner.Model
	message string
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new loading component
func New(message string) *Component {
	theme := styles.CurrentTheme()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &Component{
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
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Foreground(theme.Primary)

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
