// Package compliance renders the compliance indicator in the status bar.
package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
	"github.com/usetero/cli/internal/tea/keymap"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a compliance policy check.
type pollMsg struct{}

// dataMsg carries the result of an async data fetch.
type dataMsg struct {
	summary      domain.AccountSummary
	categories   []domain.ComplianceCategorySummary
	totalPending int64
	totalFixed   int64
	err          error
}

// detailMsg carries the result of an async detail fetch.
type detailMsg struct {
	cat      domain.ComplianceCategorySummary
	policies []domain.CompliancePolicy
	err      error
}

// Model renders compliance status: 4 categories (PII, Secrets, PHI, Payment Data).
type Model struct {
	theme styles.Theme
	scope log.Scope
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
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("compliance"),
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
		key := m.stateKey(msg.summary, msg.categories, msg.totalPending, msg.totalFixed)
		if key != m.lastState {
			m.summary = msg.summary
			m.categories = msg.categories
			m.totalPending = msg.totalPending
			m.totalFixed = msg.totalFixed
			m.hasData = msg.totalPending > 0 || msg.totalFixed > 0
			m.lastState = key

			// Clamp cursor if categories shrank.
			if m.cursor >= len(m.categories) && len(m.categories) > 0 {
				m.cursor = len(m.categories) - 1
			}
		}

	case detailMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.cat, msg.policies)
		}
	}

	return nil
}

// fetchData returns a Cmd that queries compliance data off the event loop.
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

		categories, err := db.CompliancePolicies().ListCategorySummaries(ctx)
		if err != nil {
			scope.Error("list category summaries", "err", err)
			categories = nil
		}

		totalPending, err := db.CompliancePolicies().CountTotal(ctx)
		if err != nil {
			scope.Error("count total", "err", err)
			totalPending = 0
		}

		totalFixed, err := db.CompliancePolicies().CountFixed(ctx)
		if err != nil {
			scope.Error("count fixed", "err", err)
			totalFixed = 0
		}

		return dataMsg{
			summary:      summary,
			categories:   categories,
			totalPending: totalPending,
			totalFixed:   totalFixed,
		}
	}
}

// fetchDetail returns a Cmd that queries category detail off the event loop.
func (m *Model) fetchDetail(cat domain.ComplianceCategorySummary) tea.Cmd {
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		policies, err := db.CompliancePolicies().ListPendingPoliciesByCategory(context.Background(), cat.Category, 25)
		if err != nil {
			scope.Error("list pending policies by category", "category", cat.Category, "err", err)
			return detailMsg{err: err}
		}
		return detailMsg{cat: cat, policies: policies}
	}
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(summary domain.AccountSummary, cats []domain.ComplianceCategorySummary, pending, fixed int64) string {
	key := fmt.Sprintf("%d:%d:%d:%d",
		summary.EventCount, summary.AnalyzedCount,
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
		return m.fetchDetail(cat)
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

	// Compliance counts are accurate immediately — either sensitive data
	// was observed or it wasn't. No need to gate on analysis threshold.
	var leaking, atRisk int64
	for _, c := range m.categories {
		leaking += c.LeakingCount
		atRisk += c.AtRiskCount
	}

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

// renderHeadline renders the compliance summary: leaking/at-risk counts and analysis progress.
// Dots always shown as a legend for row colors. Progress appended when analysis is incomplete.
func (m *Model) renderHeadline() string {
	s := m.summary
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
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

	// Analysis progress when not yet ready.
	if s.EventCount > 0 && !s.AnalysisReady() {
		pct := float64(s.AnalyzedCount) / float64(s.EventCount)
		bar := m.analysisBar()
		parts = append(parts, bar.ViewAs(pct)+" "+muted.Render(fmt.Sprintf("%d/%d analyzed", s.AnalyzedCount, s.EventCount)))
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
	tbl.Headers("Category", "Status", "Services", "Volume", "Fixed")
	tbl.SetWidth(width)

	err := lipgloss.NewStyle().Foreground(m.theme.Error).Background(m.theme.Bg)
	ok := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	for i, c := range m.categories {
		name := displayCategoryName(c.Category)
		if i == m.cursor {
			name = accent.Render("▶ " + name)
		} else {
			name = formatCategoryDot(m.theme, c) + " " + name
		}

		vol := "—"
		if c.VolumePerHour > 0 {
			vol = format.Volume(c.VolumePerHour) + "/hr"
		}

		fixedStr := "—"
		if c.FixedCount > 0 {
			fixedStr = ok.Render(format.Count(c.FixedCount))
		}

		tbl.Row(
			name,
			formatCategoryStatus(m.theme, c, err, muted),
			format.Count(int64(c.ServiceCount)),
			vol,
			fixedStr,
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

// analysisBar creates a small progress bar for inline use in the headline.
func (m *Model) analysisBar() progress.Model {
	bar := progress.New(
		progress.WithColors(m.theme.GradientStart, m.theme.GradientEnd),
		progress.WithWidth(10),
		progress.WithFillCharacters('█', '░'),
	)
	bar.ShowPercentage = false
	bar.EmptyColor = m.theme.TextMuted
	return bar
}

// formatCategoryDot returns a colored dot for a category row.
func formatCategoryDot(colors styles.Theme, c domain.ComplianceCategorySummary) string {
	if c.LeakingCount > 0 {
		return lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
	}
	if c.AtRiskCount > 0 {
		return lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
	}
	return lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
}

// formatCategoryStatus renders a combined status string like "4 leaking", "3 at risk",
// or "2 leaking · 1 at risk" when both exist.
func formatCategoryStatus(colors styles.Theme, c domain.ComplianceCategorySummary, errStyle, mutedStyle lipgloss.Style) string {
	sep := mutedStyle.Render(" · ")
	var parts []string
	if c.LeakingCount > 0 {
		parts = append(parts, errStyle.Render(fmt.Sprintf("%s leaking", format.Count(c.LeakingCount))))
	}
	if c.AtRiskCount > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		parts = append(parts, warnStyle.Render(fmt.Sprintf("%s at risk", format.Count(c.AtRiskCount))))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, sep)
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
