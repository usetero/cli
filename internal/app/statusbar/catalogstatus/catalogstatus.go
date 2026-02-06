// Package catalogstatus renders the catalog pulse in the status bar.
package catalogstatus

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a catalog status check.
type pollMsg struct{}

// Model renders the catalog pulse: service count, policies, phase.
type Model struct {
	theme *styles.Theme
	db    sqlite.DB

	status    sqlite.CatalogStatus
	hasData   bool
	lastState string // change detection key
}

// New creates a new catalog status model.
// The db can be nil initially and set later via SetDB.
func New(theme *styles.Theme) *Model {
	return &Model{
		theme: theme,
	}
}

// SetDB sets the database and starts polling if not already started.
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

		status, err := m.db.DatadogAccountStatuses().GetCatalogStatus(ctx)
		if err != nil {
			return m.poll()
		}

		key := fmt.Sprintf("%d:%d:%d:%d:%s:%.0f",
			status.ServiceCount, status.WasteCount, status.SavedCount,
			status.BrokenCount, status.WorstStatus, status.PercentComplete)

		if key != m.lastState {
			m.status = status
			m.hasData = status.ServiceCount > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// View renders the catalog pulse segment.
func (m *Model) View() string {
	if !m.hasData {
		return ""
	}

	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	warn := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
	danger := lipgloss.NewStyle().Foreground(colors.Error.Fg)

	s := m.status

	// Service count is always shown
	svcs := muted.Render(fmt.Sprintf("%d svcs", s.ServiceCount))

	// Phase-aware second segment
	var phase string
	switch s.WorstStatus {
	case "DISABLED":
		phase = danger.Render("disabled")
	case "INACTIVE":
		phase = danger.Render("inactive")
	case "BROKEN":
		phase = danger.Render("broken")
	case "STALE":
		phase = warn.Render("stale")
	case "DISCOVERING":
		if s.PercentComplete > 0 {
			phase = muted.Render(fmt.Sprintf("discovering %.0f%%", s.PercentComplete))
		} else {
			phase = muted.Render("discovering")
		}
	case "ANALYZING":
		phase = muted.Render("analyzing")
	default:
		// READY — show policy count
		if s.WasteCount > 0 {
			phase = muted.Render(fmt.Sprintf("%d policies", s.WasteCount))
		}
	}

	// Alerts (broken services)
	var alerts string
	if s.BrokenCount > 0 {
		alerts = danger.Render(fmt.Sprintf("%d!", s.BrokenCount))
	}

	// Savings
	var savings string
	if s.SavedCount > 0 {
		savings = muted.Render(fmt.Sprintf("%d saved", s.SavedCount))
	}

	// Build: "14 svcs · 324 policies · 2!"
	result := svcs
	if phase != "" {
		result += muted.Render(" · ") + phase
	}
	if savings != "" {
		result += muted.Render(" · ") + savings
	}
	if alerts != "" {
		result += muted.Render(" · ") + alerts
	}

	return result
}
