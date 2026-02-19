// Package services renders the services health indicator in the status bar.
package services

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
)

const (
	pollInterval = 2 * time.Second
	maxServices  = 50

	// discoveryDoneThreshold is the classification coverage percentage above
	// which we consider discovery complete. Volume ratios are never exactly
	// 100% due to throughput fluctuations.
	discoveryDoneThreshold = 95

	// levelDisplayThreshold is the minimum fraction of total volume a
	// non-info level must reach to be shown (1%).
	levelDisplayThreshold = 0.01
)

// pollMsg triggers a catalog status check.
type pollMsg struct{}

// dataMsg carries the result of an async data fetch.
type dataMsg struct {
	summary  domain.AccountSummary
	services []domain.ServiceStatus
	err      error
}

// Model renders the catalog health: dot color + service count + cost.
type Model struct {
	theme styles.Theme
	scope log.Scope
	db    sqlite.DB

	summary   domain.AccountSummary
	services  []domain.ServiceStatus
	hasData   bool
	lastState string
}

// New creates a new catalog status model.
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("services"),
	}
}

// SetDB sets the database and starts polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	m.db = db
	return m.poll()
}

// Init starts polling.
func (m *Model) Init() tea.Cmd {
	if m.db == nil {
		return nil
	}
	return m.poll()
}

func (m *Model) poll() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollMsg{}
	})
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case pollMsg:
		if m.db == nil {
			return nil
		}
		return tea.Batch(m.fetchData(), m.poll())

	case dataMsg:
		if msg.err != nil {
			return nil
		}
		key := m.stateKey(msg.summary, msg.services)
		if key != m.lastState {
			m.summary = msg.summary
			m.services = msg.services
			m.hasData = msg.summary.ActiveServices > 0
			m.lastState = key
		}
	}

	return nil
}

// fetchData returns a Cmd that queries service data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		ctx := context.Background()
		summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			scope.Error("get summary", "err", err)
			return dataMsg{err: err}
		}

		services, err := db.ServiceStatuses().ListEnabledServiceStatuses(ctx, maxServices)
		if err != nil {
			scope.Error("list service statuses", "err", err)
			services = nil
		}

		return dataMsg{summary: summary, services: services}
	}
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.AccountSummary, services []domain.ServiceStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%d:%d:%s:%s:%s:%v:%v:%d",
		s.ReadyForUse, s.ServiceCount, s.ActiveServices,
		s.EventCount, s.AnalyzedCount, s.OkServices,
		s.ErrorServices, s.QuarantinedCount,
		s.Health, s.Error, s.Warning,
		s.TotalServiceVolumePerHour, s.TotalVolumePerHour,
		len(services))

	for _, svc := range services {
		key += fmt.Sprintf("|%s:%s:%s:%d:%d:%v:%v:%v:%v:%v:%v:%v",
			svc.Name, svc.Health, svc.Warning, svc.LogEventCount, svc.LogEventAnalyzedCount,
			svc.ServiceVolumePerHour, svc.LogEventVolumePerHour, svc.ServiceCostPerHourVolumeUSD,
			svc.ServiceDebugVolumePerHour, svc.ServiceInfoVolumePerHour,
			svc.ServiceWarnVolumePerHour, svc.ServiceErrorVolumePerHour)
	}

	return key
}

// HasData returns true when catalog data has been loaded.
func (m *Model) HasData() bool {
	return m.hasData
}

// CompactView renders the catalog indicator for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	s := m.summary
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	d := status.ServiceDot(m.theme, s.Health)

	label := fmt.Sprintf("%d svcs", s.ActiveServices)

	return d + " " + muted.Render(label)
}

// ExpandedView renders the detailed catalog status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		if m.db == nil {
			return muted.Render("Waiting for sync to start...")
		}
		return muted.Render("No services discovered yet.")
	}

	// Height budget: summary (1) + gap (1) + table header+border (2) = 4 lines overhead.
	maxRows := height - 4
	if maxRows < 1 {
		maxRows = 1
	}

	var lines []string
	lines = append(lines, m.renderSummary())
	lines = append(lines, "")
	lines = append(lines, m.renderServiceTable(width, maxRows))
	return strings.Join(lines, "\n")
}

