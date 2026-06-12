// Package surfaces renders high-level product surfaces in the status drawer.
package surfaces

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

const (
	pollInterval = 2 * time.Second
	fetchTimeout = 2 * time.Second

	// edgeConnectedWindow is how recently an edge instance must have synced to
	// count as connected.
	edgeConnectedWindow = 10 * time.Minute
)

type fetchFunc func(context.Context, graphql.ServiceSet) (Snapshot, error)

// Metric is one line of supporting state for a product surface.
type Metric struct {
	Label string
	Value string
	Tone  string
}

// Snapshot is the presentation state for one product surface.
type Snapshot struct {
	Title       string
	Description string
	Primary     Metric
	Metrics     []Metric
	Rows        [][]string
	Loaded      bool
}

// Model renders a non-interactive product surface tab.
type Model struct {
	theme styles.Theme
	scope log.Scope

	services graphql.ServiceSet
	ready    bool

	source string
	fetch  fetchFunc

	snapshot  Snapshot
	hasData   bool
	fetching  bool
	lastState string
}

// NewIssues creates the issues surface.
func NewIssues(theme styles.Theme, scope log.Scope) *Model {
	return newModel(theme, scope, "issues", fetchIssues)
}

// NewChecks creates the checks surface.
func NewChecks(theme styles.Theme, scope log.Scope) *Model {
	return newModel(theme, scope, "checks", fetchChecks)
}

// NewLogEvents creates the log events surface.
func NewLogEvents(theme styles.Theme, scope log.Scope) *Model {
	return newModel(theme, scope, "log-events", fetchLogEvents)
}

// NewEdgeInstances creates the edge instances surface.
func NewEdgeInstances(theme styles.Theme, scope log.Scope) *Model {
	return newModel(theme, scope, "edge-instances", fetchEdgeInstances)
}

func newModel(theme styles.Theme, scope log.Scope, source string, fetch fetchFunc) *Model {
	return &Model{
		theme:  theme,
		scope:  scope.Child(source),
		source: source,
		fetch:  fetch,
	}
}

// SetServices points the surface at the account-scoped services and starts polling.
func (m *Model) SetServices(services graphql.ServiceSet) tea.Cmd {
	m.services = services
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
	return tabpoll.Tick(m.source, pollInterval)
}

// Update handles polling messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if cmd, handled := tabpoll.UpdatePollCycle(
		msg,
		m.source,
		m.ready,
		&m.fetching,
		m.fetchData(),
		m.poll(),
		func(snapshot Snapshot) {
			key := snapshotKey(snapshot)
			tabpoll.ApplyIfChanged(&m.lastState, key, nil, 0, func() {
				m.snapshot = snapshot
				m.hasData = snapshot.Loaded
			})
		},
	); handled {
		return cmd
	}
	return nil
}

func (m *Model) fetchData() tea.Cmd {
	services := m.services
	fetch := m.fetch
	scope := m.scope
	return tabpoll.Fetch(fetchTimeout, func(ctx context.Context) (Snapshot, error) {
		snapshot, err := fetch(ctx, services)
		if err != nil {
			scope.Error("fetch surface", "err", err)
			return Snapshot{}, err
		}
		return snapshot, nil
	})
}

// HasData returns true when the tab has loaded a runtime snapshot.
func (m *Model) HasData() bool {
	return m.hasData
}

// CompactView renders the surface's compact status bar signal.
func (m *Model) CompactView() string {
	if !m.hasData || m.snapshot.Primary.Value == "" || m.snapshot.Primary.Value == "0" {
		return ""
	}

	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	return muted.Render(m.snapshot.Primary.Value + " " + strings.ToLower(m.snapshot.Title))
}

