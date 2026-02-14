// Package pii renders the PII leakage indicator in the status bar.
package pii

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

const (
	pollInterval = 2 * time.Second

	// discoveryDoneThreshold is the classification coverage percentage above
	// which we consider discovery complete. Volume ratios are never exactly
	// 100% due to throughput fluctuations.
	discoveryDoneThreshold = 95
)

// pollMsg triggers a PII policy check.
type pollMsg struct{}

// Model renders PII leakage findings: per-policy detail with service/event context.
type Model struct {
	theme styles.Theme
	db    sqlite.DB

	summary    domain.AccountSummary
	policies   []domain.PIIPolicy // pending only, sorted by observed then volume
	fixedCount int64              // number of approved PII policies
	hasData    bool
	lastState  string
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

		summary, err := m.db.DatadogAccountStatuses().GetSummary(ctx)
		if err != nil {
			return m.poll()
		}

		policies, err := m.db.LogEventPolicies().ListPendingPIIPolicies(ctx)
		if err != nil {
			return m.poll()
		}

		fixedCount, err := m.db.LogEventPolicies().CountFixedPIIPolicies(ctx)
		if err != nil {
			fixedCount = 0
		}

		key := m.stateKey(summary, policies, fixedCount)
		if key != m.lastState {
			m.summary = summary
			m.policies = policies
			m.fixedCount = fixedCount
			m.hasData = len(policies) > 0 || fixedCount > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// stateKey builds a string key for change detection.
func (m *Model) stateKey(summary domain.AccountSummary, policies []domain.PIIPolicy, fixedCount int64) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%.0f:%.0f:%d:%d",
		ptrVal(summary.TotalServiceVolumePerHour), summary.TotalVolumePerHour,
		len(policies), fixedCount)
	for _, p := range policies {
		fmt.Fprintf(&b, "|%s:%s:%d:%.0f:%v:%v",
			p.ServiceName, p.LogEventName, len(p.Fields), p.VolumePerHour, p.AnyObserved, p.HasVolumes)
		for _, f := range p.Fields {
			fmt.Fprintf(&b, ":%v", f.Observed)
			for _, t := range f.PIITypes {
				fmt.Fprintf(&b, ":%s", t)
			}
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
	if !m.hasData || len(m.policies) == 0 {
		return ""
	}

	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	// Red dot if any policy has observed PII, warning otherwise.
	var anyObserved bool
	for _, p := range m.policies {
		if p.AnyObserved {
			anyObserved = true
			break
		}
	}
	dotColor := colors.Warning
	if anyObserved {
		dotColor = colors.Error
	}
	dot := lipgloss.NewStyle().Foreground(dotColor).Background(colors.Bg).Render("●")

	return dot + " " + muted.Render(fmt.Sprintf("%d PII", len(m.policies)))
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

// renderHeadline renders the PII summary line: leaking vs at-risk counts.
func (m *Model) renderHeadline() string {
	colors := m.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	sep := muted.Render(" · ")

	var parts []string

	if len(m.policies) > 0 {
		var leaking, atRisk int
		for _, p := range m.policies {
			if p.AnyObserved {
				leaking++
			} else {
				atRisk++
			}
		}

		if leaking > 0 {
			dot := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg).Render("●")
			text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
			parts = append(parts, dot+" "+text.Render(fmt.Sprintf("%d leaking", leaking)))
		}
		if atRisk > 0 {
			dot := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg).Render("●")
			parts = append(parts, dot+" "+muted.Render(fmt.Sprintf("%d at risk", atRisk)))
		}

		services := m.uniqueServiceCount()
		parts = append(parts, muted.Render(fmt.Sprintf("across %d services", services)))

		// Discovery progress when classification coverage is below threshold.
		if pct := summaryDiscoveryPercent(m.summary); pct < discoveryDoneThreshold {
			bar := m.discoveryBar()
			parts = append(parts, bar.ViewAs(float64(pct)/100)+" "+muted.Render(fmt.Sprintf("%d%%", pct)))
		}
	}

	if m.fixedCount > 0 {
		ok := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg)
		parts = append(parts, ok.Render(fmt.Sprintf("%d fixed", m.fixedCount)))
	}

	if len(parts) == 0 {
		dot := lipgloss.NewStyle().Foreground(colors.Success).Background(colors.Bg).Render("●")
		return dot + " " + muted.Render("No PII leakage detected.")
	}

	return strings.Join(parts, sep)
}

// renderTable renders the per-policy PII table.
func (m *Model) renderTable(width int) string {
	if len(m.policies) == 0 {
		return ""
	}

	tbl := table.New(m.theme, table.WithMaxValueWidth(40))
	tbl.Headers("Log Event", "Service", "Volume", "Leaking")
	tbl.SetWidth(width)

	for _, p := range m.policies {
		vol := "—"
		if p.HasVolumes {
			vol = format.Volume(p.VolumePerHour) + " evt/hr"
		}
		tbl.Row(
			p.LogEventName,
			p.ServiceName,
			vol,
			m.observedDot(p.AnyObserved)+" "+m.formatPIITypes(p.Fields, 3),
		)
	}

	return tbl.View()
}

// observedDot returns a colored dot based on whether PII was observed.
// Red for observed (leaking), muted for at-risk.
func (m *Model) observedDot(observed bool) string {
	if observed {
		return lipgloss.NewStyle().Foreground(m.theme.Error).Background(m.theme.Bg).Render("●")
	}
	return lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg).Render("●")
}

// formatPIITypes returns deduplicated PII type labels from all fields,
// showing at most maxShow before truncating with "+N".
func (m *Model) formatPIITypes(fields []domain.PIIField, maxShow int) string {
	if len(fields) == 0 {
		return "—"
	}

	// Flatten all pii_types across fields, deduplicate, preserve first-seen order.
	seen := make(map[string]struct{})
	var types []string
	for _, f := range fields {
		for _, t := range f.PIITypes {
			label := displayPIIType(t)
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{}
				types = append(types, label)
			}
		}
	}

	if len(types) == 0 {
		return "—"
	}

	visible := types
	remaining := 0
	if len(types) > maxShow {
		visible = types[:maxShow]
		remaining = len(types) - maxShow
	}

	result := strings.Join(visible, ", ")
	if remaining > 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)
		result += muted.Render(fmt.Sprintf(", +%d", remaining))
	}
	return result
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

func (m *Model) uniqueServiceCount() int {
	seen := make(map[string]struct{})
	for _, p := range m.policies {
		seen[p.ServiceName] = struct{}{}
	}
	return len(seen)
}

// discoveryBar creates a small progress bar for inline use in the headline.
func (m *Model) discoveryBar() progress.Model {
	bar := progress.New(
		progress.WithColors(m.theme.GradientStart),
		progress.WithWidth(10),
		progress.WithFillCharacters('█', '░'),
	)
	bar.ShowPercentage = false
	bar.EmptyColor = m.theme.TextMuted
	return bar
}

// summaryDiscoveryPercent computes account-level classification coverage.
func summaryDiscoveryPercent(s domain.AccountSummary) int {
	if s.TotalServiceVolumePerHour == nil || *s.TotalServiceVolumePerHour <= 0 {
		return 100
	}
	pct := int(math.Round(s.TotalVolumePerHour / *s.TotalServiceVolumePerHour * 100))
	if pct > 100 {
		pct = 100
	}
	return pct
}

func ptrVal(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
