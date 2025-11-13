package progress

import (
	"github.com/charmbracelet/bubbles/v2/progress"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/tui/styles"
)

// Progress wraps the Bubbles progress component with Tero styling.
type Progress struct {
	model *progress.Model
}

// New creates a new progress bar with Tero theming.
// Uses a gradient from Primary to Secondary colors.
func New(width int) *Progress {
	theme := styles.CurrentTheme()

	p := progress.New(
		progress.WithGradient(styles.ColorToHex(theme.Primary), styles.ColorToHex(theme.Secondary)),
		progress.WithWidth(width),
		progress.WithFillCharacters('█', '░'),
	)

	// Style the percentage text and empty sections
	p.PercentageStyle = p.PercentageStyle.Foreground(theme.Text)
	p.EmptyColor = theme.TextMuted

	return &Progress{
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
