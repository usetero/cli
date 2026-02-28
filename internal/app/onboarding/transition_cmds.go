package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) commandForTransition(event bootstrap.Event, transition bootstrap.Transition) tea.Cmd {
	switch transition.Kind {
	case bootstrap.TransitionAdvance:
		nav := m.goToGate(transition.Next)
		if event.Kind == bootstrap.EventPreflightResolved {
			m.scope.Info("preflight decision",
				slog.String("outcome", string(event.Preflight.Outcome)),
				slog.String("next_gate", transition.Next.String()))
		}
		return nav
	case bootstrap.TransitionComplete:
		m.scope.Info("onboarding complete",
			slog.String("org_id", transition.Completion.Org.ID.String()),
			slog.String("account_id", transition.Completion.Account.ID.String()),
			slog.String("workspace_id", string(transition.Completion.Workspace.ID)),
		)
		return func() tea.Msg {
			return bootstrap.OnboardingComplete{
				User:      transition.Completion.User,
				Org:       transition.Completion.Org,
				Account:   transition.Completion.Account,
				Workspace: transition.Completion.Workspace,
			}
		}
	case bootstrap.TransitionNoop:
		if event.Kind == bootstrap.EventSyncComplete {
			m.scope.Error("sync completed without required onboarding state",
				slog.Bool("has_user", m.state.User != nil),
				slog.Bool("has_org", m.state.Org != nil),
				slog.Bool("has_account", m.state.Account != nil),
				slog.Bool("has_workspace", m.state.Workspace != nil),
			)
		}
		return nil
	default:
		return nil
	}
}
