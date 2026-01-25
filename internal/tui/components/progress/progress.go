package progress

import (
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/styles"
)

// Progress wraps the Bubbles progress component with Tero styling.
type Progress struct {
	theme *styles.Theme
	model *progress.Model
}

// New creates a new progress bar with Tero theming.
// Uses a gradient from Brand.GradientStart to Brand.GradientEnd.
func New(theme *styles.Theme, width int) *Progress {
	colors := theme.Colors

	p := progress.New(
		progress.WithColors(colors.Brand.GradientStart, colors.Brand.GradientEnd),
		progress.WithWidth(width),
		progress.WithFillCharacters('█', '░'),
	)

	// Style the percentage text and empty sections
	p.PercentFormat = " %.1f%%"
	p.PercentageStyle = p.PercentageStyle.Foreground(colors.Page.Text)
	p.EmptyColor = colors.Page.TextMuted

	return &Progress{
		theme: theme,
		model: &p,
	}
}

// Update forwards messages to the underlying progress model.
func (p *Progress) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	model, cmd := p.model.Update(msg)
	p.model = &model
	return cmd
}

// SetPercent sets the progress percentage (0-1).
func (p *Progress) SetPercent(percent float64) tea.Cmd {
	return p.model.SetPercent(percent)
}

// ViewAs renders the progress bar at a specific percentage without animation.
// Percent should be 0-100.
func (p *Progress) ViewAs(percent float64) string {
	// Bubbles expects 0-1, we work with 0-100
	return p.model.ViewAs(percent / 100)
}

// View renders the progress bar.
func (p *Progress) View() string {
	return p.model.View()
}

// SetWidth updates the width of the progress bar.
func (p *Progress) SetWidth(width int) {
	p.model.SetWidth(width)
}
