// Package compliance renders the compliance indicator in the status bar.
package compliance

import (
	"context"
	"encoding/json"
	"fmt"
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
	pollSource = "compliance"
)

type fetchedData struct {
	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
}

// complianceDetailLoadedMsg carries the result of an async detail fetch.
type complianceDetailLoadedMsg struct {
	cat      domain.PolicyCategoryStatus
	policies []domain.CompliancePolicy
	err      error
}

// Model renders compliance status: 4 categories (PII, Secrets, PHI, Payment Data).
type Model struct {
	theme styles.Theme
	scope log.Scope
	core  policytab.Base

	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus

	// Drawer navigation
	detail *detail // non-nil when viewing a single category's policies
}

// New creates a new compliance status model.
func New(theme styles.Theme, scope log.Scope) *Model {
	return &Model{
		theme: theme,
		scope: scope.Child("compliance"),
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
				m.core.SetHasData(len(data.categories) > 0)
			})
		},
	); handled {
		return cmd
	}

	switch msg := msg.(type) {
	case complianceDetailLoadedMsg:
		if msg.err == nil {
			m.detail = newDetail(m.theme, msg.cat, msg.policies)
		}
	}

	return nil
}

// fetchData returns a Cmd that queries compliance data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	db := m.core.DB()
	scope := m.scope
	return tabpoll.Fetch(dbTimeout, func(ctx context.Context) (fetchedData, error) {
		summary, err := db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			scope.Error("get summary", "err", err)
			return fetchedData{}, err
		}

		categories, err := db.LogEventPolicyCategoryStatuses().ListComplianceCategoryStatuses(ctx)
		if err != nil {
			scope.Error("list compliance category statuses", "err", err)
			categories = nil
		}

		// Merge observed (leaking) counts into category statuses.
		observed, err := db.LogEventPolicyCategoryStatuses().CountObservedByComplianceCategory(ctx)
		if err != nil {
			scope.Error("count observed by compliance category", "err", err)
		} else {
			for i := range categories {
				categories[i].ObservedCount = observed[categories[i].Category]
			}
		}

		return fetchedData{summary: summary, categories: categories}, nil
	})
}

// fetchDetail returns a Cmd that queries category detail off the event loop.
func (m *Model) fetchDetail(cat domain.PolicyCategoryStatus) tea.Cmd {
	db := m.core.DB()
	scope := m.scope
	return tabpoll.FetchDetail(dbTimeout, func(ctx context.Context) ([]domain.CompliancePolicy, error) {
		policies, err := db.CompliancePolicies().ListPendingPoliciesByCategory(ctx, cat.Category, 25)
		if err != nil {
			scope.Error("list pending policies by category", "category", cat.Category, "err", err)
			return nil, err
		}
		return policies, nil
	}, func(policies []domain.CompliancePolicy, err error) tea.Msg {
		if err != nil {
			return complianceDetailLoadedMsg{err: err}
		}
		return complianceDetailLoadedMsg{cat: cat, policies: policies}
	})
}

// buildStateKey builds a string key for change detection.
func (m *Model) buildStateKey(summary domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := complianceStateKey{
		Summary: complianceSummaryKey{
			EventCount:    summary.EventCount,
			AnalyzedCount: summary.AnalyzedCount,
		},
		Categories: make([]complianceCategoryKey, 0, len(cats)),
	}
	for _, c := range cats {
		key.Categories = append(key.Categories, complianceCategoryKey{
			Category:       string(c.Category),
			PendingCount:   c.PendingCount,
			ApprovedCount:  c.ApprovedCount,
			DismissedCount: c.DismissedCount,
			ObservedCount:  c.ObservedCount,
		})
	}
	data, err := json.Marshal(key)
	if err != nil {
		return ""
	}
	return string(data)
}

type complianceStateKey struct {
	Summary    complianceSummaryKey    `json:"summary"`
	Categories []complianceCategoryKey `json:"categories"`
}

