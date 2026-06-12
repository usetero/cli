package statusbar

import (
	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/core/bootstrap"
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

// Init initializes child models.
func (m *Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.tabs))
	for _, tab := range m.tabs {
		cmds = append(cmds, tab.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.ingestStatusMessages(msg)

	cmds := make([]tea.Cmd, 0, len(m.tabs))
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
