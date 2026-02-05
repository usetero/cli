// Package base provides the foundation layout with padding and footer.
package base

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/footer"
)

const (
	horizontalPadding = 1
	verticalPadding   = 1
	footerSpacing     = 1
)

// Model is the foundation layout with padding and footer.
type Model struct {
	theme  *styles.Theme
	footer *footer.Model
	width  int
	height int

	// Cached styles
	contentStyle lipgloss.Style
	outerStyle   lipgloss.Style
}

// New creates a new base layout.
func New(theme *styles.Theme) *Model {
	return &Model{
		theme:  theme,
		footer: footer.New(theme),
	}
}

// SetSize sets the layout dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.footer.SetWidth(width - (horizontalPadding * 2))
	m.updateStyles()
}

func (m *Model) updateStyles() {
	if m.width == 0 || m.height == 0 {
		return
	}
	innerWidth := m.width - (horizontalPadding * 2)
	footerHeight := m.footer.Height()
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing

	m.contentStyle = lipgloss.NewStyle().
		Width(innerWidth).
		Height(contentHeight)
	m.outerStyle = lipgloss.NewStyle().
		Padding(verticalPadding, horizontalPadding)
}

// SetKeyBindings sets the key bindings to display in the footer.
func (m *Model) SetKeyBindings(bindings []key.Binding) {
	m.footer.SetKeyBindings(bindings)
}

// SetError sets an error to display in the footer.
func (m *Model) SetError(err error) {
	m.footer.SetError(err)
}

// ClearError clears any displayed error.
func (m *Model) ClearError() {
	m.footer.ClearError()
}

// ContentSize returns the available space for content.
func (m *Model) ContentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}

	contentWidth := m.width - (horizontalPadding * 2)
	footerHeight := m.footer.Height()
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing

	return contentWidth, contentHeight
}

// Render wraps content with footer and padding.
func (m *Model) Render(content string) string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	innerWidth := m.width - (horizontalPadding * 2)
	footerView := m.footer.View()
	footerHeight := lipgloss.Height(footerView)
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing

	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(contentHeight)
	styledContent := contentStyle.Render(content)

	innerView := lipgloss.JoinVertical(
		lipgloss.Top,
		styledContent,
		"",
		footerView,
	)

	return lipgloss.NewStyle().
		Padding(verticalPadding, horizontalPadding).
		Render(innerView)
}
