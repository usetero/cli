package waste

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

// detail renders the top pending policies for a single waste category.
type detail struct {
	theme    styles.Theme
	category domain.PolicyCategoryStatus
	policies []domain.WastePolicy
}

// newDetail creates a detail view for the given category and pre-fetched policies.
func newDetail(theme styles.Theme, category domain.PolicyCategoryStatus, policies []domain.WastePolicy) *detail {
	return &detail{
		theme:    theme,
		category: category,
		policies: policies,
	}
}

// View renders the detail: a header with category summary, then a policy table.
func (d *detail) View(width int) string {
	var lines []string
	lines = append(lines, d.renderHeader())
	lines = append(lines, "")

	if len(d.policies) == 0 {
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		lines = append(lines, muted.Render("No pending policies in this category."))
	} else {
		lines = append(lines, d.renderTable(width))
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the back hint + category name + summary.
func (d *detail) renderHeader() string {
	colors := d.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	sep := muted.Render(" · ")

	back := muted.Render("esc ◀")
	name := text.Bold(true).Render(d.category.DisplayName())

	var parts []string
	parts = append(parts, back+" "+name)

	if d.category.PendingCount > 0 {
		warn := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		parts = append(parts, warn.Render("●")+" "+warn.Render(fmt.Sprintf("%d pending", d.category.PendingCount)))
	}

	if d.category.EstimatedCostPerHour > 0 {
		parts = append(parts, muted.Render("~"+format.YearlyCost(d.category.EstimatedCostPerHour)))
	}

	return strings.Join(parts, sep)
}

// renderTable renders the per-policy table.
func (d *detail) renderTable(width int) string {
	tbl := table.New(d.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Log Event", "Service", "Volume", "Est. Savings")
	tbl.SetWidth(width)

	dot := lipgloss.NewStyle().Foreground(d.theme.Warning).Background(d.theme.Bg).Render("●")

	for _, p := range d.policies {
		vol := "—"
		if p.HasVolumes {
			vol = format.Volume(p.VolumePerHour) + " evt/hr"
		}
		savings := "—"
		if p.HasVolumes {
			savings = formatPolicyCost(p)
		}
		tbl.Row(
			dot+" "+p.LogEventName,
			p.ServiceName,
			vol,
			savings,
		)
	}

	return tbl.View()
}

// formatPolicyCost returns the estimated yearly cost for a single policy.
func formatPolicyCost(p domain.WastePolicy) string {
	if p.EstimatedCostPerHour > 0 {
		yearly := p.EstimatedCostPerHour * 8760
		if yearly >= 1 {
			return "~" + format.Cost(yearly) + "/yr"
		}
	}
	return "-"
}
