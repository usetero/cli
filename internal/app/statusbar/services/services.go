// Package services renders the services health indicator in the status bar.
package services

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
	"github.com/usetero/cli/internal/tea/keymap"
)

const (
	pollInterval = 2 * time.Second
	maxServices  = 50

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

// detailMsg carries the result of an async detail fetch.
type detailMsg struct {
	service   domain.ServiceStatus
	logEvents []domain.LogEventStatus
	err       error
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

	// Drawer navigation
	cursor int     // selected row in service list
	detail *detail // non-nil when viewing a single service's log events
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
			m.hasData = msg.summary.ServiceCount > 0
			m.lastState = key

			// Clamp cursor if services shrank.
			if m.cursor >= len(m.services) && len(m.services) > 0 {
				m.cursor = len(m.services) - 1
			}
		}

	case detailMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.service, msg.logEvents)
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

// fetchDetail returns a Cmd that queries log event detail off the event loop.
func (m *Model) fetchDetail(svc domain.ServiceStatus) tea.Cmd {
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		logEvents, err := db.LogEventStatuses().ListByService(context.Background(), svc.Name, 25)
		if err != nil {
			scope.Error("list log event statuses", "service", svc.Name, "err", err)
			return detailMsg{err: err}
		}
		return detailMsg{service: svc, logEvents: logEvents}
	}
}

// HandleKeyPress handles keyboard navigation in the expanded drawer view.
func (m *Model) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if !m.hasData || len(m.services) == 0 {
		return nil
	}

	// Detail mode: navigate log events or go back.
	if m.detail != nil {
		switch {
		case key.Matches(msg, keymap.DrawerBack):
			m.detail = nil
		case key.Matches(msg, keymap.DrawerUp):
			if m.detail.cursor > 0 {
				m.detail.cursor--
			}
		case key.Matches(msg, keymap.DrawerDown):
			if m.detail.cursor < len(m.detail.logEvents)-1 {
				m.detail.cursor++
			}
		case key.Matches(msg, keymap.DrawerSelect):
			return m.detail.Prompt()
		}
		return nil
	}

	// Service list mode.
	switch {
	case key.Matches(msg, keymap.DrawerUp):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keymap.DrawerDown):
		if m.cursor < len(m.services)-1 {
			m.cursor++
		}
	case key.Matches(msg, keymap.DrawerSelect):
		return m.fetchDetail(m.services[m.cursor])
	}
	return nil
}

// InDetail returns true when the detail sub-view is active.
func (m *Model) InDetail() bool {
	return m.detail != nil
}

// CloseDetail exits the detail sub-view, returning to the service list.
func (m *Model) CloseDetail() {
	m.detail = nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.AccountSummary, services []domain.ServiceStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%d:%s:%v:%v:%d",
		s.ReadyForUse, s.ServiceCount, s.ActiveServices,
		s.EventCount, s.AnalyzedCount, s.OkServices,
		s.QuarantinedCount,
		s.Health,
		s.TotalServiceVolumePerHour, s.TotalVolumePerHour,
		len(services))

	for _, svc := range services {
		key += fmt.Sprintf("|%s:%s:%d:%d:%v:%v:%v:%v:%v:%v:%v",
			svc.Name, svc.Health, svc.LogEventCount, svc.LogEventAnalyzedCount,
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

	if s.ActiveServices == 0 {
		dot := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg).Render("●")
		return dot + " " + muted.Render(fmt.Sprintf("%d svcs (all disabled)", s.ServiceCount))
	}

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

	// Detail sub-view for a single service.
	if m.detail != nil {
		return m.detail.View(width)
	}

	// All services disabled — show a clear message with guidance.
	if m.summary.ActiveServices == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		return muted.Render(fmt.Sprintf(
			"%d services discovered, all disabled.\nAsk Tero to explore your services and pick which ones to enable.",
			m.summary.ServiceCount,
		))
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

	// Log event count.
	if s.EventCount > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("%d log events", s.EventCount)))
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
	return strings.Join(parts, sep)
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

	for i, svc := range visible {
		cost := "—"
		if svc.LogEventCostPerHourUSD != nil {
			cost = format.YearlyCost(*svc.LogEventCostPerHourUSD)
		}
		tbl.Row(
			m.renderServiceName(i, svc),
			fmt.Sprintf("%-5d", svc.LogEventCount),
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

// renderServiceName renders the service name with a colored health dot or cursor arrow.
func (m *Model) renderServiceName(index int, svc domain.ServiceStatus) string {
	name := svc.Name

	if index == m.cursor {
		accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
		return accent.Render("▶ " + name)
	}

	dot := status.ServiceDot(m.theme, svc.Health)
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