// ExpandedView renders the surface drawer body.
func (m *Model) ExpandedView(width, _ int) string {
	colors := m.theme
	if !m.hasData {
		return lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg).Render("Waiting for synced data...")
	}

	titleStyle := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg).Bold(true)
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	lines := []string{titleStyle.Render(m.snapshot.Title)}
	if m.snapshot.Description != "" {
		lines = append(lines, muted.Render(m.snapshot.Description))
	}
	if len(m.snapshot.Metrics) > 0 {
		lines = append(lines, "", m.renderMetrics())
	}
	if len(m.snapshot.Rows) > 0 {
		lines = append(lines, "", m.renderRows(width))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderMetrics() string {
	parts := make([]string, 0, len(m.snapshot.Metrics)+1)
	if m.snapshot.Primary.Label != "" {
		parts = append(parts, m.renderMetric(m.snapshot.Primary))
	}
	for _, metric := range m.snapshot.Metrics {
		parts = append(parts, m.renderMetric(metric))
	}
	return strings.Join(parts, lipgloss.NewStyle().Foreground(m.theme.TextSubtle).Background(m.theme.Bg).Render("  "))
}

func (m *Model) renderMetric(metric Metric) string {
	label := lipgloss.NewStyle().Foreground(m.theme.TextSubtle).Background(m.theme.Bg).Render(metric.Label)
	valueStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	switch metric.Tone {
	case "danger":
		valueStyle = valueStyle.Foreground(m.theme.Error).Bold(true)
	case "warning":
		valueStyle = valueStyle.Foreground(m.theme.Warning).Bold(true)
	case "success":
		valueStyle = valueStyle.Foreground(m.theme.Success).Bold(true)
	}
	return valueStyle.Render(metric.Value) + " " + label
}

func (m *Model) renderRows(width int) string {
	tbl := table.New(m.theme, table.WithMaxValueWidth(36))
	tbl.Headers("Area", "State", "Signal")
	tbl.SetWidth(width)
	for _, row := range m.snapshot.Rows {
		tbl.Row(row...)
	}
	return tbl.View()
}

func fetchIssues(ctx context.Context, services graphql.ServiceSet) (Snapshot, error) {
	summary, err := services.Issues.GetSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Title:       "Issues",
		Description: "Active issues awaiting operator judgment.",
		Primary:     Metric{Label: "open", Value: count(summary.Open), Tone: pendingTone(summary.Open)},
		Metrics: []Metric{
			{Label: "high", Value: count(summary.HighCount), Tone: highTone(summary.HighCount)},
			{Label: "medium", Value: count(summary.MediumCount), Tone: pendingTone(summary.MediumCount)},
			{Label: "low", Value: count(summary.LowCount)},
		},
		Rows: [][]string{
			{"High", count(summary.HighCount), "high-priority review queue"},
			{"Medium", count(summary.MediumCount), "normal review queue"},
			{"Low", count(summary.LowCount), "background cleanup"},
		},
		Loaded: true,
	}, nil
}

