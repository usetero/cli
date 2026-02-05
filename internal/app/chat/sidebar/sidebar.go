// Package sidebar renders session info for the chat view.
package sidebar

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/logo"
)

const (
	diag                = "╱"
	logoHeightThreshold = 30
)

// Model renders session info for the chat view.
type Model struct {
	theme     *styles.Theme
	logo      *logo.Model
	width     int
	height    int
	title     string
	version   string
	workspace string
}

// New creates a new sidebar.
func New(theme *styles.Theme) *Model {
	return &Model{
		theme: theme,
		logo:  logo.New(theme),
	}
}

// SetSize sets the dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetTitle sets the conversation title.
func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetVersion sets the version string.
func (m *Model) SetVersion(version string) {
	m.version = version
}

// SetWorkspace sets the workspace name.
func (m *Model) SetWorkspace(workspace string) {
	m.workspace = workspace
}

// Width returns the sidebar width.
func (m *Model) Width() int {
	return m.width
}

// View renders the sidebar.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	colors := m.theme.Colors
	diagStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)

	var parts []string

	if m.height > logoHeightThreshold {
		// Full logo with diagonal lines and version
		diagLine := diagStyle.Render(strings.Repeat(diag, m.width))

		// Version row (right-aligned)
		versionStyle := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientStart).
			Width(m.width).
			Align(lipgloss.Right)
		versionRow := versionStyle.Render(m.version)

		logoView := m.logo.View()
		parts = append(parts, diagLine, diagLine, versionRow, logoView, diagLine, "")
	} else {
		// Compact logo
		m.logo.SetCompact(true)
		parts = append(parts, m.logo.View(), "")
	}

	// Title
	title := m.title
	if title == "" {
		title = "New conversation"
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted).
		Width(m.width)
	parts = append(parts, titleStyle.Render(title), "")

	// Workspace
	if m.workspace != "" {
		workspaceStyle := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Width(m.width)
		parts = append(parts, workspaceStyle.Render(m.workspace), "")
	}

	// Section helper
	renderSection := func(name string) string {
		nameWidth := lipgloss.Width(name)
		lineWidth := m.width - nameWidth - 1
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
