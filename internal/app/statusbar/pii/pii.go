// Package pii renders the PII leakage indicator in the status bar.
package pii

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a PII policy check.
type pollMsg struct{}

// Model renders PII leakage findings: per-policy detail with service/event context.
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	policies  []domain.PIIPolicy
	hasData   bool
	lastState string
}

// New creates a new PII status model.
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

		policies, err := m.db.LogEventPolicies().ListPIIPolicies(ctx)
		if err != nil {
			return m.poll()
		}

		key := m.stateKey(policies)
		if key != m.lastState {
			m.policies = policies
			m.hasData = len(policies) > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(policies []domain.PIIPolicy) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d", len(policies))
	for _, p := range policies {
		fmt.Fprintf(&b, "|%s:%s:%s:%d",
			p.ServiceName, p.LogEventName, p.Status, len(p.Fields))
		for _, f := range p.Fields {
			fmt.Fprintf(&b, ":%s", f.PIIType)
		}
	}
	return b.String()
}

// HasData returns true when PII policy data has been loaded.
func (m *Model) HasData() bool {
	return m.hasData
}

// CompactView renders the PII indicator for the collapsed statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	pending := m.pendingCount()
	if pending == 0 {
		return ""
	}

	colors := m.theme
	dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	return dot + " " + muted.Render(fmt.Sprintf("%d PII", pending))
}

// ExpandedView renders the detailed PII status for the drawer.
func (m *Model) ExpandedView(width, height int) string {
	if !m.hasData {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		if m.db == nil {
			return muted.Render("Waiting for sync to start...")
		}
		return muted.Render("No PII leakage detected.")
	}

	var lines []string
	lines = append(lines, m.renderHeadline())
	lines = append(lines, "")

	if tbl := m.renderTable(width); tbl != "" {
		lines = append(lines, tbl)
	}

	return strings.Join(lines, "\n")
}

// renderHeadline renders the PII summary line.
func (m *Model) renderHeadline() string {
	colors := m.theme

	pending := m.pendingCount()
	total := len(m.policies)
	services := m.uniqueServiceCount()

	var parts []string

	if pending > 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
		text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
		parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d PII findings across %d services", total, services)))
	} else {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
		parts = append(parts, dot+" "+muted.Render(fmt.Sprintf("%d PII findings reviewed", total)))
	}

	return strings.Join(parts, "")
}

// renderTable renders the per-policy PII table.
func (m *Model) renderTable(width int) string {
	if len(m.policies) == 0 {
		return ""
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(40))
	tbl.Headers("Log Event", "Service", "Leaking")
	tbl.SetWidth(width)

	for _, p := range m.policies {
		tbl.Row(
			p.LogEventName,
			p.ServiceName,
			m.formatPIITypes(p.Fields, 3),
		)
	}

	return tbl.View()
}

// piiEntry holds a deduplicated PII type with its highest severity.
type piiEntry struct {
	label    string
	severity domain.PIISeverity
}

// formatPIITypes returns deduplicated, severity-colored PII type labels,
// showing at most maxShow before truncating with "+N".
func (m *Model) formatPIITypes(fields []domain.PIIField, maxShow int) string {
	if len(fields) == 0 {
		return "—"
	}

	// Deduplicate and preserve first-seen order, keeping highest severity.
	seen := make(map[string]int) // label → index in entries
	var entries []piiEntry
	for _, f := range fields {
		label := displayPIIType(f.PIIType)
		if idx, ok := seen[label]; ok {
			if f.Severity() > entries[idx].severity {
				entries[idx].severity = f.Severity()
			}
		} else {
			seen[label] = len(entries)
			entries = append(entries, piiEntry{label, f.Severity()})
		}
	}

	visible := entries
	remaining := 0
	if len(entries) > maxShow {
		visible = entries[:maxShow]
		remaining = len(entries) - maxShow
	}

	parts := make([]string, len(visible))
	for i, e := range visible {
		parts[i] = m.colorSeverity(e.label, e.severity)
	}

	result := strings.Join(parts, ", ")
	if remaining > 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		result += muted.Render(fmt.Sprintf(", +%d", remaining))
	}
	return result
}

// colorSeverity applies theme color based on PII severity.
func (m *Model) colorSeverity(label string, s domain.PIISeverity) string {
	switch s {
	case domain.PIISeverityCritical:
		return lipgloss.NewStyle().Foreground(m.theme.Error).Background(m.theme.Bg).Render(label)
	case domain.PIISeverityHigh:
		return lipgloss.NewStyle().Foreground(m.theme.Warning).Background(m.theme.Bg).Render(label)
	default:
		return label // table cell style handles default color
	}
}

// displayPIIType returns a human-readable label for a pii_type value.
func displayPIIType(t string) string {
	switch t {
	case domain.PIITypeCreditCard:
		return "credit card"
	case domain.PIITypeIPAddress:
		return "IP address"
	case domain.PIITypeSSN:
		return "SSN"
	case domain.PIITypeDateOfBirth:
		return "date of birth"
	default:
		return t
	}
}

func (m *Model) pendingCount() int {
	count := 0
	for _, p := range m.policies {
		if p.Status == domain.PolicyLogStatusPending {
			count++
		}
	}
	return count
}

func (m *Model) uniqueServiceCount() int {
	seen := make(map[string]struct{})
	for _, p := range m.policies {
		seen[p.ServiceName] = struct{}{}
	}
	return len(seen)
}
