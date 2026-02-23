package services

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
)

// detail renders the log events for a single service.
type detail struct {
	theme     styles.Theme
	service   domain.ServiceStatus
	logEvents []domain.LogEventStatus
	cursor    int
}

// newDetail creates a detail view for the given service and pre-fetched log events.
func newDetail(theme styles.Theme, service domain.ServiceStatus, logEvents []domain.LogEventStatus) *detail {
	return &detail{
		theme:     theme,
		service:   service,
		logEvents: logEvents,
	}
}

// Prompt returns a tea.Cmd that emits a DrawerPrompt for the selected log event.
func (d *detail) Prompt() tea.Cmd {
	if len(d.logEvents) == 0 {
		return nil
	}
	le := d.logEvents[d.cursor]
	svc := d.service.Name
	text := fmt.Sprintf("Tell me about the %q log event in the %s service.", le.Name, svc)
	return func() tea.Msg { return appmsg.DrawerPrompt{Text: text} }
}

// View renders the detail: a header with service summary, then a log event table.
func (d *detail) View(width int) string {
	var lines []string
	lines = append(lines, d.renderHeader())
	lines = append(lines, "")

	if len(d.logEvents) == 0 {
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		lines = append(lines, muted.Render("No log events discovered for this service."))
	} else {
		lines = append(lines, d.renderTable(width))
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the back hint + service name + summary.
func (d *detail) renderHeader() string {
	colors := d.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	sep := muted.Render(" · ")

	back := muted.Render("esc ◀")
	dot := status.ServiceDot(colors, d.service.Health)
	name := dot + " " + text.Bold(true).Render(d.service.Name)

	var parts []string
	parts = append(parts, back+" "+name)

	if d.service.LogEventCount > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("%d log events", d.service.LogEventCount)))
	}

	if d.service.ServiceCostPerHourVolumeUSD != nil && *d.service.ServiceCostPerHourVolumeUSD > 0 {
		parts = append(parts, muted.Render(format.YearlyCost(*d.service.ServiceCostPerHourVolumeUSD)))
	}

	return strings.Join(parts, sep)
}

// renderTable renders the per-log-event table.
func (d *detail) renderTable(width int) string {
	tbl := table.New(d.theme, table.WithMaxValueWidth(35))
	tbl.Headers("Log Event", "Volume", "Bytes", "Cost")
	tbl.SetWidth(width)

	muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(d.theme.Accent).Background(d.theme.Bg)

	for i, le := range d.logEvents {
		name := le.Name
		if i == d.cursor {
			name = accent.Render("▶ " + name)
		} else if le.PendingPolicyCount > 0 {
			warn := lipgloss.NewStyle().Foreground(d.theme.Warning).Background(d.theme.Bg)
			name = warn.Render("●") + " " + name
		} else {
			name = muted.Render("●") + " " + name
		}

		vol := "—"
		if le.VolumePerHour != nil {
			vol = format.Volume(*le.VolumePerHour) + "/hr"
		}

		bytes := "—"
		if le.BytesPerHour != nil {
			bytes = format.Bytes(*le.BytesPerHour) + "/hr"
		}

		cost := "—"
		if le.CostPerHourUSD != nil {
			cost = format.YearlyCost(*le.CostPerHourUSD)
		}

		tbl.Row(name, vol, bytes, cost)
	}

	return tbl.View()
}
