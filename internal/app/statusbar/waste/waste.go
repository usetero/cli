// Package waste renders the waste indicator in the status bar.
package waste

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
	"github.com/usetero/cli/internal/app/statusbar/viewkit"
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
	pollSource   = "waste"
)

type fetchedData struct {
	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
}

// detailMsg carries the result of an async detail fetch.
type detailMsg struct {
	cat      domain.PolicyCategoryStatus
	policies []domain.WastePolicy
	err      error
}

// Model renders the policy status: pending count, estimated savings, observed savings.
type Model struct {
	theme styles.Theme
	scope log.Scope
	db    sqlite.DB

	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
	hasData    bool
	lastState  string
	fetching   bool

	// Drawer navigation
	cursor int     // selected row in category list
	detail *detail // non-nil when viewing a single category's policies
}

// New creates a new policy status model.
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("waste"),
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
	return tabpoll.Tick(pollSource, pollInterval)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if cmd, handled := tabpoll.UpdatePollCycle(
		msg,
		pollSource,
		m.db != nil,
		&m.fetching,
		m.fetchData(),
		m.poll(),
		func(data fetchedData) {
			key := m.stateKey(data.summary, data.categories)
			tabpoll.ApplyIfChanged(&m.lastState, key, &m.cursor, len(data.categories), func() {
				m.summary = data.summary
				m.categories = data.categories
				m.hasData = len(data.categories) > 0 || data.summary.PendingPolicyCount+data.summary.ApprovedPolicyCount+data.summary.DismissedPolicyCount > 0
			})
		},
	); handled {
		return cmd
	}

	switch msg := msg.(type) {
	case detailMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.cat, msg.policies)
		}

	default:
		_ = msg
	}

	return nil
}

// fetchData returns a Cmd that queries waste data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	db := m.db
	scope := m.scope
	return tabpoll.Fetch(dbTimeout, func(ctx context.Context) (fetchedData, error) {
		summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			scope.Error("get summary", "err", err)
			return fetchedData{}, err
		}

		categories, err := db.LogEventPolicyCategoryStatuses().ListWasteCategoryStatuses(ctx)
		if err != nil {
			scope.Error("list waste category statuses", "err", err)
			categories = nil
		}

		return fetchedData{summary: summary, categories: categories}, nil
	})
}

// fetchDetail returns a Cmd that queries category detail off the event loop.
func (m *Model) fetchDetail(cat domain.PolicyCategoryStatus) tea.Cmd {
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		ctx, cancel := sqlite.WithTimeout(context.Background(), dbTimeout)
		defer cancel()
		policies, err := db.LogEventPolicyStatuses().ListTopPendingPoliciesByCategory(ctx, cat.Category, 25)
		if err != nil {
			scope.Error("list top pending policies", "category", cat.Category, "err", err)
			return detailMsg{err: err}
		}
		return detailMsg{cat: cat, policies: policies}
	}
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(s domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%d",
		s.ReadyForUse, s.EventCount, s.AnalyzedCount, s.PendingPolicyCount,
		s.ApprovedPolicyCount, s.DismissedPolicyCount,
		s.EstimatedCostPerHour, s.EstimatedCostPerHourBytes, s.EstimatedCostPerHourVolume,
		s.EstimatedVolumePerHour, s.EstimatedBytesPerHour,
		s.ObservedCostBefore, s.ObservedCostAfter,
		s.ObservedVolumeBefore, s.ObservedVolumeAfter,
		s.TotalCostPerHour, s.TotalVolumePerHour,
		s.TotalBytesPerHour, len(cats))

	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%v:%v:%v:%d:%d",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount,
			c.EstimatedVolumePerHour, c.EstimatedBytesPerHour, c.EstimatedCostPerHour,
			c.EventsWithVolumes, c.TotalEvents)
	}

	return key
}

// HasData returns true when policy data has been loaded.
func (m *Model) HasData() bool {
	return m.hasData
}

// HandleKeyPress handles keyboard navigation in the expanded drawer view.
func (m *Model) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	return m.navController().HandleKeyPress(msg)
}

// InDetail returns true when the detail sub-view is active.
func (m *Model) InDetail() bool {
	return m.detail != nil
}

