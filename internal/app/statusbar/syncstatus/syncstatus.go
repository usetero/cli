// Package syncstatus renders sync connection status.
package syncstatus

import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const (
	pollInterval        = 500 * time.Millisecond
	pendingPollInterval = 2 * time.Second
	dbTimeout           = 2 * time.Second
)

// pollMsg triggers a sync status check.
type pollMsg struct{}

// pendingPollMsg triggers a pending-upload count refresh.
type pendingPollMsg struct{}

// pendingMsg carries the result of an async pending-upload count.
type pendingMsg struct {
	total int64
}

// Model renders sync connection status.
type Model struct {
	theme  styles.Theme
	scope  log.Scope
	syncer powersync.Syncer
	db     sqlite.DB
	host   string

	// Cached state for change detection
	lastState    powersync.State
	totalPending int64
	pendingFetch bool
}

// New creates a new sync status model.
func New(theme styles.Theme, scope log.Scope, syncer powersync.Syncer, endpoint string) *Model {
	host := endpoint
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		host = u.Host
	}

	return &Model{
		theme:  theme,
		scope:  scope.Child("syncstatus"),
		syncer: syncer,
		host:   host,
	}
}

// SetDB sets the database for record count polling.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	m.db = db
	return nil
}

// Init starts polling sync status.
func (m *Model) Init() tea.Cmd {
	if m.syncer == nil {
		return nil
	}
	return tea.Batch(m.poll(), m.pollPending())
}

func (m *Model) poll() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollMsg{}
	})
}

func (m *Model) pollPending() tea.Cmd {
	return tea.Tick(pendingPollInterval, func(time.Time) tea.Msg {
		return pendingPollMsg{}
	})
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case pollMsg:
		if m.syncer == nil {
			return nil
		}

		// syncer.State() is an atomic load — safe to call inline.
		currentState := m.syncer.State()
		stateChanged := m.stateChanged(currentState)
		if stateChanged {
			m.lastState = currentState
		}

		cmds := []tea.Cmd{m.poll()}
		if stateChanged {
			cmds = append(cmds, func() tea.Msg { return appevents.SyncStateChanged{State: currentState} })

			if errState, ok := currentState.(*powersync.Error); ok {
				cmds = append(cmds, appevents.ErrorCmd("Sync error", errState.Err, true))
			}
		}

		return tea.Batch(cmds...)

	case pendingPollMsg:
		if m.db == nil {
			return m.pollPending()
		}
		if m.pendingFetch {
			return m.pollPending()
		}
		m.pendingFetch = true
		return tea.Batch(m.pollPending(), m.fetchPending())

	case pendingMsg:
		m.pendingFetch = false
		m.totalPending = msg.total
	}

	return nil
}

// fetchPending returns a Cmd that queries pending upload counts off the event loop.
func (m *Model) fetchPending() tea.Cmd {
	if m.db == nil {
		return nil
	}
	db := m.db
	scope := m.scope
	return func() tea.Msg {
		ctx, cancel := sqlite.WithTimeout(context.Background(), dbTimeout)
		defer cancel()
		pending, err := db.PendingUploadCounts(ctx)
		if err != nil {
			scope.Error("pending upload counts", "err", err)
			return pendingMsg{}
		}

		var total int64
		for _, count := range pending {
			total += count
		}
		return pendingMsg{total: total}
	}
}

// stateChanged returns true if the sync state has meaningfully changed.
func (m *Model) stateChanged(current powersync.State) bool {
	if m.lastState == nil {
		return current != nil
	}
	if current == nil {
		return true
	}

	switch last := m.lastState.(type) {
	case *powersync.Disconnected:
		_, ok := current.(*powersync.Disconnected)
		return !ok
	case *powersync.Connecting:
		_, ok := current.(*powersync.Connecting)
		return !ok
	case *powersync.Syncing:
		_, ok := current.(*powersync.Syncing)
		return !ok
	case *powersync.Ready:
		_, ok := current.(*powersync.Ready)
		return !ok
	case *powersync.Reconnecting:
		cur, ok := current.(*powersync.Reconnecting)
		return !ok || cur.Degraded != last.Degraded
	case *powersync.Error:
		_, ok := current.(*powersync.Error)
		return !ok
	}
	return true
}

// HasData returns true when the syncer has reported at least one state update.
func (m *Model) HasData() bool {
	return m.lastState != nil
}

// CompactView renders the sync status for the statusbar: "● api.usetero.com" or "● syncing 45%"
func (m *Model) CompactView() string {
	if m.lastState == nil {
		return ""
	}

	colors := m.theme

	switch state := m.lastState.(type) {
	case *powersync.Disconnected:
		return ""

	case *powersync.Connecting:
		return dot(colors.Warning, colors.Bg)

	case *powersync.Syncing:
		return dot(colors.Warning, colors.Bg)

	case *powersync.Ready:
		return dot(colors.Success, colors.Bg)

	case *powersync.Reconnecting:
		if state.Degraded {
			return dot(colors.Error, colors.Bg)
		}
		return dot(colors.Warning, colors.Bg)

	case *powersync.Error:
		return dot(colors.Error, colors.Bg)

	default:
		return ""
	}
}

// ExpandedView renders the detailed sync status for the drawer.
func (m *Model) ExpandedView(width, _ int) string {
	if m.lastState == nil {
		return ""
	}

	colors := m.theme
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	var headline string
	var description string

	switch state := m.lastState.(type) {
	case *powersync.Disconnected:
		headline = dot(colors.TextSubtle, colors.Bg) + " " + text.Render("Disconnected")
		description = "Sync has not started yet."

	case *powersync.Connecting:
		headline = dot(colors.Warning, colors.Bg) + " " + text.Render("Connecting...")
		description = "Establishing connection to the control plane."

	case *powersync.Syncing:
		if state.Progress != nil && state.Progress.Total > 0 {
			pct := state.Progress.Downloaded * 100 / state.Progress.Total
			headline = dot(colors.Warning, colors.Bg) + " " + text.Render(fmt.Sprintf("Syncing your data... %d%%", pct))
			description = fmt.Sprintf("%d / %d rows downloaded.", state.Progress.Downloaded, state.Progress.Total)
		} else {
			headline = dot(colors.Warning, colors.Bg) + " " + text.Render("Syncing your data...")
			description = "Downloading from the control plane."
		}

	case *powersync.Ready:
		headline = dot(colors.Success, colors.Bg) + " " + text.Render("Connected")
		description = "Your data is synced and up to date."

	case *powersync.Reconnecting:
		if state.Degraded {
			headline = dot(colors.Error, colors.Bg) + " " + text.Render("Connection issues")
			description = "Multiple retries failed. Still trying to reconnect."
		} else {
			headline = dot(colors.Warning, colors.Bg) + " " + text.Render("Reconnecting...")
			description = "Temporarily lost connection. Retrying automatically."
		}

	case *powersync.Error:
		errStyle := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg)
		headline = dot(colors.Error, colors.Bg) + " " + errStyle.Render("Sync failed")
		description = state.Err.Error()
	}

	var lines []string
	lines = append(lines, headline)
	lines = append(lines, "")
	lines = append(lines, muted.Render(description))

	if m.totalPending > 0 {
		lines = append(lines, "")
		warn := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		lines = append(lines, warn.Render(fmt.Sprintf("%d pending uploads", m.totalPending)))
	}

	return strings.Join(lines, "\n")
}

func dot(c color.Color, bg color.Color) string {
	return lipgloss.NewStyle().Foreground(c).Background(bg).Render("●")
}