// renderSummary renders the top summary line.
func (m *Model) renderSummary() string {
	s := m.summary
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	var parts []string

	// Health legend: colored dots with counts, matching waste/compliance headline pattern.
	okDot := status.ServiceDot(m.theme, domain.ServiceHealthOK)
	if s.OkServices > 0 {
		parts = append(parts, okDot+" "+muted.Render(fmt.Sprintf("%d ok", s.OkServices)))
	}
	if s.ErrorServices > 0 {
		errDot := status.ServiceDot(m.theme, domain.ServiceHealthError)
		parts = append(parts, errDot+" "+muted.Render(fmt.Sprintf("%d errors", s.ErrorServices)))
	}
	// Hidden services (not in the table).
	var hiddenParts []string
	if s.DisabledServices > 0 {
		hiddenParts = append(hiddenParts, fmt.Sprintf("%d disabled", s.DisabledServices))
	}
	if s.InactiveServices > 0 {
		hiddenParts = append(hiddenParts, fmt.Sprintf("%d inactive", s.InactiveServices))
	}
	if len(hiddenParts) > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("+%s", strings.Join(hiddenParts, ", "))))
	}

	// Log event count with discovery progress.
	if s.EventCount > 0 {
		evtLabel := muted.Render(fmt.Sprintf("%d log events", s.EventCount))
		if pct := summaryDiscoveryPercent(s); pct < discoveryDoneThreshold {
			bar := m.discoveryBar()
			evtLabel += " " + bar.ViewAs(float64(pct)/100) + " " + muted.Render(fmt.Sprintf("%d%%", pct))
		}
		parts = append(parts, evtLabel)
	}

	// Total volume.
	if s.TotalVolumePerHour != nil && *s.TotalVolumePerHour > 0 {
		parts = append(parts, muted.Render(format.Volume(*s.TotalVolumePerHour)+" evt/hr"))
	}

	// Total cost from account summary (pre-aggregated by control plane).
	if s.TotalCostPerHour != nil {
		if cost := format.YearlyCost(*s.TotalCostPerHour); cost != "$0/yr" {
			parts = append(parts, muted.Render(cost))
		}
	}

	sep := muted.Render(" · ")
	result := strings.Join(parts, sep)
	if s.Warning != "" {
		warn := lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg)
		result += " " + warn.Render("⚠")
	}
	return result
}

// renderServiceTable renders enabled services (query already excludes DISABLED/INACTIVE).
func (m *Model) renderServiceTable(width, maxRows int) string {
	if len(m.services) == 0 {
		return lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg).Render("No services")
	}

	// Reserve a row for "+N more" if we need to clip.
	clipped := 0
	visible := m.services
	if len(m.services) > maxRows {
		visible = m.services[:maxRows-1]
		clipped = len(m.services) - len(visible)
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(35))
	tbl.Headers("Service", "Log Events", "Volume", "Cost")
	tbl.SetWidth(width)

	bar := m.discoveryBar()
	for _, svc := range visible {
		cost := "—"
		if svc.ServiceCostPerHourVolumeUSD != nil {
			cost = format.YearlyCost(*svc.ServiceCostPerHourVolumeUSD)
		}
		tbl.Row(
			m.renderServiceName(svc),
			m.renderLogEvents(svc, bar),
			m.renderVolume(svc),
			cost,
		)
	}

	result := tbl.View()
	if clipped > 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		result += "\n" + muted.Render(fmt.Sprintf("+%d more", clipped))
	}
	return result
}

