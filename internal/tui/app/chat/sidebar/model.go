package sidebar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/logo"
)

const diag = "╱"

// Model renders session info for the chat view.
type Model struct {
	theme   *styles.Theme
	logger  log.Logger
	width   int
	height  int
	title   string
	version string
}

// New creates a new sidebar model.
func New(theme *styles.Theme, logger log.Logger) Model {
	return Model{
		theme:  theme,
		logger: logger,
	}
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetTitle returns a new Model with the given title.
func (m Model) SetTitle(title string) Model {
	m.title = title
	return m
}

// SetVersion returns a new Model with the given version.
func (m Model) SetVersion(version string) Model {
	m.version = version
	return m
}

const logoHeightBreakpoint = 30

// View renders the sidebar.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors
	diagStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)
	innerWidth := m.width

	var parts []string

	// Logo with diagonal lines above and below
	logoOpts := logo.Opts{
		TitleColorA: colors.Brand.GradientStart,
		TitleColorB: colors.Brand.GradientEnd,
	}

	if m.height > logoHeightBreakpoint {
		// Full logo with diagonal lines and version
		diagLine := diagStyle.Render(strings.Repeat(diag, innerWidth))

		// Version row (right-aligned, between top diagonals and wordmark)
		versionStyle := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientStart).
			Width(innerWidth).
			Align(lipgloss.Right)
		versionRow := versionStyle.Render(m.version)

		logoView := logo.Render(logoOpts)
		// Diagonals at top, version, wordmark, diagonal at bottom
		parts = append(parts, diagLine, diagLine, versionRow, logoView, diagLine, "")
	} else {
		// Small logo, no diagonal lines
		parts = append(parts, logo.RenderSmall(logoOpts), "")
	}

	// Title
	title := m.title
	if title == "" {
		title = "New conversation"
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Width(innerWidth)
	parts = append(parts, titleStyle.Render(title), "")

	// Section helper
	renderSection := func(name string) string {
		nameWidth := lipgloss.Width(name)
		lineWidth := innerWidth - nameWidth - 1 // -1 for space
		if lineWidth < 3 {
			lineWidth = 3
		}
		line := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render(strings.Repeat("─", lineWidth))
		return lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render(name) + " " + line
	}

	// Context section (placeholder for future)
	parts = append(parts, renderSection("Context"), "")
	noneStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	parts = append(parts, noneStyle.Render("None"), "")

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content)
}

// Width returns the sidebar width.
func (m Model) Width() int {
	return m.width
}