// CloseDetail exits the detail sub-view, returning to the category list.
func (m *Model) CloseDetail() {
	m.detail = nil
}

func (m *Model) navController() listdetail.Controller {
	return listdetail.Controller{
		HasList: func() bool { return m.hasData && len(m.categories) > 0 },
		IsDetail: func() bool {
			return m.detail != nil
		},
		CloseDetail: func() { m.detail = nil },
		GetListCursor: func() int {
			return m.cursor
		},
		SetListCursor: func(v int) { m.cursor = v },
		ListLen:       func() int { return len(m.categories) },
		OnListSelect: func(index int) tea.Cmd {
			cat := m.categories[index]
			if cat.PendingCount == 0 {
				return nil
			}
			return m.fetchDetail(cat)
		},
		GetDetailCursor: func() int {
			if m.detail == nil {
				return 0
			}
			return m.detail.cursor
		},
		SetDetailCursor: func(v int) {
			if m.detail != nil {
				m.detail.cursor = v
			}
		},
		DetailLen: func() int {
			if m.detail == nil {
				return 0
			}
			return len(m.detail.policies)
		},
		OnDetailSelect: func() tea.Cmd {
			if m.detail == nil {
				return nil
			}
			return m.detail.Prompt()
		},
	}
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

	// Observed savings from approved policies (always shown — these are measured).
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		segments = append(segments, savingStyle.Render("saving "+saving))
	}

	if s.AnalysisReady() {
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
			segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d policies", s.PendingPolicyCount)))
		}
	} else if s.EventCount > 0 {
		// Analysis still in progress — show progress instead of waste %.
		segments = append(segments, muted.Render(fmt.Sprintf("%d/%d analyzed", s.AnalyzedCount, s.EventCount)))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("healthy")
	}

	return strings.Join(segments, sep)
}

// wastePercent computes the estimated waste as a percentage of total bytes.
func wastePercent(s domain.AccountSummary) int {
	if s.TotalBytesPerHour != nil && *s.TotalBytesPerHour > 0 &&
		s.EstimatedBytesPerHour != nil && *s.EstimatedBytesPerHour > 0 {
		return int(math.Round(*s.EstimatedBytesPerHour / *s.TotalBytesPerHour * 100))
	}
	return 0
}

// ExpandedView renders the detailed policy status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		return viewkit.RenderPolicyEmptyState(
			m.theme,
			m.db != nil,
			m.summary,
			"Enable services to start detecting waste.",
			"No waste detected. Your logs look clean.",
		)
	}

	// Detail sub-view for a single category.
	if m.detail != nil {
		return m.detail.View(width)
	}

	return viewkit.ComposeSummaryTableView(
		m.theme,
		m.renderWasteHeadline(),
		m.renderCategoryTable(width),
		m.cursorPrinciple(),
	)
}

// renderWasteHeadline renders the waste summary: waste % and pending count.
func (m *Model) renderWasteHeadline() string {
	s := m.summary
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	// Observed savings from approved policies (always shown — these are measured).
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		parts = append(parts, savingStyle.Render("saving "+saving))
	}

	// Pending count + waste %. Count first for consistency with other tabs.
	if s.PendingPolicyCount > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		pending := dot + " " + text.Render(fmt.Sprintf("%d policies", s.PendingPolicyCount))
		if wp := wastePercent(s); wp > 0 {
			pending += sep + text.Render(fmt.Sprintf("%d%% waste", wp))
		}
		parts = append(parts, pending)
	} else if wp := wastePercent(s); wp > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d%% waste", wp)))
	}

	// Analysis progress when not yet ready.
	if s.EventCount > 0 && !s.AnalysisReady() {
		pct := float64(s.AnalyzedCount) / float64(s.EventCount)
		bar := m.analysisBar()
		parts = append(parts, bar.ViewAs(pct)+" "+muted.Render(fmt.Sprintf("%d/%d analyzed", s.AnalyzedCount, s.EventCount)))
	}

	if len(parts) == 0 {
		return muted.Render("All policies reviewed")
	}

	return strings.Join(parts, sep)
}

