package base

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/footer"
)

const (
	horizontalPadding = 1
	verticalPadding   = 1
	footerSpacing     = 1
)

// Model is the foundation layout with padding and footer.
type Model struct {
	theme  *styles.Theme
	logger log.Logger
	footer footer.Model
	width  int
	height int
}

// New creates a new base layout.
func New(theme *styles.Theme, logger log.Logger) Model {
	return Model{
		theme:  theme,
		logger: logger,
		footer: footer.New(theme, logger),
	}
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	innerWidth := width - (horizontalPadding * 2)
	m.footer = m.footer.SetWidth(innerWidth)
	return m
}

// SetKeyBindings returns a new Model with the given key bindings.
func (m Model) SetKeyBindings(bindings []key.Binding) Model {
	m.footer = m.footer.SetKeyBindings(bindings)
	return m
}

// SetError returns a new Model with the given error.
func (m Model) SetError(err error) Model {
	m.footer = m.footer.SetError(err)
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.footer, cmd = m.footer.Update(msg)
	return m, cmd
}

// ContentSize returns the available space for content.
func (m Model) ContentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}

	contentWidth := m.width - (horizontalPadding * 2)
	footerHeight := m.footer.Height()
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing

	return contentWidth, contentHeight
}

// Render wraps content with footer and padding.
func (m Model) Render(content string) string {
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
