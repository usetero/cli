package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

func (m *Model) handlePreflightAndAuthTransition(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.PreflightResolved:
		m.scope.Debug("preflight complete",
			slog.String("outcome", string(msg.State.Outcome)),
			slog.Bool("has_valid_auth", msg.State.HasValidAuth),
			slog.String("role", msg.State.Role),
			slog.String("active_org_id", msg.State.ActiveOrgID.String()),
			slog.String("default_account_id", msg.State.DefaultAccountID.String()),
			slog.String("default_workspace_id", string(msg.State.DefaultWorkspaceID)),
			slog.Bool("org_resolved", msg.State.Org != nil),
			slog.Bool("account_resolved", msg.State.Account != nil),
			slog.String("error", msg.State.Error))

		nextGate := decidePreflightGate(msg)
		if msg.State.Org != nil {
			m.state.org = msg.State.Org
			m.services = m.services.WithAccountID("")
		}
		if msg.State.Account != nil {
			m.state.account = msg.State.Account
			m.services = m.services.WithAccountID(msg.State.Account.ID)
		}
		m.scope.Info("preflight decision",
			slog.String("outcome", string(msg.State.Outcome)),
			slog.String("next_gate", nextGate.String()))
		return advance(nextGate), true

	case msgs.Authenticated:
		m.scope.Info("user authenticated", "user_id", msg.User.ID)
		m.state.user = &msg.User
		return advance(GateRoleSelect), true

	case msgs.RoleSelected:
		m.scope.Info("role selected", slog.String("role", msg.Role))
		return advance(GateOrgSelect), true

	default:
		return noop(), false
	}
}

func decidePreflightGate(msg msgs.PreflightResolved) Gate {
	if msg.State.Outcome == msgs.PreflightOutcomeFailed || !msg.State.HasValidAuth {
		return GateAuthenticate
	}
	if msg.State.Role != msgs.RolePlatform && msg.State.Role != msgs.RoleEngineer {
		return GateRoleSelect
	}
	if msg.State.Org == nil {
		return GateOrgSelect
	}
	if msg.State.Account == nil {
		return GateAccountSelect
	}
	return GateRuntimeInit
}
