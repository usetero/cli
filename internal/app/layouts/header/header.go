// Package header provides a layout with header at top, content below, wrapped in base.
package header

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/layouts/base"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/header"
)

const horizontalPadding = 1

// Model is a layout with header at top, content below, wrapped in base.
type Model struct {
	theme  *styles.Theme
	base   *base.Model
	header *header.Model
	width  int
	height int
}

// New creates a new header layout.
func New(theme *styles.Theme) *Model {
	return &Model{
		theme:  theme,
		base:   base.New(theme),
		header: header.New(theme),
	}
}

// SetSize sets the layout dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.base.SetSize(width, height)
	m.header.SetWidth(width - (horizontalPadding * 2))
}

// SetKeyBindings sets the key bindings to display in the footer.
func (m *Model) SetKeyBindings(bindings []key.Binding) {
	m.base.SetKeyBindings(bindings)
}

// SetError sets an error to display in the footer.
func (m *Model) SetError(err error) {
	m.base.SetError(err)
}

// ClearError clears any displayed error.
func (m *Model) ClearError() {
	m.base.ClearError()
}

// SetTitle sets the header title.
func (m *Model) SetTitle(title string) {
	m.header.SetTitle(title)
}

// SetOrgName sets the organization name in the header.
func (m *Model) SetOrgName(name string) {
	m.header.SetOrgName(name)
}

// ContentSize returns the available space for content.
func (m *Model) ContentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}

	baseWidth, baseHeight := m.base.ContentSize()
	headerHeight := m.header.Height()
	contentHeight := baseHeight - headerHeight

	return baseWidth, contentHeight
}

// Render composes header + content, then wraps in base layout.
func (m *Model) Render(content string) string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	headerView := m.header.View()

	composedView := lipgloss.JoinVertical(
		lipgloss.Left,
		headerView,
		content,
	)

	return m.base.Render(composedView)
}
