// Package header provides a branding header component.
package header

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/logo"
)

const (
	diag                  = `╱`
	compactWidthThreshold = 80
)

// Model renders branding and metadata in the header area.
type Model struct {
	theme   *styles.Theme
	logo    *logo.Model
	width   int
	title   string
	orgName string
}

// New creates a new header.
func New(theme *styles.Theme) *Model {
	return &Model{
		theme: theme,
		logo:  logo.New(theme),
	}
}

// SetWidth sets the header width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetTitle sets the header title.
func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetOrgName sets the organization name displayed in the header.
func (m *Model) SetOrgName(name string) {
	m.orgName = name
}

// View renders the header.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	if m.width < compactWidthThreshold {
		return m.viewCompact()
	}
	return m.viewFull()
}

// Height returns the rendered height of the header.
func (m *Model) Height() int {
	return lipgloss.Height(m.View())
}

// viewFull renders the full header with logo and diagonal lines.
func (m *Model) viewFull() string {
	colors := m.theme.Colors

	logoView := m.logo.View()

	fieldHeight := lipgloss.Height(logoView)
	logoWidth := lipgloss.Width(strings.Split(logoView, "\n")[0])

	const leftWidth = 6
	fieldStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)
	leftFieldRow := fieldStyle.Render(strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		leftField.WriteString(leftFieldRow + "\n")
	}

	rightWidth := max(15, m.width-logoWidth-leftWidth-2)
	rightFieldRow := fieldStyle.Render(strings.Repeat(diag, rightWidth))
	rightField := new(strings.Builder)
	for range fieldHeight {
		rightField.WriteString(rightFieldRow + "\n")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.TrimSpace(leftField.String()),
		" ",
		logoView,
		" ",
		strings.TrimSpace(rightField.String()),
	)
}

// viewCompact renders a compact single-line header for narrow terminals.
func (m *Model) viewCompact() string {
	colors := m.theme.Colors

	brandStyle := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientStart).
		Bold(true)
	brand := brandStyle.Render("tero")

	var titlePart string
	if m.title != "" && m.title != "Chat" {
		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientStart).
			Bold(true)
		titlePart = " " + titleStyle.Render(m.title)
	}

	diagStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)

	var rightParts []string
	if m.orgName != "" {
		orgStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
		rightParts = append(rightParts, orgStyle.Render(m.orgName))
	}

	separator := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render(" · ")
	rightContent := strings.Join(rightParts, separator)

	leftWidth := lipgloss.Width(brand + titlePart)
	rightWidth := lipgloss.Width(rightContent)
	diagWidth := m.width - leftWidth - rightWidth - 4
	if diagWidth < 10 {
		diagWidth = 10
	}

	diagPart := diagStyle.Render(strings.Repeat(diag, diagWidth))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		brand+titlePart,
		" ",
		diagPart,
		" ",
		rightContent,
	)
}
