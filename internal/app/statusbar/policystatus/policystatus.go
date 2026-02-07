// Package policystatus renders the policy work indicator in the status bar.
package policystatus

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
	"github.com/usetero/cli/internal/tea/components/table"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a policy status check.
type pollMsg struct{}

// Model renders the policy status: pending count, estimated savings, observed savings.
type Model struct {
	theme *styles.Theme
	db    sqlite.DB

	summary    domain.PolicySummary
	categories []domain.PolicyCategoryStatus
	hasData    bool
	lastState  string
}

// New creates a new policy status model.
func New(theme *styles.Theme) *Model {
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

		summary, err := m.db.DatadogAccountStatuses().GetPolicySummary(ctx)
		if err != nil {
			return m.poll()
		}

		categories, err := m.db.LogEventPolicies().ListCategoryStatuses(ctx)
		if err != nil {
			categories = nil
		}

		key := m.stateKey(summary, categories)
		if key != m.lastState {
			m.summary = summary
			m.categories = categories
			m.hasData = summary.ReadyForUse
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.PolicySummary, cats []domain.PolicyCategoryStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%v:%.0f:%.0f:%v:%v:%.0f:%.0f:%v:%.0f:%.0f:%d",
		s.ReadyForUse, s.PendingPolicyCount,
		s.PolicyCount, s.ApprovedPolicyCount,
		ptrVal(s.EstimatedCostPerHour),
		s.EstimatedVolumePerHour, s.EstimatedBytesPerHour,
		ptrVal(s.ObservedCostBefore), ptrVal(s.ObservedCostAfter),
		s.ObservedVolumeBefore, s.ObservedVolumeAfter,
		ptrVal(s.TotalCostPerHour), s.TotalVolumePerHour,
		s.TotalBytesPerHour, len(cats))

	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%.0f:%.0f:%s:%s",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount,
			c.EstimatedVolumePerHour, c.EstimatedBytesPerHour,
			c.RiskLevel, c.Benefit)
	}

	return key
}

// CompactView renders the policy status for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	s := m.summary
	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	sep := muted.Render(" · ")

	var segments []string

	// Waste percentage with colored dot and pending count.
	if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning.Fg).Render("●")
		waste := fmt.Sprintf("%d%% waste", wp)
		if s.PendingPolicyCount > 0 {
			waste += fmt.Sprintf(" (%d)", s.PendingPolicyCount)
		}
		segments = append(segments, dot+" "+muted.Render(waste))
	} else if s.PendingPolicyCount > 0 {
		// No waste % but pending policies.
		dot := lipgloss.NewStyle().Foreground(colors.Warning.Fg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount)))
	}

	// Observed savings from approved policies.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success.Fg)
		segments = append(segments, savingStyle.Render("saving "+saving))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success.Fg).Render("●")
		return dot + " " + muted.Render("healthy")
	}

	return strings.Join(segments, sep)
}

// ExpandedView renders the detailed policy status for the drawer.
func (m *Model) ExpandedView(width int) string {
	if !m.hasData {
		return ""
	}

	var lines []string
	lines = append(lines, m.renderHeadline())
	lines = append(lines, "")
	lines = append(lines, m.renderCategoryTable(width))
	return strings.Join(lines, "\n")
}

// renderHeadline renders the top summary lines.
func (m *Model) renderHeadline() string {
	s := m.summary
	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	sep := muted.Render(" · ")

	var parts []string

	text := lipgloss.NewStyle().Foreground(colors.Page.Text)

	// Waste percentage with colored dot.
	if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning.Fg).Render("●")
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d%% waste", wp)))
	}

	// Observed savings.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success.Fg)
		parts = append(parts, savingStyle.Render("saving "+saving))
	}

	// Pending count.
	if s.PendingPolicyCount > 0 {
		est := formatEstimated(s)
		if est != "" {
			parts = append(parts, text.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount))+sep+muted.Render("~"+est))
		} else {
			parts = append(parts, text.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount)))
		}
	}

	if len(parts) == 0 {
		return muted.Render("All policies reviewed")
	}

	return strings.Join(parts, sep)
}

