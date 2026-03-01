package waste

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appevents "github.com/usetero/cli/internal/app/events"
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
	cursor   int
}

func (d *detail) Len() int        { return len(d.policies) }
func (d *detail) Cursor() int     { return d.cursor }
func (d *detail) SetCursor(v int) { d.cursor = v }

// newDetail creates a detail view for the given category and pre-fetched policies.
func newDetail(theme styles.Theme, category domain.PolicyCategoryStatus, policies []domain.WastePolicy) *detail {
	return &detail{
		theme:    theme,
		category: category,
		policies: policies,
	}
}

// Prompt returns a tea.Cmd that emits a DrawerPrompt for the selected policy.
func (d *detail) Prompt() tea.Cmd {
	if len(d.policies) == 0 {
		return nil
	}
	p := d.policies[d.cursor]
	text := fmt.Sprintf(
		"Pull up the %q policy for the %q log event in the %s service.",
		d.category.Name(), p.LogEventName, p.ServiceName,
	)
	return appevents.DrawerPromptCmd(text)
}

// View renders the detail: a header with category summary, then a policy table.
func (d *detail) View(width int) string {
	var lines []string
	lines = append(lines, d.renderHeader())
	if d.category.Principle != "" {
		lines = append(lines, "")
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		lines = append(lines, muted.Render(d.category.Principle))
	}
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
	name := text.Bold(true).Render(d.category.Name())

	var parts []string
	parts = append(parts, back+" "+name)

	if d.category.PendingCount > 0 {
		warn := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		parts = append(parts, warn.Render("●")+" "+warn.Render(fmt.Sprintf("%d pending", d.category.PendingCount)))
	}

	if d.category.EstimatedCostPerHour != nil && *d.category.EstimatedCostPerHour > 0 {
		parts = append(parts, muted.Render(format.YearlyCostPtr(d.category.EstimatedCostPerHour)))
	}

	return strings.Join(parts, sep)
}

// renderTable renders the per-policy table.
func (d *detail) renderTable(width int) string {
	tbl := table.New(d.theme, table.WithMaxValueWidth(30))

	showVolume := d.category.ReducesVolume()
	if showVolume {
		tbl.Headers("Log Event", "Service", "Volume", "Bytes", "Est. Impact")
	} else {
		tbl.Headers("Log Event", "Service", "Bytes", "Est. Impact")
	}
	tbl.SetWidth(width)

	accent := lipgloss.NewStyle().Foreground(d.theme.Accent).Background(d.theme.Bg)
	dot := lipgloss.NewStyle().Foreground(d.theme.Warning).Background(d.theme.Bg).Render("●")

	for i, p := range d.policies {
		name := p.LogEventName
		if i == d.cursor {
			name = accent.Render("▶ " + name)
		} else {
			name = dot + " " + name
		}

		bytes := "—"
		if p.BytesPerHour != nil {
			bytes = format.Bytes(*p.BytesPerHour) + "/hr"
		}
		savings := format.YearlyCostPtr(p.EstimatedCostPerHour)

		if showVolume {
			vol := "—"
			if p.VolumePerHour != nil {
				vol = format.Volume(*p.VolumePerHour) + " evt/hr"
			}
			tbl.Row(name, p.ServiceName, vol, bytes, savings)
		} else {
			tbl.Row(name, p.ServiceName, bytes, savings)
		}
	}

	return tbl.View()
}
