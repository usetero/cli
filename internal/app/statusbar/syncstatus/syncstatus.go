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

	"github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

const pollInterval = 500 * time.Millisecond

// pollMsg triggers a sync status check.
type pollMsg struct{}

// entityRow holds record and pending upload counts for a single entity.
type entityRow struct {
	name    string
	table   sqlite.Table
	records int64
	pending int64
}

// Model renders sync connection status.
type Model struct {
	theme  *styles.Theme
	syncer powersync.Syncer
	db     sqlite.DB
	host   string

	// Cached state for change detection
	lastState    powersync.State
	entities     []entityRow
	totalPending int64
	lastDataKey  string
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
		stateChanged := m.stateChanged(currentState)
		if stateChanged {
			m.lastState = currentState
		}

		m.pollCounts()

		cmds := []tea.Cmd{m.poll()}
		if stateChanged {
			cmds = append(cmds, func() tea.Msg { return msgs.SyncStateChanged{State: currentState} })

			if errState, ok := currentState.(*powersync.Error); ok {
				cmds = append(cmds, msgs.ErrorCmd("Sync error", errState.Err, true))
			}
		}

		return tea.Batch(cmds...)
	}

	return nil
}

// entityDefs defines which entities to show and in what order.
var entityDefs = []struct {
	name  string
	table sqlite.Table
}{
	{"Services", sqlite.TableServices},
	{"Log Events", sqlite.TableLogEvents},
	{"Policies", sqlite.TableLogEventPolicies},
	{"Conversations", sqlite.TableConversations},
	{"Messages", sqlite.TableMessages},
}

// pollCounts fetches record counts and pending uploads from the database.
func (m *Model) pollCounts() {
	if m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pending, _ := m.db.PendingUploadCounts(ctx)

	rows := make([]entityRow, len(entityDefs))
	var total int64
	var key string

	for i, def := range entityDefs {
		rows[i] = entityRow{name: def.name, table: def.table}

		switch def.table {
		case sqlite.TableServices:
			rows[i].records, _ = m.db.Services().Count(ctx)
		case sqlite.TableLogEvents:
			rows[i].records, _ = m.db.LogEvents().Count(ctx)
		case sqlite.TableLogEventPolicies:
			rows[i].records, _ = m.db.LogEventPolicies().Count(ctx)
		case sqlite.TableConversations:
			rows[i].records, _ = m.db.Conversations().Count(ctx)
		case sqlite.TableMessages:
			rows[i].records, _ = m.db.Messages().Count(ctx)
		}

		rows[i].pending = pending[def.table]
		total += rows[i].pending
		key += fmt.Sprintf("%s:%d:%d|", def.name, rows[i].records, rows[i].pending)
	}

	if key != m.lastDataKey {
		m.entities = rows
		m.totalPending = total
		m.lastDataKey = key
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

// CompactView renders the sync status for the statusbar: "● api.usetero.com" or "● syncing 45%"
func (m *Model) CompactView() string {
	if m.lastState == nil {
		return ""
	}

	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	switch state := m.lastState.(type) {
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

// ExpandedView renders the detailed sync status for the drawer.
func (m *Model) ExpandedView(width int) string {
	if m.lastState == nil {
		return ""
	}

	var lines []string
	lines = append(lines, m.renderStatusLine())
	lines = append(lines, "")
	lines = append(lines, m.renderEntityTable(width))
	return strings.Join(lines, "\n")
}

// renderStatusLine renders sync and upload status on one line.
func (m *Model) renderStatusLine() string {
	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	var syncPart string
	switch state := m.lastState.(type) {
	case *powersync.Disconnected:
		syncPart = dot(colors.Page.TextSubtle) + " " + muted.Render("Disconnected")
	case *powersync.Connecting:
		syncPart = dot(colors.Warning.Fg) + " " + muted.Render("Connecting")
	case *powersync.Syncing:
		if state.Progress != nil && state.Progress.Total > 0 {
			pct := state.Progress.Downloaded * 100 / state.Progress.Total
			syncPart = dot(colors.Warning.Fg) + " " + muted.Render(fmt.Sprintf("Syncing %d%%", pct))
		} else {
			syncPart = dot(colors.Warning.Fg) + " " + muted.Render("Syncing")
		}
	case *powersync.Ready:
		syncPart = dot(colors.Success.Fg) + " " + muted.Render("Connected")
	case *powersync.Reconnecting:
		if state.Degraded {
			syncPart = dot(colors.Error.Fg) + " " + muted.Render("Reconnecting (degraded)")
		} else {
			syncPart = dot(colors.Warning.Fg) + " " + muted.Render("Reconnecting")
		}
	case *powersync.Error:
		errStyle := lipgloss.NewStyle().Foreground(colors.Error.Fg)
		syncPart = errStyle.Render("○") + " " + errStyle.Render(state.Err.Error())
	}

	var uploadPart string
	if m.totalPending > 0 {
		warn := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
		uploadPart = warn.Render(fmt.Sprintf("%d uploading", m.totalPending))
	} else {
		uploadPart = muted.Render("Upload queue empty")
	}

	return syncPart + muted.Render(" · ") + uploadPart
}

// renderEntityTable renders record counts with optional pending column.
func (m *Model) renderEntityTable(width int) string {
	if len(m.entities) == 0 {
		return ""
	}

	tbl := table.New(m.theme)
	tbl.SetWidth(width)

	if m.totalPending > 0 {
		tbl.Headers("Entity", "Records", "Pending")
		for _, e := range m.entities {
			p := "-"
			if e.pending > 0 {
				p = fmt.Sprintf("%d", e.pending)
			}
			tbl.Row(e.name, fmt.Sprintf("%d", e.records), p)
		}
	} else {
		tbl.Headers("Entity", "Records")
		for _, e := range m.entities {
			tbl.Row(e.name, fmt.Sprintf("%d", e.records))
		}
	}

	return tbl.View()
}

func dot(c color.Color) string {
	return lipgloss.NewStyle().Foreground(c).Render("●")
}
