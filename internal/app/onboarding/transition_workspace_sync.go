package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

func (m *Model) handleWorkspaceSyncTransition(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.WorkspaceSelected:
		m.scope.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
		m.state.workspace = &msg.Workspace
		return advance(GateSync), true

	case msgs.SyncComplete:
		m.scope.Info("onboarding complete",
			slog.String("org_id", m.state.org.ID.String()),
			slog.String("account_id", m.state.account.ID.String()),
			slog.String("workspace_id", string(m.state.workspace.ID)),
		)
		return advanceWith("", func() tea.Msg {
			return msgs.OnboardingComplete{
				User:      m.state.user,
				Org:       *m.state.org,
				Account:   *m.state.account,
				Workspace: *m.state.workspace,
			}
		}), true

	default:
		return noop(), false
	}
}
