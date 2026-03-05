package policycard

import (
	"charm.land/lipgloss/v2"
)

// viewRationale renders the AI rationale as flowing body text after the header.
// No divider — this is the natural continuation of "why you're looking at this".
func (m *Model) viewRationale() string {
	if m.policy.Analysis == nil {
		return ""
	}

	rationale := m.policy.Analysis.Rationale()
	if rationale == "" {
		return ""
	}

	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)

	return text.Render(wordWrap(rationale, m.width))
}
