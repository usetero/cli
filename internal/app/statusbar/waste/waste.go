// Package waste renders the waste indicator in the status bar.
package waste

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/app/statusbar/policytab"
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
	dbTimeout  = 2 * time.Second
	pollSource = "waste"
)

type fetchedData struct {
	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
}

// wasteDetailLoadedMsg carries the result of an async detail fetch.
type wasteDetailLoadedMsg struct {
	cat      domain.PolicyCategoryStatus
	policies []domain.WastePolicy
	err      error
}

// Model renders the policy status: pending count, estimated savings, observed savings.
type Model struct {
	theme styles.Theme
	scope log.Scope
	core  policytab.Base

	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus

	// Drawer navigation
	detail *detail // non-nil when viewing a single category's policies
}

// New creates a new policy status model.
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("waste"),
		core:  policytab.New(pollSource),
	}
}

// SetDB sets the database and starts polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	return m.core.SetDB(db)
}

// Init starts polling.
func (m *Model) Init() tea.Cmd {
	return m.core.Init()
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if cmd, handled := policytab.UpdatePoll(&m.core,
		msg,
		m.fetchData(),
		func(data fetchedData) {
			key := m.buildStateKey(data.summary, data.categories)
			m.core.ApplyIfChanged(key, len(data.categories), func() {
				m.summary = data.summary
				m.categories = data.categories
				m.core.SetHasData(len(data.categories) > 0 || data.summary.PendingPolicyCount+data.summary.ApprovedPolicyCount+data.summary.DismissedPolicyCount > 0)
			})
		},
	); handled {
		return cmd
	}

	switch msg := msg.(type) {
	case wasteDetailLoadedMsg:
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
	db := m.core.DB()
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
	db := m.core.DB()
	scope := m.scope
	return tabpoll.FetchDetail(dbTimeout, func(ctx context.Context) ([]domain.WastePolicy, error) {
		policies, err := db.LogEventPolicyStatuses().ListTopPendingPoliciesByCategory(ctx, cat.Category, 25)
		if err != nil {
			scope.Error("list top pending policies", "category", cat.Category, "err", err)
			return nil, err
		}
		return policies, nil
	}, func(policies []domain.WastePolicy, err error) tea.Msg {
		if err != nil {
			return wasteDetailLoadedMsg{err: err}
		}
		return wasteDetailLoadedMsg{cat: cat, policies: policies}
	})
}

// buildStateKey builds a string key for change detection.
func (m *Model) buildStateKey(s domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := wasteStateKey{
		Summary: wasteSummaryKey{
			ReadyForUse:             s.ReadyForUse,
			EventCount:              s.EventCount,
			AnalyzedCount:           s.AnalyzedCount,
			PendingPolicyCount:      s.PendingPolicyCount,
			ApprovedPolicyCount:     s.ApprovedPolicyCount,
			DismissedPolicyCount:    s.DismissedPolicyCount,
			EstimatedCostPerHour:    s.EstimatedCostPerHour,
			EstimatedCostPerHourB:   s.EstimatedCostPerHourBytes,
			EstimatedCostPerHourVol: s.EstimatedCostPerHourVolume,
			EstimatedVolumePerHour:  s.EstimatedVolumePerHour,
			EstimatedBytesPerHour:   s.EstimatedBytesPerHour,
			TotalCostPerHour:        s.TotalCostPerHour,
			TotalVolumePerHour:      s.TotalVolumePerHour,
			TotalBytesPerHour:       s.TotalBytesPerHour,
		},
		Categories: make([]wasteCategoryKey, 0, len(cats)),
	}
	for _, c := range cats {
		key.Categories = append(key.Categories, wasteCategoryKey{
			Category:               string(c.Category),
			PendingCount:           c.PendingCount,
			ApprovedCount:          c.ApprovedCount,
			DismissedCount:         c.DismissedCount,
			EstimatedVolumePerHour: c.EstimatedVolumePerHour,
			EstimatedBytesPerHour:  c.EstimatedBytesPerHour,
			EstimatedCostPerHour:   c.EstimatedCostPerHour,
			EventsWithVolumes:      c.EventsWithVolumes,
			TotalEvents:            c.TotalEvents,
		})
	}
	data, err := json.Marshal(key)
	if err != nil {
		return ""
	}
	return string(data)
}

type wasteStateKey struct {
	Summary    wasteSummaryKey    `json:"summary"`
	Categories []wasteCategoryKey `json:"categories"`
}

type wasteSummaryKey struct {
	ReadyForUse             bool     `json:"ready_for_use"`
	EventCount              int64    `json:"event_count"`
	AnalyzedCount           int64    `json:"analyzed_count"`
	PendingPolicyCount      int64    `json:"pending_policy_count"`
	ApprovedPolicyCount     int64    `json:"approved_policy_count"`
	DismissedPolicyCount    int64    `json:"dismissed_policy_count"`
	EstimatedCostPerHour    *float64 `json:"estimated_cost_per_hour"`
	EstimatedCostPerHourB   *float64 `json:"estimated_cost_per_hour_bytes"`
	EstimatedCostPerHourVol *float64 `json:"estimated_cost_per_hour_volume"`
	EstimatedVolumePerHour  *float64 `json:"estimated_volume_per_hour"`
	EstimatedBytesPerHour   *float64 `json:"estimated_bytes_per_hour"`
	TotalCostPerHour        *float64 `json:"total_cost_per_hour"`
	TotalVolumePerHour      *float64 `json:"total_volume_per_hour"`
	TotalBytesPerHour       *float64 `json:"total_bytes_per_hour"`
}

type wasteCategoryKey struct {
	Category               string   `json:"category"`
	PendingCount           int64    `json:"pending_count"`
	ApprovedCount          int64    `json:"approved_count"`
	DismissedCount         int64    `json:"dismissed_count"`
	EstimatedVolumePerHour *float64 `json:"estimated_volume_per_hour"`
	EstimatedBytesPerHour  *float64 `json:"estimated_bytes_per_hour"`
	EstimatedCostPerHour   *float64 `json:"estimated_cost_per_hour"`
	EventsWithVolumes      int64    `json:"events_with_volumes"`
	TotalEvents            int64    `json:"total_events"`
}

// HasData returns true when policy data has been loaded.
func (m *Model) HasData() bool {
	return m.core.HasData()
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
	return m.core.NavController(
		func() int { return len(m.categories) },
		func(index int) tea.Cmd {
			cat := m.categories[index]
			if cat.PendingCount == 0 {
				return nil
			}
			return m.fetchDetail(cat)
		},
		func() listdetail.Detail { return m.detail },
		func() { m.detail = nil },
	)
}

// CompactView renders the policy status for the statusbar.
func (m *Model) CompactView() string {
	if !m.core.HasData() {
		return ""
	}

	s := m.summary
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var segments []string

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
	if !m.core.HasData() {
		return viewkit.RenderPolicyEmptyState(
			m.theme,
			m.core.DB() != nil,
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
		if i == m.core.Cursor() {
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
	if m.core.Cursor() < len(m.categories) {
		return m.categories[m.core.Cursor()].Principle
	}
	return ""
}