func fetchChecks(ctx context.Context, services graphql.ServiceSet) (Snapshot, error) {
	catalog, err := services.Checks.List(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	var openFindings, activeIssues int64
	for _, check := range catalog.Checks {
		openFindings += check.OpenFindingCount
		activeIssues += check.ActiveIssueCount
	}

	return Snapshot{
		Title:       "Checks",
		Description: "Product checks and their account-scoped posture.",
		Primary:     Metric{Label: "checks", Value: count(catalog.Total)},
		Metrics: []Metric{
			{Label: "cost", Value: count(catalog.DomainCount(domain.CheckDomainCost))},
			{Label: "compliance", Value: count(catalog.DomainCount(domain.CheckDomainCompliance))},
			{Label: "active issues", Value: count(activeIssues), Tone: pendingTone(activeIssues)},
		},
		Rows: [][]string{
			{"Cost", checkDomainSummary(catalog, domain.CheckDomainCost), "spend-reduction checks"},
			{"Compliance", checkDomainSummary(catalog, domain.CheckDomainCompliance), "data-protection checks"},
			{"Open findings", count(openFindings), "across all checks"},
		},
		Loaded: true,
	}, nil
}

func fetchLogEvents(ctx context.Context, services graphql.ServiceSet) (Snapshot, error) {
	summary, err := services.Status.GetAccountSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	total := summary.EventCount
	coverage := "not analyzed"
	if total > 0 {
		coverage = fmt.Sprintf("%d%% analyzed", int(float64(summary.AnalyzedCount)/float64(total)*100))
	}
	volume := "waiting for volume"
	if summary.TotalVolumePerHour != nil {
		volume = format.Volume(*summary.TotalVolumePerHour) + " evt/hr"
	}

	return Snapshot{
		Title:       "Log events",
		Description: "Discovered log-event catalog and analysis coverage.",
		Primary:     Metric{Label: "events", Value: count(total)},
		Metrics: []Metric{
			{Label: "analyzed", Value: count(summary.AnalyzedCount), Tone: analyzedTone(summary, total)},
			{Label: "volume", Value: volume},
		},
		Rows: [][]string{
			{"Catalog", count(total), "events discovered"},
			{"Analysis", count(summary.AnalyzedCount), coverage},
			{"Runtime", volume, "current observed throughput"},
		},
		Loaded: true,
	}, nil
}

func fetchEdgeInstances(ctx context.Context, services graphql.ServiceSet) (Snapshot, error) {
	fleet, err := services.EdgeInstances.List(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	connected := fleet.ConnectedCount(time.Now(), edgeConnectedWindow)

	rows := make([][]string, 0, len(fleet.Instances))
	for _, inst := range fleet.Instances {
		state := "idle"
		if inst.LastSyncAt.After(time.Now().Add(-edgeConnectedWindow)) {
			state = "connected"
		}
		rows = append(rows, []string{inst.ServiceName, state, "last sync " + relativeTime(inst.LastSyncAt)})
	}
	if len(rows) == 0 {
		rows = [][]string{{"Runtime", "none", "no edge instances registered yet"}}
	}

	return Snapshot{
		Title:       "Edge instances",
		Description: "Edge runtimes syncing policies from this account.",
		Primary:     Metric{Label: "instances", Value: count(fleet.Total)},
		Metrics: []Metric{
			{Label: "connected", Value: count(connected), Tone: connectedTone(connected, fleet.Total)},
		},
		Rows:   rows,
		Loaded: true,
	}, nil
}

func snapshotKey(snapshot Snapshot) string {
	var b strings.Builder
	b.WriteString(snapshot.Title)
	b.WriteString(snapshot.Description)
	b.WriteString(snapshot.Primary.Label)
	b.WriteString(snapshot.Primary.Value)
	for _, metric := range snapshot.Metrics {
		b.WriteString(metric.Label)
		b.WriteString(metric.Value)
		b.WriteString(metric.Tone)
	}
	for _, row := range snapshot.Rows {
		b.WriteString(strings.Join(row, "\x00"))
	}
	return b.String()
}

func count(n int64) string {
	return fmt.Sprintf("%d", n)
}

// relativeTime renders a compact "Xm ago" style duration since t.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func pendingTone(n int64) string {
	if n > 0 {
		return "warning"
	}
	return "success"
}

func highTone(n int64) string {
	if n > 0 {
		return "danger"
	}
	return "success"
}

func connectedTone(connected, total int64) string {
	if total == 0 {
		return ""
	}
	if connected == 0 {
		return "danger"
	}
	if connected < total {
		return "warning"
	}
	return "success"
}

func analyzedTone(summary domain.AccountSummary, total int64) string {
	if total == 0 {
		return ""
	}
	if summary.AnalysisReady() {
		return "success"
	}
	return "warning"
}

func checkDomainSummary(catalog domain.CheckCatalog, d domain.CheckDomain) string {
	var openFindings, activeIssues int64
	for _, check := range catalog.Checks {
		if check.Domain != d {
			continue
		}
		openFindings += check.OpenFindingCount
		activeIssues += check.ActiveIssueCount
	}
	return fmt.Sprintf("%d checks, %d findings, %d issues", catalog.DomainCount(d), openFindings, activeIssues)
}
