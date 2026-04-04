package progressbar

import (
	"math"

	"charm.land/bubbles/v2/progress"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model renders a themed progress bar.
type Model struct {
	bar   progress.Model
	width int
}

// New constructs a progress bar.
func New(appTheme theme.Theme, width int) *Model {
	bar := progress.New(
		progress.WithColors(appTheme.Palette.Accent, appTheme.Palette.Brand),
		progress.WithFillCharacters('█', '░'),
		progress.WithWidth(max(width, 1)),
	)
	bar.PercentFormat = " %.0f%%"
	bar.PercentageStyle = appTheme.Text.Subtle
	bar.EmptyColor = appTheme.Palette.TextMuted

	return &Model{
		bar:   bar,
		width: max(width, 1),
	}
}

// SetWidth updates the rendered bar width.
func (m *Model) SetWidth(width int) {
	m.width = max(width, 1)
	m.bar.SetWidth(m.width)
}

// ViewAs renders the bar at the provided percentage from 0 to 100.
func (m *Model) ViewAs(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return m.bar.ViewAs(math.Max(0, math.Min(1, percent/100)))
}