// renderCategoryTable renders all waste categories in a single table with cursor highlighting.
func (m *Model) renderCategoryTable(width int) string {
	if len(m.categories) == 0 {
		return ""
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(35))
	tbl.Headers("Category", "Pending", "Impact", "Approved")
	tbl.SetWidth(width)

	warn := lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg)
	ok := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
	bar := m.discoveryBar()

	var totalCost float64
	if m.summary.EstimatedCostPerHour != nil {
		totalCost = *m.summary.EstimatedCostPerHour
	}

	// Find widest pending count for alignment.
	maxPendingW := 1
	for _, c := range m.categories {
		if w := len(format.Count(c.PendingCount)); w > maxPendingW {
			maxPendingW = w
		}
	}

	for i, c := range m.categories {
		dot := ok.Render("●")
		if c.PendingCount > 0 {
			dot = warn.Render("●")
		}

		name := c.Name()
		if i == m.cursor {
			name = accent.Render("▶ " + name)
		} else {
			name = dot + " " + name
		}

		// Clean categories: single checkmark row.
		if c.PendingCount == 0 && c.ApprovedCount == 0 {
			tbl.Row(name, ok.Render("✓"), "—", "—")
			continue
		}

		// Pending count with optional discovery progress bar.
		pending := fmt.Sprintf("%-*s", maxPendingW, format.Count(c.PendingCount))
		if c.TotalEvents > 0 {
			pct := int(c.EventsWithVolumes * 100 / c.TotalEvents)
			if pct < 80 {
				pending += "  " + bar.ViewAs(float64(pct)/100) + " " + muted.Render(fmt.Sprintf("%d%%", pct))
			}
		}

		tbl.Row(
			name,
			pending,
			formatCategoryCost(c, totalCost, ok, muted),
			format.Count(c.ApprovedCount),
		)
	}

	return tbl.View()
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

// discoveryBar creates a small progress bar for inline use in table cells.
func (m *Model) discoveryBar() progress.Model {
	bar := progress.New(
		progress.WithColors(m.theme.GradientStart, m.theme.GradientEnd),
		progress.WithWidth(10),
		progress.WithFillCharacters('█', '░'),
	)
	bar.ShowPercentage = false
	bar.EmptyColor = m.theme.TextMuted
	return bar
}

// formatCategoryCost returns estimated yearly cost for a category, with its
// share of total estimated waste. Shows dollar amount + percentage for
// categories ≥1% of total waste, just "<1%" for tiny categories.
func formatCategoryCost(c domain.PolicyCategoryStatus, totalCostPerHour float64, success, muted lipgloss.Style) string {
	if c.EstimatedCostPerHour == nil || *c.EstimatedCostPerHour <= 0 {
		return format.YearlyCostPtr(c.EstimatedCostPerHour)
	}

	if totalCostPerHour > 0 {
		pct := int(math.Round(*c.EstimatedCostPerHour / totalCostPerHour * 100))
		if pct <= 1 {
			return muted.Render("≤1%")
		}
		cost := success.Render(format.YearlyCostPtr(c.EstimatedCostPerHour))
		if pct < 100 {
			cost += " " + muted.Render(fmt.Sprintf("(%d%%)", pct))
		}
		return cost
	}

	return success.Render(format.YearlyCostPtr(c.EstimatedCostPerHour))
}

// cursorPrinciple returns the principle text for the currently selected category.
func (m *Model) cursorPrinciple() string {
	if m.cursor < len(m.categories) {
		return m.categories[m.cursor].Principle
	}
	return ""
}

// formatObservedSaving returns the observed savings from approved policies.
func formatObservedSaving(s domain.AccountSummary) string {
	if s.ObservedCostBefore != nil && s.ObservedCostAfter != nil {
		diff := *s.ObservedCostBefore - *s.ObservedCostAfter
		if diff > 0 {
			return format.YearlyCost(diff)
		}
		return ""
	}
	if s.ObservedVolumeBefore != nil && s.ObservedVolumeAfter != nil {
		diff := *s.ObservedVolumeBefore - *s.ObservedVolumeAfter
		if diff > 0 {
			return format.Volume(diff) + " evt/hr"
		}
	}
	return ""
}