// renderCategoryTable renders the per-category breakdown.
func (m *Model) renderCategoryTable(width int) string {
	// Only show categories with pending policies.
	var pending []domain.PolicyCategoryStatus
	for _, c := range m.categories {
		if c.PendingCount > 0 {
			pending = append(pending, c)
		}
	}

	if len(pending) == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.Colors.Page.TextMuted)
		return muted.Render("No pending policies")
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Category", "Pending", "Impact", "Benefit", "Risk")
	tbl.SetWidth(width)

	for _, c := range pending {
		tbl.Row(
			c.Category,
			fmt.Sprintf("%d", c.PendingCount),
			formatCategoryImpact(c),
			formatBenefitLabel(c.Benefit),
			c.RiskLevel,
		)
	}

	return tbl.View()
}

// wastePercent computes the estimated waste as a percentage.
// Prefers bytes (where most waste lives), then cost, then volume.
func wastePercent(s domain.PolicySummary) int {
	if s.TotalBytesPerHour > 0 && s.EstimatedBytesPerHour > 0 {
		return int(math.Round(s.EstimatedBytesPerHour / s.TotalBytesPerHour * 100))
	}
	if s.EstimatedCostPerHour != nil && s.TotalCostPerHour != nil && *s.TotalCostPerHour > 0 {
		return int(math.Round(*s.EstimatedCostPerHour / *s.TotalCostPerHour * 100))
	}
	if s.TotalVolumePerHour > 0 && s.EstimatedVolumePerHour > 0 {
		return int(math.Round(s.EstimatedVolumePerHour / s.TotalVolumePerHour * 100))
	}
	return 0
}

// formatCategoryImpact returns the best available impact string for a category.
// Cascades: volume → bytes → dash.
func formatCategoryImpact(c domain.PolicyCategoryStatus) string {
	if c.EstimatedVolumePerHour > 0 {
		return "~" + formatVolume(c.EstimatedVolumePerHour) + " evt/hr"
	}
	if c.EstimatedBytesPerHour > 0 {
		return "~" + formatBytes(c.EstimatedBytesPerHour) + "/hr"
	}
	return "—"
}

// formatBenefitLabel returns a human-readable benefit label.
func formatBenefitLabel(benefit string) string {
	switch {
	case strings.Contains(benefit, "volume_reduction"):
		return "cost"
	case strings.Contains(benefit, "bytes_reduction"):
		return "bytes"
	case strings.Contains(benefit, "signal_quality"):
		return "signal quality"
	case strings.Contains(benefit, "compliance"):
		return "compliance"
	case strings.Contains(benefit, "resilience"):
		return "resilience"
	default:
		return benefit
	}
}

// formatEstimated returns the estimated impact of pending policies.
func formatEstimated(s domain.PolicySummary) string {
	if s.EstimatedCostPerHour != nil {
		monthly := *s.EstimatedCostPerHour * 730
		if monthly >= 1 {
			return formatCost(monthly) + "/mo"
		}
	}
	return ""
}

// formatObservedSaving returns the observed savings from approved policies.
func formatObservedSaving(s domain.PolicySummary) string {
	if s.ObservedCostBefore != nil && s.ObservedCostAfter != nil {
		diff := (*s.ObservedCostBefore - *s.ObservedCostAfter) * 730
		if diff >= 1 {
			return formatCost(diff) + "/mo"
		}
		return ""
	}
	diff := s.ObservedVolumeBefore - s.ObservedVolumeAfter
	if diff > 0 {
		return formatVolume(diff) + " evt/hr"
	}
	return ""
}

// formatCost formats a dollar amount: $142, $9.4k, $1.2M.
func formatCost(dollars float64) string {
	abs := math.Abs(dollars)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("$%.1fM", dollars/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("$%.1fk", dollars/1_000)
	default:
		return fmt.Sprintf("$%.0f", dollars)
	}
}

// formatVolume formats events/hr: 892, 45.3k, 2.1M.
func formatVolume(eventsPerHour float64) string {
	abs := math.Abs(eventsPerHour)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", eventsPerHour/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", eventsPerHour/1_000)
	default:
		return fmt.Sprintf("%.0f", eventsPerHour)
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

func ptrVal(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
