// Package compliance renders the compliance indicator in the status bar.
package compliance

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
	"github.com/usetero/cli/internal/tea/keymap"
)

const (
	pollInterval = 2 * time.Second

	// discoveryDoneThreshold is the classification coverage percentage above
	// which we consider discovery complete. Volume ratios are never exactly
	// 100% due to throughput fluctuations.
	discoveryDoneThreshold = 95
)

// pollMsg triggers a compliance policy check.
type pollMsg struct{}

// Model renders compliance status: 4 categories (PII, Secrets, PHI, Payment Data).
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	summary      domain.AccountSummary
	categories   []domain.ComplianceCategorySummary
	totalPending int64
	totalFixed   int64
	hasData      bool
	lastState    string

	// Drawer navigation
	cursor int     // selected row in category list
	detail *detail // non-nil when viewing a single category's policies
}

// New creates a new compliance status model.
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

		categories, err := m.db.CompliancePolicies().ListCategorySummaries(ctx)
		if err != nil {
			categories = nil
		}

		totalPending, err := m.db.CompliancePolicies().CountTotal(ctx)
		if err != nil {
			totalPending = 0
		}

		totalFixed, err := m.db.CompliancePolicies().CountFixed(ctx)
		if err != nil {
			totalFixed = 0
		}

		key := m.stateKey(summary, categories, totalPending, totalFixed)
		if key != m.lastState {
			m.summary = summary
			m.categories = categories
			m.totalPending = totalPending
			m.totalFixed = totalFixed
			m.hasData = totalPending > 0 || totalFixed > 0
			m.lastState = key

			// Clamp cursor if categories shrank.
			if m.cursor >= len(m.categories) && len(m.categories) > 0 {
				m.cursor = len(m.categories) - 1
			}
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(summary domain.AccountSummary, cats []domain.ComplianceCategorySummary, pending, fixed int64) string {
	key := fmt.Sprintf("%v:%v:%d:%d",
		summary.TotalServiceVolumePerHour, summary.TotalVolumePerHour,
		pending, fixed)

	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%v:%d",
			c.Category, c.LeakingCount, c.AtRiskCount, c.FixedCount,
			c.VolumePerHour, c.ServiceCount)
	}

	return key
}

// HasData returns true when compliance policy data has been loaded.
func (m *Model) HasData() bool {
	return m.hasData
}

// HandleKeyPress handles keyboard navigation in the expanded drawer view.
func (m *Model) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if !m.hasData || len(m.categories) == 0 {
		return nil
	}

	// Detail mode: backspace returns to category list (esc handled by statusbar).
	if m.detail != nil {
		if key.Matches(msg, keymap.DrawerBack) {
			m.detail = nil
		}
		return nil
	}

	// Category list mode.
	switch {
	case key.Matches(msg, keymap.DrawerUp):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keymap.DrawerDown):
		if m.cursor < len(m.categories)-1 {
			m.cursor++
		}
	case key.Matches(msg, keymap.DrawerSelect):
		cat := m.categories[m.cursor]
		if cat.LeakingCount+cat.AtRiskCount == 0 {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		policies, err := m.db.CompliancePolicies().ListPendingPoliciesByCategory(ctx, cat.Category, 25)
		if err != nil {
			return nil
		}
		m.detail = newDetail(m.theme, cat, policies)
	}
	return nil
}

// InDetail returns true when the detail sub-view is active.
func (m *Model) InDetail() bool {
	return m.detail != nil
}

// CloseDetail exits the detail sub-view, returning to the category list.
func (m *Model) CloseDetail() {
	m.detail = nil
}

// CompactView renders the compliance indicator for the collapsed statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var segments []string

	// Count leaking vs at-risk across all categories.
	var leaking, atRisk int64
	for _, c := range m.categories {
		leaking += c.LeakingCount
		atRisk += c.AtRiskCount
	}

	// Red dot if any category has leaking data, warning otherwise.
	if leaking > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d leaking", leaking)))
	} else if atRisk > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d at risk", atRisk)))
	}

	// Fixed count.
	if m.totalFixed > 0 {
		ok := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		segments = append(segments, ok.Render(fmt.Sprintf("%d fixed", m.totalFixed)))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("compliant")
	}

	return strings.Join(segments, sep)
}

