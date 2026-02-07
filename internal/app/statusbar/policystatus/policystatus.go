// Package policystatus renders the policy work indicator in the status bar.
package policystatus

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a policy status check.
type pollMsg struct{}

// Model renders the policy status: pending count, estimated savings, observed savings.
type Model struct {
	theme *styles.Theme
	db    sqlite.DB

	status    sqlite.PolicyStatus
	hasData   bool
	lastState string
}

// New creates a new policy status model.
func New(theme *styles.Theme) *Model {
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

		status, err := m.db.DatadogAccountStatuses().GetPolicyStatus(ctx)
		if err != nil {
			return m.poll()
		}

		key := fmt.Sprintf("%v:%d:%d:%d:%v:%.0f:%.0f:%v:%v:%.0f:%.0f",
			status.ReadyForUse, status.PendingPolicyCount,
			status.PolicyCount, status.ApprovedPolicyCount,
			ptrVal(status.EstimatedCostPerHour),
			status.EstimatedVolumePerHour, status.EstimatedBytesPerHour,
			ptrVal(status.ObservedCostBefore), ptrVal(status.ObservedCostAfter),
			status.ObservedVolumeBefore, status.ObservedVolumeAfter)

		if key != m.lastState {
			m.status = status
			m.hasData = status.ReadyForUse
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// CompactView renders the policy status for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	s := m.status
	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	sep := muted.Render(" · ")

	var segments []string

	// Pending policies + estimated savings.
	if s.PendingPolicyCount > 0 {
		pending := muted.Render(fmt.Sprintf("%d pending", s.PendingPolicyCount))
		if est := formatEstimated(s); est != "" {
			segments = append(segments, pending+sep+muted.Render("~"+est))
		} else {
			segments = append(segments, pending)
		}
	}

	// Observed savings from approved policies.
	if saving := formatObservedSaving(s); saving != "" {
		savingStyle := lipgloss.NewStyle().Foreground(colors.Success.Fg)
		segments = append(segments, savingStyle.Render("saving "+saving))
	}

	if len(segments) == 0 {
		return muted.Render("healthy")
	}

	return strings.Join(segments, sep)
}

// ExpandedView renders the detailed policy status for the drawer.
func (m *Model) ExpandedView() string {
	return m.CompactView()
}

// formatEstimated returns the estimated impact of pending policies.
// Prefers cost when available, falls back to volume.
func formatEstimated(s sqlite.PolicyStatus) string {
	if s.EstimatedCostPerHour != nil {
		monthly := *s.EstimatedCostPerHour * 730
		if monthly > 0 {
			return formatCost(monthly) + "/mo"
		}
	}
	if s.EstimatedVolumePerHour > 0 {
		return formatVolume(s.EstimatedVolumePerHour) + " evt/hr"
	}
	return ""
}

// formatObservedSaving returns the observed savings from approved policies.
// Prefers cost when available, falls back to volume.
func formatObservedSaving(s sqlite.PolicyStatus) string {
	if s.ObservedCostBefore != nil && s.ObservedCostAfter != nil {
		diff := (*s.ObservedCostBefore - *s.ObservedCostAfter) * 730
		if diff > 0 {
			return formatCost(diff) + "/mo"
		}
		return ""
	}
	diff := s.ObservedVolumeBefore - s.ObservedVolumeAfter
	if diff > 0 {
		return formatVolume(diff) + " evt/hr"
	}
	return ""
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

// formatVolume formats events/hr: 892, 45.3k, 2.1M.
func formatVolume(eventsPerHour float64) string {
	abs := math.Abs(eventsPerHour)
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", eventsPerHour/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", eventsPerHour/1_000)
	default:
		return fmt.Sprintf("%.0f", eventsPerHour)
	}
}

func ptrVal(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
