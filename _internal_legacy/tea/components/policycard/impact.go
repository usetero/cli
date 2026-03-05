package policycard

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// viewImpact renders the before/after metrics table:
//
//	── Impact ─────────────────────────────────────────────────────
//
//	  Volume    45.2k evt/hr  →  360 evt/hr                -99.2%
//	  Storage   12.4 MB/hr    →  99.2 KB/hr                -99.2%
//	  Savings                                            ~$8.7k/yr
func (m *Model) viewImpact() string {
	if m.impact == nil {
		return ""
	}

	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	success := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg)

	var lines []string

	lines = append(lines, renderDivider(m.theme, m.width, "Impact"))
	lines = append(lines, "")

	if m.impact.VolumeFrom != "" {
		left := muted.Render("Volume   ") + text.Render(m.impact.VolumeFrom) + muted.Render("  →  ") + text.Render(m.impact.VolumeTo)
		var right string
		if m.impact.VolumePct != "" {
			right = success.Render(m.impact.VolumePct)
		}
		lines = append(lines, alignLeftRight(left, right, m.width, m.theme))
	}

	if m.impact.StorageFrom != "" {
		left := muted.Render("Storage  ") + text.Render(m.impact.StorageFrom) + muted.Render("  →  ") + text.Render(m.impact.StorageTo)
		var right string
		if m.impact.StoragePct != "" {
			right = success.Render(m.impact.StoragePct)
		}
		lines = append(lines, alignLeftRight(left, right, m.width, m.theme))
	}

	if m.impact.Savings != "" {
		left := muted.Render("Savings")
		right := success.Bold(true).Render(m.impact.Savings)
		lines = append(lines, alignLeftRight(left, right, m.width, m.theme))
	}

	return strings.Join(lines, "\n")
}
