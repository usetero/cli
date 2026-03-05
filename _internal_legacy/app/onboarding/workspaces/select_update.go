package workspaces

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		return m.handleLoadResult(msg)

	case tea.KeyPressMsg:
		if m.list.IsLoading() {
			return nil
		}
		switch msg.String() {
		case "enter":
			if item := m.list.SelectedItem(); item != nil {
				if ws, ok := item.(domain.Workspace); ok {
					_ = m.prefs.SetDefaultWorkspaceID(ws.ID)
					m.scope.Info("workspace selected", slog.String("workspace_id", string(ws.ID)))
					return m.emitSelected(ws)
				}
			}
		case "r":
			if m.list.HasError() {
				m.scope.Debug("retrying workspace load")
				return m.list.Retry()
			}
		}
	}

	return m.list.Update(msg)
}

func (m *SelectModel) handleLoadResult(msg remotelist.LoadResult) tea.Cmd {
	if msg.Err != nil {
		return tea.Batch(m.list.Update(msg), appevents.PublishErrorToastCmd("Failed to load workspaces", msg.Err, false))
	}
	m.workspaces = stepkit.CastItems[domain.Workspace](msg.Items)

	if prefWS := findWorkspaceByID(m.workspaces, m.prefs.GetDefaultWorkspaceID()); prefWS != nil {
		m.scope.Info("workspace restored from preference", slog.String("workspace_id", string(prefWS.ID)))
		return m.emitSelected(*prefWS)
	}

	if len(m.workspaces) == 1 {
		ws := m.workspaces[0]
		_ = m.prefs.SetDefaultWorkspaceID(ws.ID)
		m.scope.Info("workspace auto-selected", slog.String("workspace_id", string(ws.ID)))
		return m.emitSelected(ws)
	}

	return m.list.Update(msg)
}
