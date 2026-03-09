package chrome

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

type WidthMode uint8

const (
	WidthFill WidthMode = iota
	WidthIntrinsic
)

type HeightMode uint8

const (
	HeightFill HeightMode = iota
	HeightIntrinsic
)

type VerticalAlign uint8

const (
	AlignTop VerticalAlign = iota
	AlignCenter
	AlignBottom
)

// BodyLayout hints tell chrome how to place body content inside the shell body region.
type BodyLayout struct {
	WidthMode     WidthMode
	HeightMode    HeightMode
	VerticalAlign VerticalAlign
	MaxWidth      int
}

// BodySlot is the shell body content plus placement hints.
type BodySlot struct {
	Content string
	Layout  BodyLayout
}

// Viewport is the current terminal viewport size.
type Viewport struct {
	Width  int
	Height int
}

// Metrics are the measured shell dimensions for the current viewport and slots.
type Metrics struct {
	InnerWidth        int
	InnerHeight       int
	HeaderHeight      int
	FooterHeight      int
	BodyRegionHeight  int
	BodyContentHeight int
	BodyContentWidth  int
}

// Slots defines the shell content regions.
type Slots struct {
	Header string
	Body   BodySlot
	Footer string
}

// DefaultBodyLayout returns the default shell body placement policy.
func DefaultBodyLayout() BodyLayout {
	return BodyLayout{
		WidthMode:     WidthFill,
		HeightMode:    HeightFill,
		VerticalAlign: AlignTop,
	}
}

// Measure returns the shell layout metrics for the provided viewport and slots.
func Measure(t theme.Theme, slots Slots, viewport Viewport) Metrics {
	if viewport.Width <= 0 || viewport.Height <= 0 {
		return Metrics{}
	}

	innerWidth := viewport.Width - t.Shell.Outer.GetHorizontalFrameSize()
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := viewport.Height - t.Shell.Outer.GetVerticalFrameSize()
	if innerHeight < 0 {
		innerHeight = 0
	}

	header := normalizeHeader(t, slots.Header)
	footer := normalizeFooter(slots.Footer)

	headerLine := t.Shell.HeaderBar.Width(innerWidth).Render(header)
	footerHeight := 0
	if footer != "" {
		footerLine := t.Shell.Footer.Width(innerWidth).Render(footer)
		footerHeight = lipgloss.Height(footerLine)
	}

	bodyHeight := innerHeight - lipgloss.Height(headerLine) - footerHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	bodyContentHeight := bodyHeight - t.Shell.Body.GetVerticalFrameSize()
	if bodyContentHeight < 1 {
		bodyContentHeight = 1
	}
	bodyContentWidth := innerWidth - t.Shell.Body.GetHorizontalFrameSize()
	if bodyContentWidth < 1 {
		bodyContentWidth = 1
	}

	return Metrics{
		InnerWidth:        innerWidth,
		InnerHeight:       innerHeight,
		HeaderHeight:      lipgloss.Height(headerLine),
		FooterHeight:      footerHeight,
		BodyRegionHeight:  bodyHeight,
		BodyContentHeight: bodyContentHeight,
		BodyContentWidth:  effectiveBodyWidth(bodyContentWidth, slots.Body.Layout),
	}
}

// Render draws the shared app shell around body content.
func Render(t theme.Theme, slots Slots, viewport Viewport) tea.View {
	body := normalizeBody(slots.Body.Content)
	header := normalizeHeader(t, slots.Header)
	footer := normalizeFooter(slots.Footer)

	headerLine := t.Shell.HeaderBar.Render(header)
	footerLine := t.Shell.Footer.Render(footer)
	bodyBlock := t.Shell.Body.Render(body)
	sections := []string{headerLine}

	if viewport.Width > 0 && viewport.Height > 0 {
		metrics := Measure(t, slots, viewport)
		headerLine = t.Shell.HeaderBar.Width(metrics.InnerWidth).Render(header)
		footerLine = t.Shell.Footer.Width(metrics.InnerWidth).Render(footer)
		bodyBlock = renderBody(t, slots.Body, metrics.InnerWidth, metrics.BodyRegionHeight)
	}

	sections = append(sections, bodyBlock)
	if footer != "" {
		sections = append(sections, footerLine)
	}

	rendered := t.Shell.Outer.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
	if viewport.Width > 0 && viewport.Height > 0 {
		rendered = t.Shell.Outer.Width(viewport.Width).Height(viewport.Height).Render(
			lipgloss.JoinVertical(lipgloss.Left, sections...),
		)
	}
	return tea.NewView(rendered)
}

func normalizeBody(body string) string {
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return " "
	}
	return body
}

func normalizeHeader(t theme.Theme, header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return t.Shell.HeaderLead.Render(" ")
	}
	return header
}

func normalizeFooter(footer string) string {
	return strings.TrimSpace(footer)
}

func renderBody(t theme.Theme, body BodySlot, width, height int) string {
	layout := body.Layout.withDefaults()
	innerWidth := width - t.Shell.Body.GetHorizontalFrameSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := height - t.Shell.Body.GetVerticalFrameSize()
	if innerHeight < 1 {
		innerHeight = 1
	}

	contentWidth := effectiveBodyWidth(innerWidth, layout)
	if contentWidth < 1 {
		contentWidth = 1
	}

	child := lipgloss.NewStyle().Width(contentWidth)
	if layout.HeightMode == HeightFill {
		child = child.Height(innerHeight)
	}
	content := child.Render(normalizeBody(body.Content))

	if layout.HeightMode == HeightIntrinsic {
		naturalHeight := lipgloss.Height(content)
		if naturalHeight > innerHeight {
			naturalHeight = innerHeight
		}
		if naturalHeight > 0 {
			content = lipgloss.NewStyle().Width(contentWidth).Height(naturalHeight).Render(normalizeBody(body.Content))
		}
	}

	bodyContent := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		MaxHeight(innerHeight).
		AlignVertical(layout.lipglossAlign()).
		Render(content)

	bodyContent = t.Shell.Body.Width(width).Height(height).MaxHeight(height).Render(bodyContent)
	return lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Render(bodyContent)
}

func effectiveBodyWidth(available int, layout BodyLayout) int {
	if available < 1 {
		return 1
	}
	if layout.MaxWidth > 0 && layout.MaxWidth < available {
		return layout.MaxWidth
	}
	return available
}

func (l BodyLayout) withDefaults() BodyLayout {
	if l.WidthMode > WidthIntrinsic {
		l.WidthMode = WidthFill
	}
	if l.HeightMode > HeightIntrinsic {
		l.HeightMode = HeightFill
	}
	if l.VerticalAlign > AlignBottom {
		l.VerticalAlign = AlignTop
	}
	if l.WidthMode == 0 && l.HeightMode == 0 && l.VerticalAlign == 0 && l.MaxWidth == 0 {
		return DefaultBodyLayout()
	}
	return l
}

func (l BodyLayout) lipglossAlign() lipgloss.Position {
	switch l.VerticalAlign {
	case AlignCenter:
		return lipgloss.Center
	case AlignBottom:
		return lipgloss.Bottom
	default:
		return lipgloss.Top
	}
}
