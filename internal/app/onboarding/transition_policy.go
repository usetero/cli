package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	onboardingmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// transitionCmdFor maps onboarding messages to deterministic state transitions.
func (m *Model) transitionCmdFor(msg tea.Msg) (tea.Cmd, bool) {
	event, ok := bootstrapEventFor(msg)
	if !ok {
		return nil, false
	}

	if preflight, ok := msg.(onboardingmsg.PreflightResolved); ok {
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

	transition := bootstrap.ApplyEvent(m.state, event)
	m.state = transition.State
	m.syncServicesToState()

	switch transition.Kind {
	case bootstrap.TransitionAdvance:
		nav := m.goToGate(transition.Next)
		if event.Kind == bootstrap.EventPreflightResolved {
			m.scope.Info("preflight decision",
				slog.String("outcome", string(event.Preflight.Outcome)),
				slog.String("next_gate", transition.Next.String()))
		}
		return nav, true
	case bootstrap.TransitionComplete:
		m.scope.Info("onboarding complete",
			slog.String("org_id", transition.Completion.Org.ID.String()),
			slog.String("account_id", transition.Completion.Account.ID.String()),
			slog.String("workspace_id", string(transition.Completion.Workspace.ID)),
		)
		return func() tea.Msg {
			return onboardingmsg.OnboardingComplete{
				User:      transition.Completion.User,
				Org:       transition.Completion.Org,
				Account:   transition.Completion.Account,
				Workspace: transition.Completion.Workspace,
			}
		}, true
	case bootstrap.TransitionNoop:
		if event.Kind == bootstrap.EventSyncComplete {
			m.scope.Error("sync completed without required onboarding state",
				slog.Bool("has_user", m.state.User != nil),
				slog.Bool("has_org", m.state.Org != nil),
				slog.Bool("has_account", m.state.Account != nil),
				slog.Bool("has_workspace", m.state.Workspace != nil),
			)
		}
		return nil, true
	default:
		return nil, true
	}
}