type complianceSummaryKey struct {
	EventCount    int64 `json:"event_count"`
	AnalyzedCount int64 `json:"analyzed_count"`
}

type complianceCategoryKey struct {
	Category       string `json:"category"`
	PendingCount   int64  `json:"pending_count"`
	ApprovedCount  int64  `json:"approved_count"`
	DismissedCount int64  `json:"dismissed_count"`
	ObservedCount  int64  `json:"observed_count"`
}

// HasData returns true when compliance policy data has been loaded.
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

// CompactView renders the compliance indicator for the collapsed statusbar.
func (m *Model) CompactView() string {
	if !m.core.HasData() {
		return ""
	}

	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var segments []string

	leaking := totalObserved(m.categories)
	pending := totalPending(m.categories)
	atRisk := pending - leaking
	approved := totalApproved(m.categories)

	if leaking > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d leaking", leaking)))
	}

	if atRisk > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		segments = append(segments, dot+" "+muted.Render(fmt.Sprintf("%d at risk", atRisk)))
	}

	if approved > 0 {
		ok := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		segments = append(segments, ok.Render(fmt.Sprintf("%d fixed", approved)))
	}

	if len(segments) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("compliant")
	}

	return strings.Join(segments, sep)
}

// ExpandedView renders the detailed compliance status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.core.HasData() {
		return viewkit.RenderPolicyEmptyState(
			m.theme,
			m.core.DB() != nil,
			m.summary,
			"Enable services to start compliance scanning.",
			"No compliance issues detected.",
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

// renderHeadline renders the compliance summary: pending/approved counts and analysis progress.
func (m *Model) renderHeadline() string {
	s := m.summary
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	leaking := totalObserved(m.categories)
	pending := totalPending(m.categories)
	atRisk := pending - leaking
	approved := totalApproved(m.categories)

	if leaking > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d leaking", leaking)))
	}

	if atRisk > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg).Render("●")
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d at risk", atRisk)))
	}

	if approved > 0 {
		ok := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		parts = append(parts, ok.Render(fmt.Sprintf("%d fixed", approved)))
	}

	// Analysis progress when not yet ready.
	if s.EventCount > 0 && !s.AnalysisReady() {
		pct := float64(s.AnalyzedCount) / float64(s.EventCount)
		bar := m.analysisBar()
		parts = append(parts, bar.ViewAs(pct)+" "+muted.Render(fmt.Sprintf("%d/%d analyzed", s.AnalyzedCount, s.EventCount)))
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
	tbl.Headers("Category", "Leaking", "At Risk", "Approved")
	tbl.SetWidth(width)

	errStyle := lipgloss.NewStyle().Foreground(m.theme.Error).Background(m.theme.Bg)
	warn := lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg)
	ok := lipgloss.NewStyle().Foreground(m.theme.Success).Background(m.theme.Bg)
	accent := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg)

	for i, c := range m.categories {
		dot := ok.Render("●")
		if c.PendingCount > 0 {
			if c.IsLeaking() {
				dot = errStyle.Render("●")
			} else {
				dot = warn.Render("●")
			}
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

		atRisk := c.PendingCount - c.ObservedCount

		leaking := "—"
		if c.ObservedCount > 0 {
			leaking = errStyle.Render(format.Count(c.ObservedCount))
		}

		risk := "—"
		if atRisk > 0 {
			risk = warn.Render(format.Count(atRisk))
		}

		tbl.Row(
			name,
			leaking,
			risk,
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

// cursorPrinciple returns the principle text for the currently selected category.
func (m *Model) cursorPrinciple() string {
	if m.core.Cursor() < len(m.categories) {
		return m.categories[m.core.Cursor()].Principle
	}
	return ""
}

func totalObserved(cats []domain.PolicyCategoryStatus) int64 {
	var n int64
	for _, c := range cats {
		n += c.ObservedCount
	}
	return n
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
