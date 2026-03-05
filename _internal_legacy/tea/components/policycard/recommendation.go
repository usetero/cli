package policycard

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// viewRecommendation renders the action section:
//
//	── Recommendation ─────────────────────────────────────────────
//
//	  Sample — keep ~1% of volume
//
//	  Approving keeps 1 in every ~125 events.
func (m *Model) viewRecommendation() string {
	p := m.policy
	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	var lines []string

	lines = append(lines, renderDivider(m.theme, m.width, "Recommendation"))

	lines = append(lines, "")
	lines = append(lines, text.Bold(true).Render(p.Headline()))

	if mechanism := p.Mechanism(); mechanism != "" {
		lines = append(lines, "")
		lines = append(lines, muted.Render(wordWrap(mechanism, m.width)))
	}

	return strings.Join(lines, "\n")
}
