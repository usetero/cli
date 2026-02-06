// Package syncstatus renders sync connection status.
package syncstatus

import (
	"fmt"
	"image/color"
	"net/url"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
)

const pollInterval = 500 * time.Millisecond

// pollMsg triggers a sync status check.
type pollMsg struct{}

// Model renders sync connection status.
type Model struct {
	theme  *styles.Theme
	syncer powersync.Syncer
	host   string

	// Cached state for change detection
	lastState powersync.State
}

// New creates a new sync status model.
func New(theme *styles.Theme, syncer powersync.Syncer, endpoint string) *Model {
	host := endpoint
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		host = u.Host
	}

	return &Model{
		theme:  theme,
		syncer: syncer,
		host:   host,
	}
}

// Init starts polling sync status.
func (m *Model) Init() tea.Cmd {
	if m.syncer == nil {
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
		if m.syncer == nil {
			return nil
		}

		currentState := m.syncer.State()
		if m.stateChanged(currentState) {
			m.lastState = currentState

			cmds := []tea.Cmd{
				m.poll(),
				func() tea.Msg { return msgs.SyncStateChanged{State: currentState} },
			}

			if errState, ok := currentState.(*powersync.Error); ok {
				cmds = append(cmds, msgs.ErrorCmd("Sync error", errState.Err, true))
			}

			return tea.Batch(cmds...)
		}

		return m.poll()
	}

	return nil
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

// View renders the sync status: "● api.usetero.com" or "● syncing 45%"
func (m *Model) View() string {
	if m.syncer == nil {
		return ""
	}

	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch state := m.syncer.State().(type) {
	case *powersync.Disconnected:
		return ""

	case *powersync.Connecting:
		return dot(colors.Warning.Fg) + " " + muted.Render("connecting...")

	case *powersync.Syncing:
		if state.Progress != nil && state.Progress.Total > 0 {
			pct := state.Progress.Downloaded * 100 / state.Progress.Total
			return dot(colors.Warning.Fg) + " " + muted.Render(fmt.Sprintf("syncing %d%%", pct))
		}
		return dot(colors.Warning.Fg) + " " + muted.Render("syncing...")

	case *powersync.Ready:
		return dot(colors.Success.Fg) + " " + muted.Render(m.host)

	case *powersync.Reconnecting:
		if state.Degraded {
			return dot(colors.Error.Fg) + " " + muted.Render("reconnecting...")
		}
		return dot(colors.Warning.Fg) + " " + muted.Render("reconnecting...")

	case *powersync.Error:
		errStyle := lipgloss.NewStyle().Foreground(colors.Error.Fg)
		return errStyle.Render("○") + " " + errStyle.Render("error")

	default:
		return ""
	}
}

func dot(c color.Color) string {
	return lipgloss.NewStyle().Foreground(c).Render("●")
}
