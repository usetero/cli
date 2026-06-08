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
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

const (
	pollInterval = 2 * time.Second
	dbTimeout    = 2 * time.Second
)

type fetchFunc func(context.Context, sqlite.DB) (Snapshot, error)

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
	db    sqlite.DB

	source string
	fetch  fetchFunc

	snapshot  Snapshot
	hasData   bool
	fetching  bool
	lastState string
}

// NewPolicies creates the policies surface.
func NewPolicies(theme styles.Theme, scope log.Scope) *Model {
	return newModel(theme, scope, "policies", fetchPolicies)
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

// SetDB sets the database and starts polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	m.db = db
	return m.poll()
}

// Init starts polling when the runtime database is available.
func (m *Model) Init() tea.Cmd {
	if m.db == nil {
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
		m.db != nil,
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
	db := m.db
	fetch := m.fetch
	scope := m.scope
	return tabpoll.Fetch(dbTimeout, func(ctx context.Context) (Snapshot, error) {
		snapshot, err := fetch(ctx, db)
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

func fetchPolicies(ctx context.Context, db sqlite.DB) (Snapshot, error) {
	summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	total, err := db.LogEventPolicies().Count(ctx)
	if err != nil {
		total = summary.PendingPolicyCount + summary.ApprovedPolicyCount + summary.DismissedPolicyCount
	}

	return Snapshot{
		Title:       "Policies",
		Description: "Reviewable policy state synced from the control plane.",
		Primary:     Metric{Label: "total", Value: count(total)},
		Metrics: []Metric{
			{Label: "pending", Value: count(summary.PendingPolicyCount), Tone: pendingTone(summary.PendingPolicyCount)},
			{Label: "approved", Value: count(summary.ApprovedPolicyCount), Tone: "success"},
			{Label: "dismissed", Value: count(summary.DismissedPolicyCount)},
		},
		Rows: [][]string{
			{"Awaiting review", count(summary.PendingPolicyCount), highSignal(summary)},
			{"Approved", count(summary.ApprovedPolicyCount), "running policy decisions"},
			{"Dismissed", count(summary.DismissedPolicyCount), "reviewed and set aside"},
		},
		Loaded: true,
	}, nil
}

func fetchIssues(ctx context.Context, db sqlite.DB) (Snapshot, error) {
	summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	high := summary.PolicyPendingCriticalCount + summary.PolicyPendingHighCount
	return Snapshot{
		Title:       "Issues",
		Description: "Policy-remediable issues currently awaiting operator judgment.",
		Primary:     Metric{Label: "open", Value: count(summary.PendingPolicyCount), Tone: pendingTone(summary.PendingPolicyCount)},
		Metrics: []Metric{
			{Label: "high", Value: count(high), Tone: highTone(high)},
			{Label: "medium", Value: count(summary.PolicyPendingMediumCount), Tone: pendingTone(summary.PolicyPendingMediumCount)},
			{Label: "low", Value: count(summary.PolicyPendingLowCount)},
		},
		Rows: [][]string{
			{"Critical", count(summary.PolicyPendingCriticalCount), "needs immediate review"},
			{"High", count(summary.PolicyPendingHighCount), "high-priority review queue"},
			{"Medium", count(summary.PolicyPendingMediumCount), "normal review queue"},
			{"Low", count(summary.PolicyPendingLowCount), "background cleanup"},
		},
		Loaded: true,
	}, nil
}

func fetchChecks(ctx context.Context, db sqlite.DB) (Snapshot, error) {
	statuses := db.LogEventPolicyCategoryStatuses()
	waste, err := statuses.ListWasteCategoryStatuses(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	quality, err := statuses.ListQualityCategoryStatuses(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	compliance, err := statuses.ListComplianceCategoryStatuses(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Title:       "Checks",
		Description: "Control-plane check categories represented in the local projection.",
		Primary:     Metric{Label: "categories", Value: count(int64(len(waste) + len(quality) + len(compliance)))},
		Metrics: []Metric{
			{Label: "cost", Value: count(int64(len(waste)))},
			{Label: "quality", Value: count(int64(len(quality)))},
			{Label: "compliance", Value: count(int64(len(compliance)))},
		},
		Rows: [][]string{
			{"Cost", categorySummary(waste), pendingCategorySignal(waste)},
			{"Data quality", categorySummary(quality), pendingCategorySignal(quality)},
			{"Compliance", categorySummary(compliance), pendingCategorySignal(compliance)},
		},
		Loaded: true,
	}, nil
}

func fetchLogEvents(ctx context.Context, db sqlite.DB) (Snapshot, error) {
	summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	total, err := db.LogEvents().Count(ctx)
	if err != nil {
		total = summary.EventCount
	}
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

func fetchEdgeInstances(context.Context, sqlite.DB) (Snapshot, error) {
	return Snapshot{
		Title:       "Edge instances",
		Description: "Edge runtime projection is not synced into this CLI yet.",
		Primary:     Metric{Label: "instances", Value: "0"},
		Metrics: []Metric{
			{Label: "connected", Value: "0"},
			{Label: "projection", Value: "pending"},
		},
		Rows: [][]string{
			{"Runtime", "pending", "waiting for edge instance sync"},
			{"Control plane", "available in webapp", "CLI surface reserved"},
		},
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

func highSignal(summary domain.AccountSummary) string {
	high := summary.PolicyPendingCriticalCount + summary.PolicyPendingHighCount
	if high > 0 {
		return fmt.Sprintf("%d high priority", high)
	}
	return "no high-priority issues"
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

func categorySummary(categories []domain.PolicyCategoryStatus) string {
	var pending int64
	for _, category := range categories {
		pending += category.PendingCount
	}
	return fmt.Sprintf("%d categories, %d pending", len(categories), pending)
}

func pendingCategorySignal(categories []domain.PolicyCategoryStatus) string {
	var high int64
	var estimatedCost float64
	for _, category := range categories {
		high += category.PolicyPendingCriticalCount + category.PolicyPendingHighCount
		if category.EstimatedCostPerHour != nil {
			estimatedCost += *category.EstimatedCostPerHour
		}
	}
	if high > 0 {
		return fmt.Sprintf("%d high-priority policies", high)
	}
	if estimatedCost > 0 {
		return format.YearlyCost(estimatedCost)
	}
	return "no high-priority policies"
}