// renderServiceName renders the service name with a colored health dot.
// Warning overrides the dot color to orange.
func (m *Model) renderServiceName(svc domain.ServiceStatus) string {
	dot := status.ServiceDot(m.theme, svc.Health)
	if svc.Warning != "" {
		dot = lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg).Render("●")
	}

	name := svc.Name
	switch svc.Health {
	case domain.ServiceHealthError:
		if svc.Error != "" {
			name += " — " + svc.Error
		}
	case domain.ServiceHealthDisabled, domain.ServiceHealthInactive, domain.ServiceHealthOK:
		// No suffix needed.
	}

	return dot + " " + name
}

// renderVolume renders the service volume with colored level-composition tags.
// Non-info levels >= 1% are shown as compact colored tags (e.g. "e31% w5% d4%").
// Info is omitted since it's the expected baseline — only interesting levels appear.
func (m *Model) renderVolume(svc domain.ServiceStatus) string {
	if svc.ServiceVolumePerHour == nil {
		return "—"
	}
	total := *svc.ServiceVolumePerHour
	vol := fmt.Sprintf("%-9s", format.Volume(total)+"/hr")
	if total <= 0 {
		return vol
	}

	type levelTag struct {
		letter string
		value  *float64
		color  color.Color
	}
	levels := []levelTag{
		{"e", svc.ServiceErrorVolumePerHour, m.theme.Error},
		{"w", svc.ServiceWarnVolumePerHour, m.theme.Warning},
		{"d", svc.ServiceDebugVolumePerHour, m.theme.TextMuted},
	}

	var tags []string
	for _, l := range levels {
		if l.value == nil {
			continue
		}
		rate := *l.value / total
		if rate < levelDisplayThreshold {
			continue
		}
		style := lipgloss.NewStyle().Foreground(l.color).Background(m.theme.Bg)
		tags = append(tags, style.Render(fmt.Sprintf("%s%d%%", l.letter, int(math.Round(rate*100)))))
	}

	if len(tags) > 0 {
		vol += " " + strings.Join(tags, " ")
	}
	return vol
}

// discoveryBar creates a small progress bar for inline use in table cells.
func (m *Model) discoveryBar() progress.Model {
	bar := progress.New(
		progress.WithColors(m.theme.GradientStart, m.theme.GradientEnd),
		progress.WithWidth(10),
		progress.WithFillCharacters('█', '░'),
	)
	bar.ShowPercentage = false
	bar.EmptyColor = m.theme.TextMuted
	return bar
}

// renderLogEvents renders the log event count with a mini discovery progress bar.
// The bar is hidden when classification coverage is >= 95% (volume ratios are
// never exactly 100% due to throughput fluctuations).
func (m *Model) renderLogEvents(svc domain.ServiceStatus, bar progress.Model) string {
	count := fmt.Sprintf("%-5d", svc.LogEventCount)
	if svc.ServiceVolumePerHour == nil || *svc.ServiceVolumePerHour <= 0 {
		return count
	}
	pct := discoveryPercent(svc)
	if pct >= discoveryDoneThreshold {
		return count
	}
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	return count + " " + bar.ViewAs(float64(pct)/100) + " " + muted.Render(fmt.Sprintf("%3d%%", pct))
}

// summaryDiscoveryPercent computes account-level classification coverage.
func summaryDiscoveryPercent(s domain.AccountSummary) int {
	if s.TotalServiceVolumePerHour == nil || *s.TotalServiceVolumePerHour <= 0 {
		return 100
	}
	if s.TotalVolumePerHour == nil {
		return 0
	}
	pct := int(math.Round(*s.TotalVolumePerHour / *s.TotalServiceVolumePerHour * 100))
	if pct > 100 {
		pct = 100
	}
	return pct
}

// discoveryPercent computes how much of a service's volume has been classified.
func discoveryPercent(svc domain.ServiceStatus) int {
	if svc.ServiceVolumePerHour == nil || *svc.ServiceVolumePerHour <= 0 {
		return 100
	}
	if svc.LogEventVolumePerHour == nil {
		return 0
	}
	pct := int(math.Round(*svc.LogEventVolumePerHour / *svc.ServiceVolumePerHour * 100))
	if pct > 100 {
		pct = 100
	}
	return pct
}
