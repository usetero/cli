package progress

import "strings"

import "github.com/usetero/cli/internal/interfaces/tui/theme"

// Model renders a simple text progress bar.
type Model struct {
	theme theme.Theme
	width int
}

// New constructs a progress bar with the given width.
func New(theme theme.Theme, width int) *Model {
	m := &Model{}
	m.theme = theme
	m.SetWidth(width)
	return m
}

// SetWidth updates bar width.
func (m *Model) SetWidth(width int) {
	if width < 10 {
		width = 10
	}
	m.width = width
}

// ViewAs renders progress as percent [0..100].
func (m *Model) ViewAs(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int((percent / 100.0) * float64(m.width))
	if filled > m.width {
		filled = m.width
	}
	if filled < 0 {
		filled = 0
	}
	bar := m.theme.Progress.Fill.Render(strings.Repeat("#", filled)) +
		m.theme.Progress.Empty.Render(strings.Repeat("-", m.width-filled))
	return "[" + bar + "]"
}
