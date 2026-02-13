// Package categorysummary provides the category summary step for the policy approval wizard.
package categorysummary

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

// Model handles the category summary view — the first step of the approval wizard.
// Shows policy categories grouped from log_event_policy_statuses_cache, sorted by
// pending count. User can drill into a category or approve all low-risk at once.
type Model struct {
	ctx        context.Context
	theme      styles.Theme
	db         sqlite.DB
	scope      log.Scope
	categories []domain.PolicyCategoryStatus
	tbl        *tableselect.Model
	width      int
	height     int
	err        error
}

// New creates a new category summary step.
func New(ctx context.Context, theme styles.Theme, db sqlite.DB, scope log.Scope) *Model {
	return &Model{
		ctx:   ctx,
		theme: theme,
		db:    db,
		scope: scope.Child("categorysummary"),
	}
}

// categoriesLoadedMsg is sent when categories are loaded from the database.
type categoriesLoadedMsg struct {
	categories []domain.PolicyCategoryStatus
	err        error
}

// Init loads category data from the local SQLite database.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("category summary step started")
	return m.loadCategories()
}

func (m *Model) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.db.LogEventPolicies().ListCategoryStatuses(m.ctx)
		return categoriesLoadedMsg{categories: categories, err: err}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case categoriesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.scope.Error("failed to load categories", "error", msg.err)
			return nil
		}
		// Filter to categories with pending policies only.
		var pending []domain.PolicyCategoryStatus
		for _, c := range msg.categories {
			if c.PendingCount > 0 {
				pending = append(pending, c)
			}
		}
		m.categories = pending
		m.scope.Info("categories loaded", "count", len(m.categories))

		if len(m.categories) == 0 {
			// Nothing to approve — cancel the wizard.
			return func() tea.Msg { return msgs.Cancelled{} }
		}

		m.buildTable()
		return nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyEnter):
			if m.tbl != nil && len(m.categories) > 0 {
				selected := m.categories[m.tbl.Cursor()]
				m.scope.Info("category selected", "category", selected.Category)
				return func() tea.Msg {
					return msgs.CategorySelected{Category: selected.Category}
				}
			}
		case key.Matches(msg, keyApproveAllLowRisk):
			return m.approveAllLowRisk()
		case key.Matches(msg, keyEscape):
			return func() tea.Msg { return msgs.Cancelled{} }
		default:
			// Delegate navigation to the table.
			if m.tbl != nil {
				return m.tbl.Update(msg)
			}
		}
	}
	return nil
}

// buildTable creates the selectable table from loaded categories.
func (m *Model) buildTable() {
	cols := []tableselect.Column{
		{Title: "Category", Width: 25},
		{Title: "Risk", Width: 8},
		{Title: "Pending", Width: 8},
		{Title: "Savings", Width: 12},
	}

	rows := make([]tableselect.Row, len(m.categories))
	for i, c := range m.categories {
		rows[i] = tableselect.Row{
			c.Category,
			c.RiskLevel.String(),
			fmt.Sprintf("%d", c.PendingCount),
			formatSavings(c),
		}
	}

	m.tbl = tableselect.New(m.theme, cols, rows)
}

// approveAllLowRisk collects all policy IDs from low-risk categories.
func (m *Model) approveAllLowRisk() tea.Cmd {
	var lowRiskCategories []string
	for _, c := range m.categories {
		if c.RiskLevel == domain.RiskLevelLow && c.PendingCount > 0 {
			lowRiskCategories = append(lowRiskCategories, c.Category)
		}
	}
	if len(lowRiskCategories) == 0 {
		return nil
	}
	m.scope.Info("approve all low-risk", "categories", len(lowRiskCategories))
	return func() tea.Msg {
		return msgs.ApproveAllLowRisk{Categories: lowRiskCategories}
	}
}

// View renders the category summary table.
func (m *Model) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(m.theme.ErrorBg)
		return errStyle.Render(fmt.Sprintf("Error loading policies: %v", m.err))
	}

	if m.tbl == nil {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		return muted.Render("Loading policies...")
	}

	if len(m.categories) == 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		return muted.Render("No pending policies.")
	}

	var lines []string

	// Header.
	header := lipgloss.NewStyle().Foreground(m.theme.Text).Bold(true)
	lines = append(lines, header.Render("Review policy categories"))
	lines = append(lines, "")

	// Category table.
	lines = append(lines, m.tbl.View())
	lines = append(lines, "")

	// Footer: totals.
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	totalPending, totalCost := m.totals()
	footer := fmt.Sprintf("Total: %d pending", totalPending)
	if totalCost > 0 {
		footer += fmt.Sprintf(" · Est. savings: ~%s/yr", formatCost(totalCost*8760))
	}
	lines = append(lines, muted.Render(footer))

	return strings.Join(lines, "\n")
}

// totals returns aggregate pending count and hourly cost across all categories.
func (m *Model) totals() (int64, float64) {
	var pending int64
	var cost float64
	for _, c := range m.categories {
		pending += c.PendingCount
		cost += c.EstimatedCostPerHour
	}
	return pending, cost
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{keyEnter, keyApproveAllLowRisk, keyEscape}
}

// Key bindings.
var (
	keyEnter             = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))
	keyApproveAllLowRisk = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "approve all low-risk"))
	keyEscape            = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

// formatSavings returns estimated yearly cost savings for a category.
func formatSavings(c domain.PolicyCategoryStatus) string {
	if c.EstimatedCostPerHour > 0 {
		yearly := c.EstimatedCostPerHour * 8760
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
