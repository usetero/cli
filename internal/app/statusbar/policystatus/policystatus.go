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
	"github.com/usetero/cli/internal/tea/components/status"
	"github.com/usetero/cli/internal/tea/components/table"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a policy status check.
type pollMsg struct{}

// Model renders the policy status: pending count, estimated savings, observed savings.
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	summary    domain.PolicySummary
	categories []domain.PolicyCategoryStatus
	hasData    bool
	lastState  string
}

// New creates a new policy status model.
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
		key += fmt.Sprintf("|%s:%d:%d:%d:%.0f:%.0f:%.2f:%s:%s",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount,
			c.EstimatedVolumePerHour, c.EstimatedBytesPerHour, c.EstimatedCostPerHour,
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
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var segments []string

	// Waste percentage with colored dot and pending count.
	if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.WarningBg).Background(colors.Bg).Render("●")
		waste := fmt.Sprintf("%d%% waste", wp)
		if s.PendingPolicyCount > 0 {
			waste += fmt.Sprintf(" (%d)", s.PendingPolicyCount)
		}
		segments = append(segments, dot+" "+muted.Render(waste))
	} else if s.PendingPolicyCount > 0 {
		// No waste % but pending policies.
		dot := lipgloss.NewStyle().Foreground(colors.WarningBg).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount)))
	}

	// Observed savings from approved policies.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.SuccessBg).Background(colors.Bg)
		segments = append(segments, savingStyle.Render("saving "+saving))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.SuccessBg).Background(colors.Bg).Render("●")
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
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	// Waste percentage with colored dot.
	if w := status.Waste(m.theme, wastePercent(s)); w != "" {
		parts = append(parts, w)
	}

	// Observed savings.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.SuccessBg).Background(colors.Bg)
		parts = append(parts, savingStyle.Render("saving "+saving))
	}

	// Pending count.
	if s.PendingPolicyCount > 0 {
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
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
	if len(m.categories) == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		return muted.Render("No policy data")
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Category", "Pending", "Risk", "Benefit", "Waste", "Cost", "Impact")
	tbl.SetWidth(width)

	for _, c := range m.categories {
		tbl.Row(
			c.Category,
			fmt.Sprintf("%d", c.PendingCount),
			m.renderRisk(c.RiskLevel),
			formatBenefitLabel(c.Benefit),
			m.formatWaste(c),
			formatCategoryCost(c),
			formatCategoryImpact(c),
		)
	}

	return tbl.View()
}

// categoryWastePercent computes a single category's waste as a percentage of total.
func (m *Model) categoryWastePercent(c domain.PolicyCategoryStatus) int {
	s := m.summary
	if s.TotalBytesPerHour > 0 && c.EstimatedBytesPerHour > 0 {
		return int(math.Round(c.EstimatedBytesPerHour / s.TotalBytesPerHour * 100))
	}
	if s.TotalVolumePerHour > 0 && c.EstimatedVolumePerHour > 0 {
		return int(math.Round(c.EstimatedVolumePerHour / s.TotalVolumePerHour * 100))
	}
	return 0
}

// formatWaste renders the waste % with a colored dot, matching the statusbar style.
func (m *Model) formatWaste(c domain.PolicyCategoryStatus) string {
	if w := status.WasteShort(m.theme, m.categoryWastePercent(c)); w != "" {
		return w
	}
	if c.PendingCount == 0 {
		return lipgloss.NewStyle().Foreground(m.theme.SuccessBg).Background(m.theme.Bg).Render("●")
	}
	return lipgloss.NewStyle().Background(m.theme.Bg).Render("—")
}

// renderRisk returns a color-coded risk label.
func (m *Model) renderRisk(level domain.RiskLevel) string {
	colors := m.theme
	s := level.String()
	switch level {
	case domain.RiskLevelHigh:
		return lipgloss.NewStyle().Foreground(colors.ErrorBg).Background(colors.Bg).Render(s)
	case domain.RiskLevelMedium:
		return lipgloss.NewStyle().Foreground(colors.WarningBg).Background(colors.Bg).Render(s)
	case domain.RiskLevelLow:
		return lipgloss.NewStyle().Foreground(colors.SuccessBg).Background(colors.Bg).Render(s)
	default:
		return lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg).Render(s)
	}
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

// formatCategoryCost returns estimated yearly cost reduction for a category.
func formatCategoryCost(c domain.PolicyCategoryStatus) string {
	if c.EstimatedCostPerHour > 0 {
		yearly := c.EstimatedCostPerHour * 8760
		if yearly >= 1 {
			return "~" + formatCost(yearly) + "/yr"
		}
	}
	return "—"
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

// formatBenefitLabel returns human-readable benefit labels.
// The input may contain JSON arrays like '["volume_reduction","signal_quality"]'
// concatenated by GROUP_CONCAT.
func formatBenefitLabel(benefits string) string {
	labelFor := map[string]string{
		"volume_reduction": "cost",
		"bytes_reduction":  "bytes",
		"signal_quality":   "signal",
		"compliance":       "compliance",
		"resilience":       "resilience",
	}

	// Strip JSON array noise: brackets, quotes, whitespace
	cleaned := strings.NewReplacer("[", "", "]", "", "\"", "").Replace(benefits)

	seen := make(map[string]bool)
	var labels []string
	for _, part := range strings.Split(cleaned, ",") {
		key := strings.TrimSpace(part)
		label, ok := labelFor[key]
		if !ok {
			continue
		}
		if !seen[label] {
			seen[label] = true
			labels = append(labels, label)
		}
	}
	if len(labels) == 0 {
		return ""
	}
	return strings.Join(labels, ", ")
}

// formatEstimated returns the estimated impact of pending policies.
func formatEstimated(s domain.PolicySummary) string {
	if s.EstimatedCostPerHour != nil {
		yearly := *s.EstimatedCostPerHour * 8760
		if yearly >= 1 {
			return formatCost(yearly) + "/yr"
		}
	}
	return ""
}

// formatObservedSaving returns the observed savings from approved policies.
func formatObservedSaving(s domain.PolicySummary) string {
	if s.ObservedCostBefore != nil && s.ObservedCostAfter != nil {
		diff := (*s.ObservedCostBefore - *s.ObservedCostAfter) * 8760
		if diff >= 1 {
			return formatCost(diff) + "/yr"
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
