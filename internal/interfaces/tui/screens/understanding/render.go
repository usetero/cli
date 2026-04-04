package understanding

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
)

const (
	minScreenWidth       = 48
	serviceLabelWidth    = 15
	minGlyphFieldWidth   = 18
	rowMetricWidth       = 9
	serviceDividerGlyph  = "│"
)

type serviceRow struct {
	name     string
	glyphs   []glyphCell
	volumeHr int
}

type glyphCell struct {
	text  string
	tone  glyphTone
}

type glyphTone uint8

const (
	toneMuted glyphTone = iota
	toneNormal
	toneWaste
	toneSpike
)

func (m *Model) View() tea.View {
	width := m.width
	if width < minScreenWidth {
		width = minScreenWidth
	}

	return present.View(
		lipgloss.NewStyle().
			Width(width).
			Render(m.renderEstate(width)),
	)
}

func (m *Model) renderEstate(width int) string {
	services := generateServices(serviceCount(m.height))
	rows := make([]string, 0, len(services))
	for i, name := range services {
		rows = append(rows, m.renderServiceRow(width, generateServiceRow(i, name, width)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderServiceRow(width int, row serviceRow) string {
	name := ansi.Truncate(row.name, serviceLabelWidth, "…")
	labelStyle := lipgloss.NewStyle().
		Inherit(m.theme.Text.Body).
		Width(serviceLabelWidth).
		Background(m.theme.Background)

	dividerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Border).
		Background(m.theme.Background)

	metricStyle := lipgloss.NewStyle().
		Inherit(m.theme.Text.Subtle).
		Width(rowMetricWidth).
		AlignHorizontal(lipgloss.Left).
		Background(m.theme.Background)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(name),
		" ",
		dividerStyle.Render(serviceDividerGlyph),
		" ",
		m.renderGlyphBand(width, row.glyphs),
		" ",
		metricStyle.Render(formatCompactCount(row.volumeHr)+"/hr"),
	)
}

func (m *Model) renderGlyphBand(width int, glyphs []glyphCell) string {
	rendered := make([]string, 0, len(glyphs))
	for _, glyph := range glyphs {
		rendered = append(rendered, m.renderGlyph(glyph))
	}

	content := strings.Join(rendered, " ")
	return lipgloss.NewStyle().
		Width(glyphFieldWidth(width)).
		Background(m.theme.Background).
		Render(ansi.Truncate(content, glyphFieldWidth(width), "…"))
}

func (m *Model) renderGlyph(cell glyphCell) string {
	style := lipgloss.NewStyle().
		Inherit(m.theme.Text.Subtle).
		Background(m.theme.Background)

	switch cell.tone {
	case toneNormal:
		style = lipgloss.NewStyle().
			Foreground(m.theme.Palette.Brand).
			Background(m.theme.Background)
	case toneWaste:
		style = lipgloss.NewStyle().
			Inherit(m.theme.Text.Warning).
			Background(m.theme.Background)
	case toneSpike:
		style = lipgloss.NewStyle().
			Inherit(m.theme.Text.Error).
			Background(m.theme.Background).
			Bold(true)
	}

	return style.Render(cell.text)
}

func glyphFieldWidth(width int) int {
	fieldWidth := width - serviceLabelWidth - 3 - 1 - rowMetricWidth
	if fieldWidth < 1 {
		return 1
	}
	if fieldWidth < minGlyphFieldWidth {
		return fieldWidth
	}
	return fieldWidth
}

func serviceCount(height int) int {
	switch {
	case height >= 42:
		return 30
	case height >= 34:
		return 24
	case height >= 26:
		return 18
	default:
		return 14
	}
}

func generateServices(count int) []string {
	names := make([]string, 0, count)
	domains := []string{
		"checkout", "identity", "search", "billing", "session", "fraud", "email",
		"reporting", "inventory", "support", "audit", "analytics", "catalog", "edge",
		"event", "queue", "db", "cache", "ops", "admin", "payments", "kafka", "worker",
		"ml", "profile", "gateway", "notify", "token", "access", "runtime",
	}
	suffixes := []string{
		"api", "worker", "gateway", "stream", "router", "proxy", "sync", "cache", "bridge", "engine",
	}
	for i := 0; len(names) < count; i++ {
		domain := domains[i%len(domains)]
		suffix := suffixes[(i/len(domains))%len(suffixes)]
		name := domain + "-" + suffix
		if i >= len(domains)*len(suffixes) {
			name = fmt.Sprintf("%s-%s-%02d", domain, suffix, i/(len(domains)*len(suffixes)))
		}
		names = append(names, name)
	}
	return names
}

func generateServiceRow(index int, name string, width int) serviceRow {
	cells := make([]glyphCell, 0, glyphCount(width, index))
	count := glyphCount(width, index)
	volumeHr := 0
	for i := 0; i < count; i++ {
		cell := generateGlyphCell(index, i)
		cells = append(cells, cell)
		volumeHr += glyphVolume(cell)
	}
	return serviceRow{
		name:     name,
		glyphs:   cells,
		volumeHr: volumeHr,
	}
}

func glyphCount(width int, index int) int {
	base := glyphFieldWidth(width) / 2
	if base < 24 {
		base = 24
	}
	return base + (index % 15)
}

func generateGlyphCell(rowIndex int, cellIndex int) glyphCell {
	v := (rowIndex*13 + cellIndex*7 + rowIndex*cellIndex) % 31
	switch {
	case v == 0 || v == 3:
		return glyphCell{text: "●", tone: toneSpike}
	case v == 6 || v == 9:
		return glyphCell{text: "•", tone: toneSpike}
	case v == 5 || v == 8 || v == 11 || v == 14 || v == 21:
		return glyphCell{text: "●", tone: toneWaste}
	case v == 18 || v == 24:
		return glyphCell{text: "•", tone: toneWaste}
	case v%2 == 0:
		return glyphCell{text: "●", tone: toneNormal}
	default:
		return glyphCell{text: "·", tone: toneMuted}
	}
}

func glyphVolume(cell glyphCell) int {
	switch cell.text {
	case "●":
		return 18000
	case "•":
		return 6000
	default:
		return 1200
	}
}

func formatCount(count int, singular string, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	return formatCompactCount(count) + " " + plural
}

func formatCompactCount(count int) string {
	switch {
	case count >= 1000000:
		return strings.TrimSuffix(strings.TrimSuffix(
			strings.TrimRight(strings.TrimRight(formatFloat(float64(count)/1000000), "0"), "."),
			".0",
		), ".0") + "m"
	case count >= 1000:
		return strings.TrimSuffix(strings.TrimSuffix(
			strings.TrimRight(strings.TrimRight(formatFloat(float64(count)/1000), "0"), "."),
			".0",
		), ".0") + "k"
	default:
		return formatFloat(float64(count))
	}
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", value), "0"), ".")
}
