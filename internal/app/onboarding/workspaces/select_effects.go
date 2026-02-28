package workspaces

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

func (m *SelectModel) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		workspaces, err := m.services.Workspaces.List(m.ctx, m.account.ID)
		if err != nil {
			m.scope.Error("failed to load workspaces", slog.Any("error", err))
			return remotelist.LoadResult{Err: err}
		}

		m.scope.Debug("loaded workspaces", slog.Int("count", len(workspaces)))
		items := make([]remotelist.Item, len(workspaces))
		for i, ws := range workspaces {
			items[i] = ws // domain.Workspace implements FilterValue()
		}
		return remotelist.LoadResult{Items: items}
	}
}

func (m *SelectModel) emitSelected(ws domain.Workspace) tea.Cmd {
	return func() tea.Msg {
		return msgs.WorkspaceSelected{Workspace: ws}
	}
}
