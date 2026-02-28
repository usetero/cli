// Package compliance renders the compliance indicator in the status bar.
package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/statusbar/listdetail"
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
	pollSource   = "compliance"
)

type fetchedData struct {
	summary    domain.AccountSummary
	categories []domain.PolicyCategoryStatus
}

// detailMsg carries the result of an async detail fetch.
type detailMsg struct {
	cat      domain.PolicyCategoryStatus
	policies []domain.CompliancePolicy
	err      error
}

// Model renders compliance status: 4 categories (PII, Secrets, PHI, Payment Data).
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
	}

	return nil
}

// fetchData returns a Cmd that queries compliance data off the event loop.
func (m *Model) fetchData() tea.Cmd {
	db := m.db
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
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		ctx, cancel := sqlite.WithTimeout(context.Background(), dbTimeout)
		defer cancel()
		policies, err := db.CompliancePolicies().ListPendingPoliciesByCategory(ctx, cat.Category, 25)
		if err != nil {
			scope.Error("list pending policies by category", "category", cat.Category, "err", err)
			return detailMsg{err: err}
		}
		return detailMsg{cat: cat, policies: policies}
	}
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(summary domain.AccountSummary, cats []domain.PolicyCategoryStatus) string {
	key := fmt.Sprintf("%d:%d", summary.EventCount, summary.AnalyzedCount)

	for _, c := range cats {
		key += fmt.Sprintf("|%s:%d:%d:%d:%d",
			c.Category, c.PendingCount, c.ApprovedCount, c.DismissedCount, c.ObservedCount)
	}

	return key
}

// HasData returns true when compliance policy data has been loaded.
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

// CompactView renders the compliance indicator for the collapsed statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
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
	if !m.hasData {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		if m.db == nil {
			return muted.Render("Waiting for sync to start...")
		}
		if m.summary.ActiveServices == 0 && m.summary.ServiceCount > 0 {
			return muted.Render(fmt.Sprintf(
				"%d services discovered, all disabled.\nEnable services to start compliance scanning.",
				m.summary.ServiceCount,
			))
		}
		if m.summary.ActiveServices == 0 {
			return muted.Render("No services discovered yet.")
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

	if desc := m.cursorPrinciple(); desc != "" {
		lines = append(lines, "")
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		lines = append(lines, muted.Render(desc))
	}

	return strings.Join(lines, "\n")
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
	if m.cursor < len(m.categories) {
		return m.categories[m.cursor].Principle
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
