// Package services renders the services health indicator in the status bar.
package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
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
)

// pollMsg triggers a catalog status check.
type pollMsg struct{}

// Model renders the catalog health: dot color + service count + cost.
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	summary   domain.AccountSummary
	services  []domain.ServiceStatus
	hasData   bool
	lastState string
}

// New creates a new catalog status model.
func New(theme styles.Theme) *Model {
	return &Model{
		theme: theme,
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
	switch msg.(type) {
	case pollMsg:
		if m.db == nil {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		summary, err := m.db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			return m.poll()
		}

		services, err := m.db.ServiceStatuses().ListEnabledServiceStatuses(ctx, maxServices)
		if err != nil {
			services = nil
		}

		key := m.stateKey(summary, services)
		if key != m.lastState {
			m.summary = summary
			m.services = services
			m.hasData = summary.ActiveServices > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.AccountSummary, services []domain.ServiceStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%d:%d:%d:%s:%s:%s:%.0f:%.0f:%d",
		s.ReadyForUse, s.ServiceCount, s.ActiveServices,
		s.EventCount, s.AnalyzedCount, s.OkServices,
		s.ErrorServices, s.StaleServices, s.QuarantinedCount,
		s.Health, s.Error, s.Warning,
		ptrVal(s.TotalServiceVolumePerHour), s.TotalVolumePerHour,
		len(services))

	for _, svc := range services {
		key += fmt.Sprintf("|%s:%s:%s:%d:%d:%.0f:%.0f:%.2f",
			svc.Name, svc.Health, svc.Warning, svc.LogEventCount, svc.LogEventAnalyzedCount,
			svc.ServiceVolumePerHour, svc.LogEventVolumePerHour, svc.ServiceCostPerHourVolumeUSD)
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

	// Service count with hidden annotation from summary counts.
	var hiddenParts []string
	if s.DisabledServices > 0 {
		hiddenParts = append(hiddenParts, fmt.Sprintf("%d disabled", s.DisabledServices))
	}
	if s.InactiveServices > 0 {
		hiddenParts = append(hiddenParts, fmt.Sprintf("%d inactive", s.InactiveServices))
	}
	svcLabel := fmt.Sprintf("%d services", s.ActiveServices)
	if len(hiddenParts) > 0 {
		svcLabel += fmt.Sprintf(" (+%s)", strings.Join(hiddenParts, ", "))
	}

	var parts []string
	parts = append(parts, svcLabel)

	// Log event count with discovery progress.
	if s.EventCount > 0 {
		evtLabel := fmt.Sprintf("%d log events", s.EventCount)
		if pct := summaryDiscoveryPercent(s); pct < discoveryDoneThreshold {
			evtLabel += fmt.Sprintf(" (%d%%)", pct)
		}
		parts = append(parts, evtLabel)
	}

	// Total volume.
	if s.TotalVolumePerHour > 0 {
		parts = append(parts, format.Volume(s.TotalVolumePerHour)+" evt/hr")
	}

	// Total cost from account summary (pre-aggregated by control plane).
	if s.TotalCostPerHour != nil {
		if cost := format.YearlyCost(*s.TotalCostPerHour); cost != "$0/yr" {
			parts = append(parts, cost)
		}
	}

	result := status.ServiceDot(m.theme, s.Health) + " " + muted.Render(strings.Join(parts, " · "))
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

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Service", "Log Events", "Volume", "Cost")
	tbl.SetWidth(width)

	bar := m.discoveryBar()
	for _, svc := range visible {
		tbl.Row(
			m.renderServiceName(svc),
			m.renderLogEvents(svc, bar),
			format.Volume(svc.ServiceVolumePerHour)+"/hr",
			format.YearlyCost(svc.ServiceCostPerHourVolumeUSD),
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
	case domain.ServiceHealthStale:
		name += " — no recent data"
	case domain.ServiceHealthDisabled, domain.ServiceHealthInactive, domain.ServiceHealthOK:
		// No suffix needed.
	}

	return dot + " " + name
}

// discoveryBar creates a small progress bar for inline use in table cells.
func (m *Model) discoveryBar() progress.Model {
	bar := progress.New(
		progress.WithColors(m.theme.GradientStart),
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
	if svc.ServiceVolumePerHour <= 0 {
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
	pct := int(math.Round(s.TotalVolumePerHour / *s.TotalServiceVolumePerHour * 100))
	if pct > 100 {
		pct = 100
	}
	return pct
}

// discoveryPercent computes how much of a service's volume has been classified.
func discoveryPercent(svc domain.ServiceStatus) int {
	if svc.ServiceVolumePerHour <= 0 {
		return 100
	}
	pct := int(math.Round(svc.LogEventVolumePerHour / svc.ServiceVolumePerHour * 100))
	if pct > 100 {
		pct = 100
	}
	return pct
}

func ptrVal(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
