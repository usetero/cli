// Package catalogstatus renders the catalog health indicator in the status bar.
package catalogstatus

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a catalog status check.
type pollMsg struct{}

// Model renders the catalog health: dot color + service count or discovery phase.
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	summary   domain.CatalogSummary
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

		summary, err := m.db.DatadogAccountStatuses().GetCatalogSummary(ctx)
		if err != nil {
			return m.poll()
		}

		services, err := m.db.ServiceStatuses().ListServiceStatuses(ctx)
		if err != nil {
			services = nil
		}

		key := m.stateKey(summary, services)
		if key != m.lastState {
			m.summary = summary
			m.services = services
			m.hasData = summary.ServiceCount > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.CatalogSummary, services []domain.ServiceStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%d:%d:%d:%s:%.0f:%s:%d",
		s.ReadyForUse, s.ServiceCount, s.ActiveServices,
		s.EventCount, s.AnalyzedCount, s.AnalyzingCount,
		s.DiscoveringCount, s.BrokenServices, s.StaleServices,
		s.WorstStatus, s.PercentComplete, s.LogError,
		len(services))

	for _, svc := range services {
		key += fmt.Sprintf("|%s:%s:%.0f:%d:%.0f:%.0f:%.2f",
			svc.Name, svc.Status, svc.PercentComplete, svc.EventCount,
			svc.VolumePerHour, svc.BytesPerHour, svc.CostPerHourUSD)
	}

	return key
}

// CompactView renders the catalog indicator for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	s := m.summary
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	d := status.ServiceDot(m.theme, s.WorstStatus)

	var readyCount int64
	for _, svc := range m.services {
		if svc.Status == domain.ServiceLogStatusReady {
			readyCount++
		}
	}
	ready := fmt.Sprintf("%d/%d svcs", readyCount, s.ActiveServices)

	var suffix string
	switch s.WorstStatus {
	case domain.ServiceLogStatusDiscovering:
		suffix = " · discovering"
	case domain.ServiceLogStatusAnalyzing:
		suffix = " · analyzing"
	case domain.ServiceLogStatusDisabled, domain.ServiceLogStatusInactive,
		domain.ServiceLogStatusBroken, domain.ServiceLogStatusStale,
		domain.ServiceLogStatusReady:
		// No suffix needed.
	}

	return d + " " + muted.Render(ready+suffix)
}

// ExpandedView renders the detailed catalog status for the drawer.
func (m *Model) ExpandedView(width int) string {
	if !m.hasData {
		return ""
	}

	var lines []string
	lines = append(lines, m.renderSummary())
	lines = append(lines, "")
	lines = append(lines, m.renderServiceTable(width))
	return strings.Join(lines, "\n")
}

// renderSummary renders the top summary line.
func (m *Model) renderSummary() string {
	s := m.summary
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	// Count hidden services.
	var hiddenParts []string
	hidden := make(map[domain.ServiceLogStatus]int)
	for _, svc := range m.services {
		if !isActiveStatus(svc.Status) {
			hidden[svc.Status]++
		}
	}
	for _, st := range []domain.ServiceLogStatus{domain.ServiceLogStatusDisabled, domain.ServiceLogStatusInactive} {
		if n := hidden[st]; n > 0 {
			hiddenParts = append(hiddenParts, fmt.Sprintf("%d %s", n, strings.ToLower(st.String())))
		}
	}

	// Service count with hidden annotation.
	svcLabel := fmt.Sprintf("%d services", s.ActiveServices)
	if len(hiddenParts) > 0 {
		svcLabel += fmt.Sprintf(" (+%s)", strings.Join(hiddenParts, ", "))
	}

	var parts []string
	parts = append(parts, svcLabel)
	parts = append(parts, fmt.Sprintf("%d log events", s.EventCount))
	if s.PercentComplete > 0 {
		parts = append(parts, fmt.Sprintf("%.0f%% coverage", s.PercentComplete))
	}

	// Total cost across all active services.
	var totalCostPerHour float64
	for _, svc := range m.services {
		if isActiveStatus(svc.Status) {
			totalCostPerHour += svc.CostPerHourUSD
		}
	}
	if cost := formatYearlyCost(totalCostPerHour); cost != "$0/yr" {
		parts = append(parts, cost)
	}

	return status.ServiceDot(m.theme, s.WorstStatus) + " " + muted.Render(strings.Join(parts, " · "))
}

// isActiveStatus returns true for statuses worth showing in the table.
func isActiveStatus(s domain.ServiceLogStatus) bool {
	switch s { //nolint:exhaustive // inactive/disabled are the only hidden ones
	case domain.ServiceLogStatusDisabled, domain.ServiceLogStatusInactive:
		return false
	}
	return true
}

// renderServiceTable renders active services and summarizes hidden ones by status.
func (m *Model) renderServiceTable(width int) string {
	if len(m.services) == 0 {
		return lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg).Render("No services")
	}

	var active []domain.ServiceStatus
	for _, svc := range m.services {
		if isActiveStatus(svc.Status) {
			active = append(active, svc)
		}
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Service", "Status", "Events", "Volume", "Bytes", "Cost")
	tbl.SetWidth(width)

	for _, svc := range active {
		tbl.Row(
			m.serviceName(svc),
			m.renderStatus(svc),
			fmt.Sprintf("%d", svc.EventCount),
			formatVolume(svc.VolumePerHour)+"/hr",
			formatBytes(svc.BytesPerHour)+"/hr",
			formatYearlyCost(svc.CostPerHourUSD),
		)
	}

	return tbl.View()
}

// serviceName returns the service name, with extra context for non-ready services.
func (m *Model) serviceName(svc domain.ServiceStatus) string {
	switch svc.Status { //nolint:exhaustive // only special cases need extra context
	case domain.ServiceLogStatusBroken:
		if svc.Error != "" {
			return svc.Name + " — " + svc.Error
		}
	case domain.ServiceLogStatusStale:
		return svc.Name + " — no recent data"
	}
	return svc.Name
}

// renderStatus renders the status badge, with percent for discovering/analyzing.
func (m *Model) renderStatus(svc domain.ServiceStatus) string {
	s := status.Service(m.theme, svc.Status, true)
	if svc.PercentComplete > 0 {
		switch svc.Status { //nolint:exhaustive // only discovering/analyzing show percent
		case domain.ServiceLogStatusDiscovering, domain.ServiceLogStatusAnalyzing:
			muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
			s += " " + muted.Render(fmt.Sprintf("%.0f%%", svc.PercentComplete))
		}
	}
	return s
}

// formatVolume formats events/hr: 892, 12.4k, 2.1M.
func formatVolume(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

// formatBytes formats bytes/hr: 540 B, 12.4 KB, 1.2 MB, 3.4 GB.
func formatBytes(b float64) string {
	abs := math.Abs(b)
	switch {
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", b/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1f MB", b/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1f KB", b/1_000)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// formatYearlyCost formats hourly USD rate as yearly: $0/yr, $1.7k/yr, $110k/yr.
func formatYearlyCost(costPerHour float64) string {
	yearly := costPerHour * 8760
	if yearly < 1 {
		return "$0/yr"
	}
	abs := math.Abs(yearly)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("$%.1fM/yr", yearly/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("$%.1fk/yr", yearly/1_000)
	default:
		return fmt.Sprintf("$%.0f/yr", yearly)
	}
}
