// Package quality renders the quality indicator in the status bar.
package quality

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
	pollSource   = "quality"
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

// Model renders the quality policy status: pending count, estimated savings.
type Model struct {
	theme styles.Theme
	scope log.Scope
	db    sqlite.DB

	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
	hasData    bool
	lastState  string

	// Drawer navigation
	cursor int     // selected row in category list
	detail *detail // non-nil when viewing a single category's policies
}

// New creates a new quality status model.
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("quality"),
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
	switch msg := msg.(type) {
	case tabpoll.PollMsg:
		if msg.Source != pollSource {
			return nil
		}
		if m.db == nil {
			return nil
		}
		return tea.Batch(m.fetchData(), m.poll())

	case tabpoll.DataMsg[fetchedData]:
		if msg.Err != nil {
			return nil
		}
		key := m.stateKey(msg.Data.summary, msg.Data.categories)
		tabpoll.ApplyIfChanged(&m.lastState, key, &m.cursor, len(msg.Data.categories), func() {
			m.summary = msg.Data.summary
			m.categories = msg.Data.categories
			m.hasData = len(msg.Data.categories) > 0
		})

	case detailMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.cat, msg.policies)
		}

	default:
		_ = msg
	}

	return nil
}

// fetchData returns a Cmd that queries quality data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	db := m.db
	scope := m.scope
	return tabpoll.Fetch(dbTimeout, func(ctx context.Context) (fetchedData, error) {
		summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			scope.Error("get summary", "err", err)
			return fetchedData{}, err
		}
		categories, err := db.LogEventPolicyCategoryStatuses().ListQualityCategoryStatuses(ctx)
		if err != nil {
			scope.Error("list quality category statuses", "err", err)
			return fetchedData{}, err
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
func (m *Model) stateKey(summary domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := fmt.Sprintf("%d:%d:%d", summary.ServiceCount, summary.ActiveServices, len(cats))
	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%v:%v:%v:%d:%d",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount,
			c.EstimatedVolumePerHour, c.EstimatedBytesPerHour, c.EstimatedCostPerHour,
			c.EventsWithVolumes, c.TotalEvents)
	}
	return key
}

// HasData returns true when quality data has been loaded.
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

// CompactView renders the quality status for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	pending := totalPending(m.categories)
	if pending > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render(fmt.Sprintf("%d quality", pending))
	}

	return ""
}

// ExpandedView renders the detailed quality status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		return viewkit.RenderPolicyEmptyState(
			m.theme,
			m.db != nil,
			m.summary,
			"Enable services to start detecting quality issues.",
			"No quality issues detected.",
		)
	}

	// Detail sub-view for a single category.
	if m.detail != nil {
		return m.detail.View(width)
	}

	return viewkit.ComposeSummaryTableView(
		m.theme,
		m.renderHeadline(),
		m.renderCategoryTable(width),
		m.cursorPrinciple(),
	)
}

// renderHeadline renders the quality summary.
func (m *Model) renderHeadline() string {
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	pending := totalPending(m.categories)
	approved := totalApproved(m.categories)

	if pending > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d pending", pending)))
	}

	if approved > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("%d approved", approved)))
	}

	// Analysis progress when not yet ready.
	s := m.summary
	if s.EventCount > 0 && !s.AnalysisReady() {
		pct := float64(s.AnalyzedCount) / float64(s.EventCount)
		bar := m.discoveryBar()
		parts = append(parts, bar.ViewAs(pct)+" "+muted.Render(fmt.Sprintf("%d/%d analyzed", s.AnalyzedCount, s.EventCount)))
	}

	if len(parts) == 0 {
		return muted.Render("All quality policies reviewed")
	}

	return strings.Join(parts, sep)
}

// renderCategoryTable renders quality categories in a table with cursor highlighting.
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

	totalCost := totalEstimatedCost(m.categories)

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

func totalPending(cats []domain.PolicyCategoryStatus) int64 {
	var n int64
	for _, c := range cats {
		n += c.PendingCount
	}
	return n
}

func totalApproved(cats []domain.PolicyCategoryStatus) int64 {
	var n int64
	for _, c := range cats {
		n += c.ApprovedCount
	}
	return n
}

func totalEstimatedCost(cats []domain.PolicyCategoryStatus) float64 {
	var total float64
	for _, c := range cats {
		if c.EstimatedCostPerHour != nil {
			total += *c.EstimatedCostPerHour
		}
	}
	return total
}

// cursorPrinciple returns the principle text for the currently selected category.
func (m *Model) cursorPrinciple() string {
	if m.cursor < len(m.categories) {
		return m.categories[m.cursor].Principle
	}
	return ""
}

// formatCategoryCost returns estimated yearly cost for a category, with its
// share of total estimated savings.
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
