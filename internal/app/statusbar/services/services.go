// Package services renders the services health indicator in the status bar.
package services

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
	"github.com/usetero/cli/internal/app/statusbar/viewkit"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
)

const (
	pollInterval = 2 * time.Second
	fetchTimeout = 2 * time.Second
	pollSource   = "services"

	// levelDisplayThreshold is the minimum fraction of total volume a
	// non-info level must reach to be shown (1%).
	levelDisplayThreshold = 0.01
)

type fetchedData struct {
	summary  domain.AccountSummary
	services []domain.ServiceStatus
}

// serviceDetailLoadedMsg carries the result of an async detail fetch.
type serviceDetailLoadedMsg struct {
	service   domain.ServiceStatus
	logEvents []domain.LogEventStatus
	err       error
}

// Model renders the catalog health: dot color + service count + cost.
type Model struct {
	theme styles.Theme
	scope log.Scope

	api   graphql.ServiceSet
	ready bool

	summary   domain.AccountSummary
	services  []domain.ServiceStatus
	hasData   bool
	lastState string
	fetching  bool

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

// SetServices points the tab at the account-scoped services and starts polling.
func (m *Model) SetServices(services graphql.ServiceSet) tea.Cmd {
	m.api = services
	m.ready = true
	return m.poll()
}

// Init starts polling once the services are available.
func (m *Model) Init() tea.Cmd {
	if !m.ready {
		return nil
	}
	return m.poll()
}

func (m *Model) poll() tea.Cmd {
	return tabpoll.Tick(pollSource, pollInterval)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if cmd, handled := tabpoll.UpdatePollCycle(
		msg,
		pollSource,
		m.ready,
		&m.fetching,
		m.fetchData(),
		m.poll(),
		func(data fetchedData) {
			key := m.buildStateKey(data.summary, data.services)
			tabpoll.ApplyIfChanged(&m.lastState, key, &m.cursor, len(data.services), func() {
				m.summary = data.summary
				m.services = data.services
				m.hasData = data.summary.ServiceCount > 0
			})
		},
	); handled {
		return cmd
	}

	switch msg := msg.(type) {
	case serviceDetailLoadedMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.service, msg.logEvents)
		}
	}

	return nil
}

// fetchData returns a Cmd that queries service data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	services := m.api
	scope := m.scope
	return tabpoll.Fetch(fetchTimeout, func(ctx context.Context) (fetchedData, error) {
		summary, err := services.Status.GetAccountSummary(ctx)
		if err != nil {
			scope.Error("get summary", "err", err)
			return fetchedData{}, err
		}

		statuses, err := services.Status.ListServiceStatuses(ctx)
		if err != nil {
			scope.Error("list service statuses", "err", err)
			statuses = nil
		}

		return fetchedData{summary: summary, services: statuses}, nil
	})
}

// fetchDetail returns a Cmd that queries log event detail off the event loop.
func (m *Model) fetchDetail(svc domain.ServiceStatus) tea.Cmd {
	services := m.api
	scope := m.scope
	return tabpoll.FetchDetail(fetchTimeout, func(ctx context.Context) ([]domain.LogEventStatus, error) {
		logEvents, err := services.Status.ListServiceLogEvents(ctx, svc.ID)
		if err != nil {
			scope.Error("list log event statuses", "service", svc.Name, "err", err)
			return nil, err
		}
		return logEvents, nil
	}, func(logEvents []domain.LogEventStatus, err error) tea.Msg {
		if err != nil {
			return serviceDetailLoadedMsg{err: err}
		}
		return serviceDetailLoadedMsg{service: svc, logEvents: logEvents}
	})
}

// HandleKeyPress handles keyboard navigation in the expanded drawer view.
func (m *Model) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	return m.navController().HandleKeyPress(msg)
}

// InDetail returns true when the detail sub-view is active.
func (m *Model) InDetail() bool {
	return m.detail != nil
}

// CloseDetail exits the detail sub-view, returning to the service list.
func (m *Model) CloseDetail() {
	m.detail = nil
}

func (m *Model) navController() listdetail.Controller {
	return listdetail.New(
		func() bool { return m.hasData && len(m.services) > 0 },
		func() int { return m.cursor },
		func(v int) { m.cursor = v },
		func() int { return len(m.services) },
		func(index int) tea.Cmd {
			return m.fetchDetail(m.services[index])
		},
		func() listdetail.Detail { return m.detail },
		func() { m.detail = nil },
	)
}

// buildStateKey builds a string key for change detection.
func (m *Model) buildStateKey(s domain.AccountSummary, services []domain.ServiceStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%s:%v:%v:%d",
		s.ReadyForUse, s.ServiceCount, s.ActiveServices,
		s.EventCount, s.AnalyzedCount, s.OkServices,
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
		return viewkit.RenderServicesEmptyState(
			m.theme,
			m.ready,
			m.summary,
			"Ask Tero to explore your services and pick which ones to enable.",
		)
	}

	// Detail sub-view for a single service.
	if m.detail != nil {
		return m.detail.View(width)
	}

	// All services disabled — show a clear message with guidance.
	if m.summary.ActiveServices == 0 {
		return viewkit.RenderServicesEmptyState(
			m.theme,
			true,
			m.summary,
			"Ask Tero to explore your services and pick which ones to enable.",
		)
	}

	// Height budget: summary (1) + gap (1) + table header+border (2) = 4 lines overhead.
	maxRows := height - 4
	if maxRows < 1 {
		maxRows = 1
	}

	return viewkit.ComposeSummaryTableView(
		m.theme,
		m.renderSummary(),
		m.renderServiceTable(width, maxRows),
		"",
	)
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
