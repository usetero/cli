package policycard

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
)

// viewHeader renders the card identity block:
//
//	payment-gateway / transaction.attempt.logged           ● PENDING
//	Broken Records · Sample to 1 event per 10s
//
//	  45.2k evt/hr · 12.4 MB/hr                           ~$8.7k/yr
func (m *Model) viewHeader() string {
	p := m.policy
	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	var lines []string

	// Line 1: service / event, status badge right-aligned.
	name := p.ServiceName
	if p.LogEventName != "" {
		name += " / " + p.LogEventName
	}
	badge := status.Policy(m.theme, p.Status, true)
	lines = append(lines, alignLeftRight(text.Bold(true).Render(name), badge, m.width, m.theme))

	// Line 2: category display name · subtitle.
	categoryName := p.CategoryDisplayName
	if categoryName == "" {
		categoryName = format.TitleCase(string(p.Category))
	}
	line2 := categoryName
	if p.Analysis != nil {
		if subtitle := p.Analysis.Subtitle(); subtitle != "" {
			line2 += " · " + subtitle
		}
	}
	lines = append(lines, muted.Italic(true).Render(line2))

	// Line 3: volume · bytes left, cost right.
	var volumeParts []string
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
	if p.VolumePerHour != nil {
		volumeParts = append(volumeParts, accent.Render(format.Volume(*p.VolumePerHour)+" evt/hr"))
	}
	if p.BytesPerHour != nil {
		volumeParts = append(volumeParts, accent.Render(format.Bytes(*p.BytesPerHour)+"/hr"))
	}
	left := strings.Join(volumeParts, muted.Render(" · "))

	var right string
	if cost := p.CostPerYear(); cost != "" {
		right = lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg).Render(cost)
	} else if p.CategoryType == domain.CategoryTypeCompliance && p.Severity != "" {
		right = renderSeverityBadge(m.theme, string(p.Severity))
	}

	if metricsLine := alignLeftRight(left, right, m.width, m.theme); metricsLine != "" {
		lines = append(lines, "")
		lines = append(lines, metricsLine)
	}

	return strings.Join(lines, "\n")
}

func renderSeverityBadge(theme styles.Theme, severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return lipgloss.NewStyle().Foreground(theme.Error).Background(theme.Bg).Render("▲ CRITICAL")
	case "high":
		return lipgloss.NewStyle().Foreground(theme.Error).Background(theme.Bg).Render("▲ HIGH")
	case "medium":
		return lipgloss.NewStyle().Foreground(theme.Warning).Background(theme.Bg).Render("▲ MEDIUM")
	case "low":
		return lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.Bg).Render("△ LOW")
	default:
		return ""
	}
}
