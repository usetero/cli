// Package categorydetail provides the category detail step for the policy approval wizard.
package categorydetail

import (
	"context"
	"fmt"
	"math"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/policyapproval/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/tableselect"
)

// Model handles the category detail view — the second step of the approval wizard.
// Shows individual pending policies within a selected category. User can toggle
// selection with space, select all/none, then press enter to proceed.
type Model struct {
	ctx      context.Context
	theme    styles.Theme
	db       sqlite.DB
	scope    log.Scope
	category string
	policies []domain.PolicyDetail
	selected map[string]bool
	tbl      *tableselect.Model
	width    int
	height   int
	err      error
}

// New creates a new category detail step.
func New(ctx context.Context, theme styles.Theme, db sqlite.DB, category string, scope log.Scope) *Model {
	return &Model{
		ctx:      ctx,
		theme:    theme,
		db:       db,
		category: category,
		scope:    scope.Child("categorydetail"),
		selected: make(map[string]bool),
	}
}

// policiesLoadedMsg is sent when policies are loaded from the database.
type policiesLoadedMsg struct {
	policies []domain.PolicyDetail
	err      error
}

// Init loads policies for the selected category.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("category detail step started", "category", m.category)
	return m.loadPolicies()
}

func (m *Model) loadPolicies() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.db.LogEventPolicies().ListPendingByCategory(m.ctx, m.category)
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.scope.Error("failed to load policies", "error", msg.err)
			return nil
		}
		m.policies = msg.policies
		m.scope.Info("policies loaded", "count", len(m.policies))

		if len(m.policies) == 0 {
			return func() tea.Msg { return msgs.BackToSummary{} }
		}

		// Select all by default.
		for _, p := range m.policies {
			m.selected[p.PolicyID] = true
		}
		m.buildTable()
		return nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keySpace):
			if m.tbl != nil && len(m.policies) > 0 {
				p := m.policies[m.tbl.Cursor()]
				m.selected[p.PolicyID] = !m.selected[p.PolicyID]
				m.rebuildRows()
			}
		case key.Matches(msg, keySelectAll):
			for _, p := range m.policies {
				m.selected[p.PolicyID] = true
			}
			m.rebuildRows()
		case key.Matches(msg, keySelectNone):
			for _, p := range m.policies {
				m.selected[p.PolicyID] = false
			}
			m.rebuildRows()
		case key.Matches(msg, keyEnter):
			return m.submit()
		case key.Matches(msg, keyEscape):
			return func() tea.Msg { return msgs.BackToSummary{} }
		default:
			if m.tbl != nil {
				return m.tbl.Update(msg)
			}
		}
	}
	return nil
}

// submit emits PoliciesSelected with the currently selected policy IDs.
func (m *Model) submit() tea.Cmd {
	var ids []string
	for _, p := range m.policies {
		if m.selected[p.PolicyID] {
			ids = append(ids, p.PolicyID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	m.scope.Info("policies selected", "count", len(ids))
	return func() tea.Msg {
		return msgs.PoliciesSelected{PolicyIDs: ids}
	}
}

// buildTable creates the selectable table from loaded policies.
func (m *Model) buildTable() {
	cols := []tableselect.Column{
		{Title: " ", Width: 3},
		{Title: "Log Event", Width: 30},
		{Title: "Risk", Width: 8},
		{Title: "Savings", Width: 12},
	}

	rows := m.makeRows()
	m.tbl = tableselect.New(m.theme, cols, rows)
}

// rebuildRows updates table rows with current checkbox state.
func (m *Model) rebuildRows() {
	if m.tbl != nil {
		m.tbl.SetRows(m.makeRows())
	}
}

// makeRows builds table rows from policies with checkbox state.
func (m *Model) makeRows() []tableselect.Row {
	rows := make([]tableselect.Row, len(m.policies))
	for i, p := range m.policies {
		check := "☐"
		if m.selected[p.PolicyID] {
			check = "☑"
		}
		rows[i] = tableselect.Row{
			check,
			truncate(p.LogEventName, 28),
			p.RiskLevel.String(),
			formatSavings(p.EstimatedCostPerHour),
		}
	}
	return rows
}

// View renders the category detail table.
func (m *Model) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(m.theme.ErrorBg)
		return errStyle.Render(fmt.Sprintf("Error loading policies: %v", m.err))
	}

	if m.tbl == nil {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		return muted.Render("Loading policies...")
	}

	if len(m.policies) == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		return muted.Render("No pending policies in this category.")
	}

	var lines []string

	// Header.
	header := lipgloss.NewStyle().Foreground(m.theme.Text).Bold(true)
	lines = append(lines, header.Render(m.category))
	lines = append(lines, "")

	// Policy table.
	lines = append(lines, m.tbl.View())
	lines = append(lines, "")

	// Footer: selection count.
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	selectedCount := 0
	for _, sel := range m.selected {
		if sel {
			selectedCount++
		}
	}
	footer := fmt.Sprintf("%d of %d selected", selectedCount, len(m.policies))
	lines = append(lines, muted.Render(footer))

	return strings.Join(lines, "\n")
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keySpace, keySelectAll, keySelectNone, keyEnter, keyEscape}
}

// Key bindings.
var (
	keySpace      = key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle"))
	keySelectAll  = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all"))
	keySelectNone = key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "select none"))
	keyEnter      = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))
	keyEscape     = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
)

// formatSavings returns estimated yearly cost savings.
func formatSavings(costPerHour float64) string {
	if costPerHour > 0 {
		yearly := costPerHour * 8760
		if yearly >= 1 {
			return "~" + formatCost(yearly) + "/yr"
		}
	}
	return "—"
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

// truncate shortens a string to max length, adding ellipsis if needed.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
