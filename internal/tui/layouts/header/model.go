package header

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/header"
	"github.com/usetero/cli/internal/tui/layouts/base"
)

const horizontalPadding = 1

// Model is a layout with header at top, content below, wrapped in base.
type Model struct {
	theme  *styles.Theme
	logger log.Logger
	base   base.Model
	header header.Model
	width  int
	height int
}

// New creates a new header layout.
func New(theme *styles.Theme, logger log.Logger) Model {
	return Model{
		theme:  theme,
		logger: logger,
		base:   base.New(theme, logger),
		header: header.New(theme, logger),
	}
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.base = m.base.SetSize(width, height)
	baseWidth := width - (horizontalPadding * 2)
	m.header = m.header.SetWidth(baseWidth)
	return m
}

// SetKeyBindings returns a new Model with the given key bindings.
func (m Model) SetKeyBindings(bindings []key.Binding) Model {
	m.base = m.base.SetKeyBindings(bindings)
	return m
}

// SetError returns a new Model with the given error.
func (m Model) SetError(err error) Model {
	m.base = m.base.SetError(err)
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.base, cmd = m.base.Update(msg)
	return m, cmd
}

// SetTitle returns a new Model with the given title.
func (m Model) SetTitle(title string) Model {
	m.header = m.header.SetTitle(title)
	return m
}

// SetOrgName returns a new Model with the given org name.
func (m Model) SetOrgName(name string) Model {
	m.header = m.header.SetOrgName(name)
	return m
}

// ContentSize returns the available space for content.
func (m Model) ContentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}

	baseWidth, baseHeight := m.base.ContentSize()
	headerHeight := m.header.Height()
	contentHeight := baseHeight - headerHeight

	return baseWidth, contentHeight
}

// Render composes header + content, then wraps in base layout.
func (m Model) Render(content string) string {
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
