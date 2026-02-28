package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
)

// commandForBootstrapMessage maps bootstrap messages to deterministic transitions.
func (m *Model) commandForBootstrapMessage(msg tea.Msg) (tea.Cmd, bool) {
	bootstrapMsg, ok := msg.(bootstrap.Message)
	if !ok {
		return nil, false
	}

	event, ok := bootstrap.EventFromMessage(bootstrapMsg)
	if !ok {
		return nil, false
	}

	transition := m.applyTransition(event, msg)
	return m.commandForTransition(event, transition), true
}

func (m *Model) logPreflightResolved(preflight bootstrap.PreflightResolved) {
	m.scope.Debug("preflight complete",
		slog.String("outcome", string(preflight.State.Outcome)),
		slog.Bool("has_valid_auth", preflight.State.HasValidAuth),
		slog.String("role", preflight.State.Role),
		slog.String("active_org_id", preflight.State.ActiveOrgID.String()),
		slog.String("default_account_id", preflight.State.DefaultAccountID.String()),
		slog.String("default_workspace_id", string(preflight.State.DefaultWorkspaceID)),
		slog.Bool("org_resolved", preflight.State.Org != nil),
		slog.Bool("account_resolved", preflight.State.Account != nil),
		slog.String("error", preflight.State.Error))
}
