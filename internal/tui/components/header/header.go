package header

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/logo"
)

const (
	diag = `╱`

	// Width threshold for switching between thick and thin
	CompactWidthThreshold = 80
)

// Header renders branding and metadata.
// Automatically switches between thick (with logo) and thin (single line)
// based on available width.
type Header struct {
	theme  *styles.Theme
	width  int
	logger log.Logger

	// Content
	title    string
	orgName  string
	metadata []page.Metadata
}

// New creates a new header
func New(theme *styles.Theme, logger log.Logger) *Header {
	return &Header{
		theme:  theme,
		logger: logger,
	}
}

// SetSize sets the width of the header
func (h *Header) SetSize(width int) {
	h.width = width
}

// SetTitle sets the page title
func (h *Header) SetTitle(title string) {
	h.title = title
}

// SetOrgName sets the organization name
func (h *Header) SetOrgName(name string) {
	h.orgName = name
}

// SetMetadata sets the metadata items to display
func (h *Header) SetMetadata(meta []page.Metadata) {
	h.metadata = meta
}

// View renders the header (thick or thin based on width)
func (h *Header) View() string {
	if h.width == 0 {
		return ""
	}

	if h.width < CompactWidthThreshold {
		return h.viewThin()
	}
	return h.viewThick()
}

// viewThick renders the full header with logo and diagonal lines
func (h *Header) viewThick() string {
	colors := h.theme.Colors

	// Render the TERO wordmark
	logoView := logo.Render(logo.Opts{
		TitleColorA: colors.Brand.GradientStart,
		TitleColorB: colors.Brand.GradientEnd,
	})

	// Calculate dimensions
	fieldHeight := lipgloss.Height(logoView)
	logoWidth := lipgloss.Width(strings.Split(logoView, "\n")[0])

	// Left diagonal field
	const leftWidth = 6
	fieldStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientStart)
	leftFieldRow := fieldStyle.Render(strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		leftField.WriteString(leftFieldRow + "\n")
	}

	// Right diagonal field fills remaining space
	rightWidth := max(15, h.width-logoWidth-leftWidth-2)
	rightFieldRow := fieldStyle.Render(strings.Repeat(diag, rightWidth))
	rightField := new(strings.Builder)
	for range fieldHeight {
		rightField.WriteString(rightFieldRow + "\n")
	}

	// Join: left diagonals + logo + right diagonals
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.TrimSpace(leftField.String()),
		" ",
		logoView,
		" ",
		strings.TrimSpace(rightField.String()),
	)
}

// viewThin renders a compact single-line header
// Format: tero [Title] //////// OrgName · metadata1 · metadata2
func (h *Header) viewThin() string {
	colors := h.theme.Colors

	// Brand name
	brandStyle := lipgloss.NewStyle().
		Foreground(colors.Brand.GradientStart).
		Bold(true)
	brand := brandStyle.Render("tero")

	// Title (if not chat, show it)
	var titlePart string
	if h.title != "" && h.title != "Chat" {
		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientEnd).
			Bold(true)
		titlePart = " " + titleStyle.Render(h.title)
	}

	// Diagonal separator
	diagStyle := lipgloss.NewStyle().Foreground(colors.Brand.GradientEnd)

	// Right side: org + metadata
	var rightParts []string

	if h.orgName != "" {
		orgStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
		rightParts = append(rightParts, orgStyle.Render(h.orgName))
	}

	// Add high-priority metadata (priority 1-2)
	for _, m := range h.metadata {
		if m.Priority <= 2 {
			valueStyle := lipgloss.NewStyle().Foreground(colors.Page.Text)
			if m.Color != nil {
				valueStyle = valueStyle.Foreground(m.Color)
			}
			rightParts = append(rightParts, valueStyle.Render(m.Value))
		}
	}

	// Join right parts with separator
	separator := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Render(" · ")
	rightContent := strings.Join(rightParts, separator)

	// Calculate diagonal width
	leftWidth := lipgloss.Width(brand + titlePart)
	rightWidth := lipgloss.Width(rightContent)
	diagWidth := h.width - leftWidth - rightWidth - 4
	if diagWidth < 10 {
		diagWidth = 10
	}

	diagPart := diagStyle.Render(strings.Repeat(diag, diagWidth))

	// Compose
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		brand+titlePart,
		" ",
		diagPart,
		" ",
		rightContent,
	)
}
