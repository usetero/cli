package statusbar

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/sqlite"
)

// SetDB sets the database for status polling.
//
// syncStatus is fed alongside the drawer tabs even though it is no longer a
// drawer tab itself: its compact sync dot lives in the brand segment and its
// pending-upload count needs the runtime database.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	cmds := []tea.Cmd{m.fetchWorkspaceCount(db), m.syncStatus.SetDB(db)}
	for _, tab := range m.tabs {
		cmds = append(cmds, tab.SetDB(db))
	}
	return tea.Batch(cmds...)
}

// Init initializes child models.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.syncStatus.Init()}
	for _, tab := range m.tabs {
		cmds = append(cmds, tab.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.ingestStatusMessages(msg)

	cmds := []tea.Cmd{m.syncStatus.Update(msg)}
	for _, tab := range m.tabs {
		cmds = append(cmds, tab.Update(msg))
	}
	return tea.Batch(cmds...)
}

func (m *Model) ingestStatusMessages(msg tea.Msg) {
	switch msg := msg.(type) {
	case bootstrap.OrgSelected:
		m.org = msg.Org.Name
	case bootstrap.OrgCreated:
		m.org = msg.Org.Name
	case bootstrap.WorkspaceSelected:
		m.workspace = msg.Workspace.Name
	case workspaceCountLoadedMsg:
		if msg.err != nil {
			m.scope.Error("scan workspace count", "err", msg.err)
			break
		}
		m.workspaceCount = msg.count
	}
}

func (m *Model) fetchWorkspaceCount(db sqlite.DB) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := sqlite.WithTimeout(context.Background(), workspaceCountTimeout)
		defer cancel()

		var count int64
		row := db.QueryRow(ctx, "SELECT COUNT(*) FROM workspaces")
		if err := row.Scan(&count); err != nil {
			return workspaceCountLoadedMsg{err: err}
		}
		return workspaceCountLoadedMsg{count: count}
	}
}

// SetWidth sets the statusbar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetTitle sets the conversation title.
func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetContextPercent sets the context window usage percentage.
func (m *Model) SetContextPercent(percent int) {
	m.contextPercent = percent
}
