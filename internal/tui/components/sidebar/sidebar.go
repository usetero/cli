package sidebar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/logo"
)

const diag = `╱`

// Sidebar renders page metadata in a sidebar format.
// Used when the window is wide enough to show sidebar + content.
type Sidebar struct {
	theme  *styles.Theme
	width  int
	height int
	logger log.Logger

	// Content
	title    string
	orgName  string
	metadata []page.Metadata
}

// New creates a new sidebar
func New(theme *styles.Theme, logger log.Logger) *Sidebar {
	return &Sidebar{
		theme:  theme,
		logger: logger,
	}
}

// SetSize sets the dimensions for the sidebar
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetTitle sets the page title
func (s *Sidebar) SetTitle(title string) {
	s.title = title
}

// SetOrgName sets the organization name
func (s *Sidebar) SetOrgName(name string) {
	s.orgName = name
}

// SetMetadata sets the metadata items to display
func (s *Sidebar) SetMetadata(meta []page.Metadata) {
	s.metadata = meta
}

// renderSection creates a section header with a text label followed by a line
func (s *Sidebar) renderSection(text string) string {
	char := "─"
	length := lipgloss.Width(text) + 1
	remainingWidth := s.width - length
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	lineStyle := lipgloss.NewStyle().Foreground(s.theme.Colors.BorderDefault)
	return text + " " + lineStyle.Render(strings.Repeat(char, remainingWidth))
}

// View renders the sidebar
func (s *Sidebar) View() string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	colors := s.theme.Colors

	// Container style
	style := lipgloss.NewStyle().
		Width(s.width).
		Height(s.height)

	// Diagonal lines (brand element)
	fieldStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)
	divider := fieldStyle.Render(strings.Repeat(diag, s.width))

	// Logo with version
	logoView := logo.Render(logo.Opts{
		TitleColorA: colors.Brand.GradientStart,
		TitleColorB: colors.Brand.GradientEnd,
	})
	logoLines := strings.Split(logoView, "\n")

	versionText := lipgloss.NewStyle().
		Foreground(colors.Panel.TextMuted).
		Render("v0.0.1")

	if len(logoLines) > 0 {
		lastLine := logoLines[len(logoLines)-1]
		lastLineWidth := lipgloss.Width(lastLine)
		spacingWidth := s.width - lastLineWidth - lipgloss.Width(versionText)
		if spacingWidth < 0 {
			spacingWidth = 0
		}
		logoLines[len(logoLines)-1] = lastLine + strings.Repeat(" ", spacingWidth) + versionText
	}
	logoWithVersion := strings.Join(logoLines, "\n")

	// Org name
	orgStyle := lipgloss.NewStyle().Foreground(colors.Panel.Text)
	orgView := orgStyle.Render(s.orgName)

	// Build content
	var parts []string
	parts = append(parts, divider, divider, "", logoWithVersion, "", orgView, "")

	// Context section with metadata
	if len(s.metadata) > 0 {
		parts = append(parts, s.renderSection("Context"), "")

		for _, m := range s.metadata {
			parts = append(parts, s.renderMetadataItem(m))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return style.Render(content)
}

// renderMetadataItem renders a single metadata item
func (s *Sidebar) renderMetadataItem(m page.Metadata) string {
	colors := s.theme.Colors
	labelStyle := lipgloss.NewStyle().Foreground(colors.Panel.TextMuted)
	valueStyle := lipgloss.NewStyle().Foreground(colors.Panel.Text)

	// Apply custom color if specified
	if m.Color != nil {
		valueStyle = valueStyle.Foreground(m.Color)
	}

	// Icon if present
	prefix := ""
	if m.Icon != "" {
		prefix = m.Icon + " "
	}

	label := labelStyle.Render(prefix + m.Label)
	value := valueStyle.Render(m.Value)

	// Right-align value
	labelWidth := lipgloss.Width(label)
	valueWidth := lipgloss.Width(value)
	spacing := s.width - labelWidth - valueWidth
	if spacing < 1 {
		spacing = 1
	}

	return label + strings.Repeat(" ", spacing) + value
}
