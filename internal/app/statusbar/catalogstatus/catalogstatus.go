// Package catalogstatus renders the catalog health indicator in the status bar.
package catalogstatus

import (
	"context"
	"fmt"
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const pollInterval = 2 * time.Second

// pollMsg triggers a catalog status check.
type pollMsg struct{}

// Model renders the catalog health: dot color + service count or discovery phase.
type Model struct {
	theme *styles.Theme
	db    sqlite.DB

	status    sqlite.CatalogStatus
	hasData   bool
	lastState string
}

// New creates a new catalog status model.
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

		status, err := m.db.DatadogAccountStatuses().GetCatalogStatus(ctx)
		if err != nil {
			return m.poll()
		}

		key := fmt.Sprintf("%v:%d:%d:%d:%d:%d:%d:%d:%s:%.0f:%s",
			status.ReadyForUse, status.ServiceCount, status.ActiveServices,
			status.EventCount, status.AnalyzedCount, status.AnalyzingCount,
			status.DiscoveringCount, status.BrokenServices,
			status.WorstStatus, status.PercentComplete, status.LogError)

		if key != m.lastState {
			m.status = status
			m.hasData = status.ServiceCount > 0
			m.lastState = key
		}

		return m.poll()
	}

	return nil
}

// CompactView renders the catalog indicator for the statusbar.
func (m *Model) CompactView() string {
	if !m.hasData {
		return ""
	}

	s := m.status
	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	errStyle := lipgloss.NewStyle().Foreground(colors.Error.Fg)

	dotColor := m.dotColor()
	d := dot(dotColor)

	if s.ReadyForUse {
		return d + " " + muted.Render(fmt.Sprintf("%d svcs", s.ServiceCount))
	}

	// Pre-ready: show discovery/analysis phase.
	switch s.WorstStatus {
	case "DISABLED":
		return d + " " + errStyle.Render("disabled")
	case "INACTIVE":
		return d + " " + errStyle.Render("inactive")
	case "BROKEN":
		return d + " " + errStyle.Render("error")
	case "STALE":
		return d + " " + lipgloss.NewStyle().Foreground(colors.Warning.Fg).Render("stale")
	case "DISCOVERING":
		if s.PercentComplete > 0 {
			return d + " " + muted.Render(fmt.Sprintf("discovering %.0f%%", s.PercentComplete))
		}
		return d + " " + muted.Render("discovering")
	case "ANALYZING":
		if s.EventCount > 0 {
			return d + " " + muted.Render(fmt.Sprintf("analyzing · %d events", s.EventCount))
		}
		return d + " " + muted.Render("analyzing")
	default:
		return d + " " + muted.Render(fmt.Sprintf("%d svcs", s.ServiceCount))
	}
}

// ExpandedView renders the detailed catalog status for the drawer.
func (m *Model) ExpandedView() string {
	return m.CompactView()
}

func (m *Model) dotColor() color.Color {
	colors := m.theme.Colors
	switch m.status.WorstStatus {
	case "BROKEN", "DISABLED", "INACTIVE":
		return colors.Error.Fg
	case "DISCOVERING", "ANALYZING", "STALE":
		return colors.Warning.Fg
	default:
		return colors.Success.Fg
	}
}

func dot(c color.Color) string {
	return lipgloss.NewStyle().Foreground(c).Render("●")
}
