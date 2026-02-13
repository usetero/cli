// Package execute provides the sequential policy approval execution step.
package execute

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/policyapproval/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// Model handles sequential policy approval — the final step of the wizard.
type Model struct {
	ctx      context.Context
	theme    styles.Theme
	db       sqlite.DB
	scope    log.Scope
	userID   string
	policies []domain.PolicyDetail
	current  int      // index of policy being approved
	results  []result // per-policy outcome
	done     bool
	width    int
	height   int

	// Accumulated for completion message
	toolUseID    string
	totalSavings float64
}

type result struct {
	err error
}

// policyApprovedMsg is sent when a single policy approval completes.
type policyApprovedMsg struct {
	index int
	err   error
}

// New creates a new execute step.
func New(
	ctx context.Context,
	theme styles.Theme,
	db sqlite.DB,
	userID string,
	toolUseID string,
	policies []domain.PolicyDetail,
	scope log.Scope,
) *Model {
	// Pre-calculate total savings.
	var totalSavings float64
	for _, p := range policies {
		totalSavings += p.EstimatedCostPerHour * 8760
	}

	return &Model{
		ctx:          ctx,
		theme:        theme,
		db:           db,
		userID:       userID,
		toolUseID:    toolUseID,
		policies:     policies,
		results:      make([]result, len(policies)),
		scope:        scope.Child("execute"),
		totalSavings: totalSavings,
	}
}

// Init starts approving the first policy.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("execute step started", "count", len(m.policies))
	if len(m.policies) == 0 {
		return m.complete()
	}
	return m.approveNext()
}

// Update handles approval results and progresses through policies.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case policyApprovedMsg:
		m.results[msg.index] = result{err: msg.err}
		if msg.err != nil {
			m.scope.Error("policy approval failed",
				"index", msg.index,
				"policy_id", m.policies[msg.index].PolicyID,
				"error", msg.err,
			)
		} else {
			m.scope.Info("policy approved",
				"index", msg.index,
				"policy_id", m.policies[msg.index].PolicyID,
			)
		}

		m.current++
		if m.current < len(m.policies) {
			return m.approveNext()
		}
		m.done = true
		return m.complete()
	}
	return nil
}

// approveNext returns a command that approves the current policy.
func (m *Model) approveNext() tea.Cmd {
	idx := m.current
	policy := m.policies[idx]
	return func() tea.Msg {
		err := m.db.LogEventPolicies().Approve(m.ctx, policy.PolicyID, m.userID)
		return policyApprovedMsg{index: idx, err: err}
	}
}

// complete emits the PolicyApprovalComplete message with counts.
func (m *Model) complete() tea.Cmd {
	var approved, failed int
	for _, r := range m.results {
		if r.err != nil {
			failed++
		} else {
			approved++
		}
	}
	m.scope.Info("execution complete", "approved", approved, "failed", failed)
	return func() tea.Msg {
		return msgs.PolicyApprovalComplete{
			ToolUseID:     m.toolUseID,
			ApprovedCount: approved,
			FailedCount:   failed,
			TotalSavings:  m.totalSavings,
		}
	}
}

// View renders the progress display.
func (m *Model) View() string {
	colors := m.theme

	header := lipgloss.NewStyle().
		Foreground(colors.Text).
		Bold(true).
		Render("Approving policies...")

	// Determine visible window based on height.
	const chrome = 5 // header + blank + blank + counter + margin
	maxVisible := m.height - chrome
	if maxVisible < 3 {
		maxVisible = 3
	}

	// Scroll window: keep current item visible with some context.
	start := 0
	end := len(m.policies)
	if end > maxVisible {
		// Center around current item.
		start = m.current - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.policies) {
			end = len(m.policies)
			start = end - maxVisible
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		line := m.renderPolicyLine(i)
		lines = append(lines, line)
	}

	policyList := strings.Join(lines, "\n")

	// Counter
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted)
	counter := muted.Render(fmt.Sprintf("%d of %d", m.current, len(m.policies)))

	return lipgloss.JoinVertical(lipgloss.Left, header, "", policyList, "", counter)
}

// renderPolicyLine renders a single policy line with status symbol.
func (m *Model) renderPolicyLine(index int) string {
	colors := m.theme
	name := m.policies[index].LogEventName

	var symbol string
	var style lipgloss.Style

	switch {
	case index < m.current:
		// Completed
		if m.results[index].err != nil {
			symbol = "✗"
			style = lipgloss.NewStyle().Foreground(colors.ErrorFg)
		} else {
			symbol = "✓"
			style = lipgloss.NewStyle().Foreground(colors.SuccessFg)
		}
	case index == m.current && !m.done:
		// In progress
		symbol = "◐"
		style = lipgloss.NewStyle().Foreground(colors.Accent)
	default:
		// Queued
		symbol = "○"
		style = lipgloss.NewStyle().Foreground(colors.TextMuted)
	}

	return style.Render(fmt.Sprintf("%s %s", symbol, name))
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns key bindings (none during execution).
func (m *Model) ShortHelp() []key.Binding {
	return nil
}
