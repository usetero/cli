package progress

import (
	bubbleprogress "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Model renders a themed progress bar.
type Model struct {
	theme    theme.Theme
	progress bubbleprogress.Model
	width    int
}

// New constructs a progress bar with the given width.
func New(appTheme theme.Theme, width int) *Model {
	p := bubbleprogress.New(
		bubbleprogress.WithWidth(clampWidth(width)),
		bubbleprogress.WithColors(appTheme.GradientStart, appTheme.GradientEnd),
		bubbleprogress.WithFillCharacters('█', '░'),
	)
	p.PercentFormat = " %.0f%%"
	p.PercentageStyle = appTheme.Text.Body
	p.EmptyColor = appTheme.TextMuted

	return &Model{
		theme:    appTheme,
		progress: p,
		width:    clampWidth(width),
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update advances progress animations.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

// View renders the progress bar at its current percentage.
func (m *Model) View() tea.View { return tea.NewView(m.progress.View()) }

// ViewAs renders the progress bar at a specific percentage without animation.
// Percent should be in the range 0..100.
func (m *Model) ViewAs(percent float64) string {
	return m.progress.ViewAs(clampPercent(percent) / 100.0)
}

// SetPercent animates to the given percentage in the range 0..1.
func (m *Model) SetPercent(percent float64) tea.Cmd {
	return m.progress.SetPercent(percent)
}

// SetWidth updates the progress width.
func (m *Model) SetWidth(width int) {
	m.width = clampWidth(width)
	m.progress.SetWidth(m.width)
}

func clampWidth(width int) int {
	if width < 14 {
		return 14
	}
	return width
}

func clampPercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}
