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
	theme styles.Theme
	db    sqlite.DB

	summary    domain.AccountSummary
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

		summary, err := m.db.DatadogAccountStatuses().GetSummary(ctx)
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
			m.hasData = summary.PolicyCount > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%v:%v:%v:%.0f:%.0f:%v:%v:%.0f:%.0f:%v:%.0f:%.0f:%d",
		s.ReadyForUse, s.PendingPolicyCount,
		s.PolicyCount, s.ApprovedPolicyCount,
		ptrVal(s.EstimatedCostPerHour),
		ptrVal(s.EstimatedCostPerHourBytes),
		ptrVal(s.EstimatedCostPerHourVolume),
		s.EstimatedVolumePerHour, s.EstimatedBytesPerHour,
		ptrVal(s.ObservedCostBefore), ptrVal(s.ObservedCostAfter),
		s.ObservedVolumeBefore, s.ObservedVolumeAfter,
		ptrVal(s.TotalCostPerHour), s.TotalVolumePerHour,
		s.TotalBytesPerHour, len(cats))

	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%.0f:%.0f:%.2f",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount,
			c.EstimatedVolumePerHour, c.EstimatedBytesPerHour, c.EstimatedCostPerHour)
	}

	return key
}

// HasData returns true when policy data has been loaded.
func (m *Model) HasData() bool {
	return m.hasData
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

	// Observed savings from approved policies.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		segments = append(segments, savingStyle.Render("saving "+saving))
	}

	// Waste percentage with pending count.
	if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		waste := fmt.Sprintf("%d%% waste", wp)
		if s.PendingPolicyCount > 0 {
			waste += fmt.Sprintf(" (%d)", s.PendingPolicyCount)
		}
		segments = append(segments, dot+" "+muted.Render(waste))
	} else if s.PendingPolicyCount > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount)))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("healthy")
	}

	return strings.Join(segments, sep)
}

// wastePercent computes the estimated waste as a percentage of total bytes.
func wastePercent(s domain.AccountSummary) int {
	if s.TotalBytesPerHour > 0 && s.EstimatedBytesPerHour > 0 {
		return int(math.Round(s.EstimatedBytesPerHour / s.TotalBytesPerHour * 100))
	}
	return 0
}

// ExpandedView renders the detailed policy status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		return ""
	}

	// Height budget: headline (1) + gap (1) + table header+border (2) = 4 lines overhead.
	maxRows := height - 4
	if maxRows < 1 {
		maxRows = 1
	}

	var lines []string
	lines = append(lines, m.renderWasteHeadline())
	lines = append(lines, "")
	lines = append(lines, m.renderWasteTable(width, maxRows))
	return strings.Join(lines, "\n")
}

// renderWasteHeadline renders the waste summary: waste % and pending count.
func (m *Model) renderWasteHeadline() string {
	s := m.summary
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	// Observed savings from approved policies.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		parts = append(parts, savingStyle.Render("saving "+saving))
	}

	// Waste % + pending count.
	if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		waste := dot + " " + text.Render(fmt.Sprintf("%d%% waste", wp))
		if s.PendingPolicyCount > 0 {
			waste += sep + text.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount))
		}
		parts = append(parts, waste)
	} else if s.PendingPolicyCount > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount)))
	}

	if len(parts) == 0 {
		return muted.Render("All policies reviewed")
	}

	return strings.Join(parts, sep)
}

// isWasteCategory returns true for categories with cost or data impact.
func isWasteCategory(c domain.PolicyCategoryStatus) bool {
	return c.EstimatedCostPerHour > 0 || c.EstimatedVolumePerHour > 0 || c.EstimatedBytesPerHour > 0
}

// renderWasteTable renders the waste category breakdown.
func (m *Model) renderWasteTable(width, maxRows int) string {
	// Filter to waste-related categories only.
	var waste []domain.PolicyCategoryStatus
	for _, c := range m.categories {
		if isWasteCategory(c) {
			waste = append(waste, c)
		}
	}

	if len(waste) == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		return muted.Render("No waste data")
	}

	// Reserve a row for "+N more" if we need to clip.
	clipped := 0
	visible := waste
	if len(waste) > maxRows {
		visible = waste[:maxRows-1]
		clipped = len(waste) - len(visible)
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Category", "Pending", "Volume", "Bytes", "Savings")
	tbl.SetWidth(width)

	for _, c := range visible {
		tbl.Row(
			c.Category,
			fmt.Sprintf("%d", c.PendingCount),
			formatCategoryVolume(c),
			formatCategoryBytes(c),
			formatCategoryCost(c),
		)
	}

	result := tbl.View()
	if clipped > 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		result += "\n" + muted.Render(fmt.Sprintf("+%d more", clipped))
	}
	return result
}

// formatCategoryCost returns estimated yearly cost for a category.
func formatCategoryCost(c domain.PolicyCategoryStatus) string {
	if c.EstimatedCostPerHour > 0 {
		yearly := c.EstimatedCostPerHour * 8760
		if yearly >= 1 {
			return "~" + formatCost(yearly) + "/yr"
		}
	}
	return "—"
}

// formatCategoryVolume returns the volume impact for a category.
func formatCategoryVolume(c domain.PolicyCategoryStatus) string {
	if c.EstimatedVolumePerHour > 0 {
		return formatVolume(c.EstimatedVolumePerHour) + "/hr"
	}
	return "—"
}

// formatCategoryBytes returns the bytes impact for a category.
func formatCategoryBytes(c domain.PolicyCategoryStatus) string {
	if c.EstimatedBytesPerHour > 0 {
		return formatBytes(c.EstimatedBytesPerHour) + "/hr"
	}
	return "—"
}

// formatObservedSaving returns the observed savings from approved policies.
func formatObservedSaving(s domain.AccountSummary) string {
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
