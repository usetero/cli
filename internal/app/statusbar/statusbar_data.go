package statusbar

import (
	"context"

	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/sqlite"
)

// SetServices points the drawer tabs at the account-scoped control-plane
// services. Each tab polls its own GraphQL reads from here.
func (m *Model) SetServices(services graphql.ServiceSet) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.tabs))
	for _, tab := range m.tabs {
		cmds = append(cmds, tab.SetServices(services))
	}
	return tea.Batch(cmds...)
}

// SetDB feeds the sync indicator and workspace count, which still read the
// local runtime database. The drawer tabs are fed via SetServices.
//
// syncStatus is fed alongside the drawer tabs even though it is no longer a
// drawer tab itself: its compact sync dot lives in the brand segment and its
// pending-upload count needs the runtime database.
func (m *Model) SetDB(db sqlite.DB) tea.Cmd {
	return tea.Batch(m.fetchWorkspaceCount(db), m.syncStatus.SetDB(db))
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
