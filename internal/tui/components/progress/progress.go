package progress

import (
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/styles"
)

// Model wraps the Bubbles progress component with Tero styling.
type Model struct {
	theme    *styles.Theme
	progress progress.Model
	width    int
}

// New creates a new progress bar with Tero theming.
func New(theme *styles.Theme, width int) Model {
	colors := theme.Colors

	p := progress.New(
		progress.WithColors(colors.Brand.GradientStart, colors.Brand.GradientEnd),
		progress.WithWidth(width),
		progress.WithFillCharacters('█', '░'),
	)

	p.PercentFormat = " %.1f%%"
	p.PercentageStyle = p.PercentageStyle.Foreground(colors.Page.Text)
	p.EmptyColor = colors.Page.TextMuted

	return Model{
		theme:    theme,
		progress: p,
		width:    width,
	}
}

// Init returns nil (no initialization needed).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

// View renders the progress bar.
func (m Model) View() string {
	return m.progress.View()
}

// ViewAs renders the progress bar at a specific percentage without animation.
// Percent should be 0-100.
func (m Model) ViewAs(percent float64) string {
	return m.progress.ViewAs(percent / 100)
}

// SetPercent returns a command to animate to the given percentage (0-1).
func (m Model) SetPercent(percent float64) tea.Cmd {
	return m.progress.SetPercent(percent)
}

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	m.progress.SetWidth(width)
	return m
}