// ExpandedView renders the detailed compliance status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		if m.db == nil {
			return muted.Render("Waiting for sync to start...")
		}
		dot := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg).Render("●")
		return dot + " " + muted.Render("No compliance issues detected.")
	}

	// Detail sub-view for a single category.
	if m.detail != nil {
		return m.detail.View(width)
	}

	var lines []string
	lines = append(lines, m.renderHeadline())
	lines = append(lines, "")

	if tbl := m.renderCategoryTable(width); tbl != "" {
		lines = append(lines, tbl)
	}

	return strings.Join(lines, "\n")
}

// renderHeadline renders the compliance summary: leaking/at-risk counts and discovery progress.
func (m *Model) renderHeadline() string {
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	// Count leaking vs at-risk across all categories.
	var leaking, atRisk int64
	for _, c := range m.categories {
		leaking += c.LeakingCount
		atRisk += c.AtRiskCount
	}

	if leaking > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d leaking", leaking)))
	}
	if atRisk > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		parts = append(parts, dot+" "+muted.Render(fmt.Sprintf("%d at risk", atRisk)))
	}

	// Service count across all categories.
	serviceCount := m.uniqueServiceCount()
	if serviceCount > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("across %d services", serviceCount)))
	}

	// Discovery progress when classification coverage is below threshold.
	if pct := summaryDiscoveryPercent(m.summary); pct < discoveryDoneThreshold {
		bar := m.discoveryBar()
		parts = append(parts, bar.ViewAs(float64(pct)/100)+" "+muted.Render(fmt.Sprintf("%d%%", pct)))
	}

	if m.totalFixed > 0 {
		ok := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		parts = append(parts, ok.Render(fmt.Sprintf("%d fixed", m.totalFixed)))
	}

	if len(parts) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("No compliance issues detected.")
	}

	return strings.Join(parts, sep)
}

// renderCategoryTable renders all compliance categories in a single table with cursor highlighting.
func (m *Model) renderCategoryTable(width int) string {
	if len(m.categories) == 0 {
		return ""
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(30))
	tbl.Headers("Category", "Leaking", "At Risk", "Fixed", "Services")
	tbl.SetWidth(width)

	warn := lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg)
	err := lipgloss.NewStyle().Foreground(m.theme.Error).Background(m.theme.Bg)
	ok := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)

	for i, c := range m.categories {
		dot := ok.Render("●")
		if c.LeakingCount > 0 {
			dot = err.Render("●")
		} else if c.AtRiskCount > 0 {
			dot = warn.Render("●")
		}

		name := displayCategoryName(c.Category)
		if i == m.cursor {
			name = accent.Render("▶ " + name)
		} else {
			name = dot + " " + name
		}

		leakingStr := "—"
		if c.LeakingCount > 0 {
			leakingStr = err.Render(format.Count(c.LeakingCount))
		}

		atRiskStr := "—"
		if c.AtRiskCount > 0 {
			atRiskStr = format.Count(c.AtRiskCount)
		}

		fixedStr := "—"
		if c.FixedCount > 0 {
			fixedStr = ok.Render(format.Count(c.FixedCount))
		}

		tbl.Row(
			name,
			leakingStr,
			atRiskStr,
			fixedStr,
			format.Count(int64(c.ServiceCount)),
		)
	}

	return tbl.View()
}

// uniqueServiceCount returns the total number of unique services across all categories.
func (m *Model) uniqueServiceCount() int {
	seen := make(map[string]struct{})
	for _, c := range m.categories {
		for _, s := range c.UniqueServices {
			seen[s] = struct{}{}
		}
	}
	return len(seen)
}

// discoveryBar creates a small progress bar for inline use in the headline.
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

// displayCategoryName returns a human-readable name for a compliance category.
func displayCategoryName(category string) string {
	switch category {
	case domain.CategoryPIILeakage:
		return "PII"
	case domain.CategorySecretsLeakage:
		return "Secrets"
	case domain.CategoryPHILeakage:
		return "PHI"
	case domain.CategoryPaymentDataLeakage:
		return "Payment Data"
	default:
		return format.TitleCase(category)
	}
}
