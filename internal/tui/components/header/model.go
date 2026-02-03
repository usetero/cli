package header

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/logo"
)

const (
	diag                  = `╱`
	CompactWidthThreshold = 80
)

// Model renders branding and metadata.
type Model struct {
	theme   *styles.Theme
	logger  log.Logger
	width   int
	title   string
	orgName string
}

// New creates a new header model.
func New(theme *styles.Theme, logger log.Logger) Model {
	return Model{
		theme:  theme,
		logger: logger,
	}
}

// SetWidth returns a new Model with the given width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	return m
}

// SetTitle returns a new Model with the given title.
func (m Model) SetTitle(title string) Model {
	m.title = title
	return m
}

// SetOrgName returns a new Model with the given org name.
func (m Model) SetOrgName(name string) Model {
	m.orgName = name
	return m
}

// View renders the header.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	if m.width < CompactWidthThreshold {
		return m.viewThin()
	}
	return m.viewThick()
}

// Height returns the rendered height of the header.
func (m Model) Height() int {
	return lipgloss.Height(m.View())
}

// viewThick renders the full header with logo and diagonal lines.
func (m Model) viewThick() string {
	colors := m.theme.Colors

	logoView := logo.Render(logo.Opts{
		TitleColorA: colors.Brand.GradientStart,
		TitleColorB: colors.Brand.GradientEnd,
	})

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

// viewThin renders a compact single-line header.
func (m Model) viewThin() string {
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
