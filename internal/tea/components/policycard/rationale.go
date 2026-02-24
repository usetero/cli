package policycard

import (
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
)

// viewRationale renders the AI rationale section:
//
//	── Rationale ──────────────────────────────────────────────────
//
//	  This event logs every payment attempt with identical structure...
func (m *Model) viewRationale() string {
	if m.policy.Analysis == nil {
		return ""
	}

	rationale := m.policy.Analysis.Rationale()
	if rationale == "" {
		return ""
	}

	// "Rationale" for waste/quality, "Findings" for compliance.
	label := "Rationale"
	if m.policy.CategoryType == domain.CategoryTypeCompliance {
		label = "Findings"
	}

	text := lipgloss.NewStyle().Foreground(m.theme.Text).Background(m.theme.Bg)

	wrapped := wordWrap(rationale, m.width)
	body := text.Render(wrapped)

	return renderDivider(m.theme, m.width, label) + "\n\n" + body
}
