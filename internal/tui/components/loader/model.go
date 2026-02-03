package loader

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

// Model is a loading indicator with an animated spinner.
type Model struct {
	theme   *styles.Theme
	spinner spinner.Model
	message string
}

// New creates a new loader model.
func New(theme *styles.Theme, message string) Model {
	colors := theme.Colors

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colors.Accent)

	return Model{
		theme:   theme,
		spinner: s,
		message: message,
	}
}

// Init starts the spinner animation.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles spinner tick messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the loading indicator.
func (m Model) View() string {
	colors := m.theme.Colors

	style := lipgloss.NewStyle().
		Foreground(colors.Accent)

	return m.spinner.View() + " " + style.Render(m.message+"...")
}

// SetMessage returns a new Model with the given message.
func (m Model) SetMessage(message string) Model {
	m.message = message
	return m
}
