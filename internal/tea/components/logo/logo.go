// Package logo renders the Tero wordmark.
package logo

import (
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/styles"
)

const art = `▀▀▀▀▀ █▀▀▀▀ █▀▀▀▄ ▄▀▀▀▄
  █   █▀▀▀▀ █▀▀▀▄ █   █
  ▀   ▀▀▀▀▀ ▀   ▀  ▀▀▀ `

// Model renders the Tero wordmark.
type Model struct {
	theme   *styles.Theme
	compact bool
}

// New creates a new logo.
func New(theme *styles.Theme) *Model {
	return &Model{theme: theme}
}

// SetCompact enables compact (single-line) mode.
func (m *Model) SetCompact(compact bool) {
	m.compact = compact
}

// View renders the logo.
func (m *Model) View() string {
	if m.compact {
		return m.viewCompact()
	}
	return m.viewFull()
}

// viewFull renders the full multi-line wordmark.
func (m *Model) viewFull() string {
	colors := m.theme.Colors

	var b strings.Builder
	for _, line := range strings.Split(art, "\n") {
		fmt.Fprintln(&b, styles.ApplyForegroundGrad(line, colors.Brand.GradientStart, colors.Brand.GradientEnd))
	}

	return strings.TrimSpace(b.String())
}

// viewCompact renders a single-line version for narrow terminals.
func (m *Model) viewCompact() string {
	colors := m.theme.Colors
	return styles.ApplyForegroundGrad("tero", colors.Brand.GradientStart, colors.Brand.GradientEnd)
}
