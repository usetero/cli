package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func (m *Model) toCoreState() bootstrap.State {
	return bootstrap.State{
		User:      m.state.user,
		Org:       m.state.org,
		Account:   m.state.account,
		Workspace: m.state.workspace,
		DDSite:    m.state.ddSite,
		DDAPIKey:  m.state.ddAPIKey,
		DDAccount: m.state.ddAccount,
	}
}

func (m *Model) applyCoreState(state bootstrap.State) {
	m.state.user = state.User
	m.state.org = state.Org
	m.state.account = state.Account
	m.state.workspace = state.Workspace
	m.state.ddSite = state.DDSite
	m.state.ddAPIKey = state.DDAPIKey
	m.state.ddAccount = state.DDAccount
	m.syncServicesToState()
}

func (m *Model) syncServicesToState() {
	accountID := domain.AccountID("")
	if m.state.account != nil {
		accountID = m.state.account.ID
	}
	m.services = m.services.WithAccountID(accountID)
}

func (m *Model) handlePreflightResolved(msg msgs.PreflightResolved) TransitionOutcome {
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

	nextState, next := bootstrap.ApplyPreflight(
		bootstrap.State{
			Org:     m.state.org,
			Account: m.state.account,
		},
		bootstrap.PreflightResolved{
			Outcome:      bootstrap.PreflightOutcome(msg.State.Outcome),
			HasValidAuth: msg.State.HasValidAuth,
			Role:         msg.State.Role,
			Org:          msg.State.Org,
			Account:      msg.State.Account,
		},
	)
	nextGate := next
	m.applyCoreState(nextState)
	m.scope.Info("preflight decision",
		slog.String("outcome", string(msg.State.Outcome)),
		slog.String("next_gate", nextGate.String()))
	return advance(nextGate)
}

func (m *Model) handleAuthenticated(msg msgs.Authenticated) TransitionOutcome {
	m.scope.Info("user authenticated", "user_id", msg.User.ID)
	nextState, next := bootstrap.ApplyAuthenticated(m.toCoreState(), msg.User)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleRoleSelected(msg msgs.RoleSelected) TransitionOutcome {
	m.scope.Info("role selected", slog.String("role", msg.Role))
	nextState, next := bootstrap.ApplyRoleSelected(m.toCoreState(), msg.Role)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleOrgSelected(msg msgs.OrgSelected) TransitionOutcome {
	m.scope.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
	nextState, next := bootstrap.ApplyOrgSelected(m.toCoreState(), msg.Org)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleNoOrgs() TransitionOutcome {
	m.scope.Debug("no organizations found")
	nextState, next := bootstrap.ApplyNoOrgs(m.toCoreState())
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleOrgCreated(msg msgs.OrgCreated) TransitionOutcome {
	m.scope.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
	nextState, next := bootstrap.ApplyOrgCreated(m.toCoreState(), msg.Org)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleAccountSelected(msg msgs.AccountSelected) TransitionOutcome {
	m.scope.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
	nextState, next := bootstrap.ApplyAccountSelected(m.toCoreState(), msg.Org, msg.Account)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleNoAccounts(msg msgs.NoAccounts) TransitionOutcome {
	m.scope.Debug("no accounts found")
	nextState, next := bootstrap.ApplyNoAccounts(m.toCoreState(), msg.Org)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleAccountCreated(msg msgs.AccountCreated) TransitionOutcome {
	m.scope.Info("account created", slog.String("account_id", msg.Account.ID.String()))
	nextState, next := bootstrap.ApplyAccountCreated(m.toCoreState(), msg.Org, msg.Account)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleRuntimeReady(msg msgs.RuntimeReady) TransitionOutcome {
	m.scope.Info("runtime initialized", slog.String("account_id", msg.Account.ID.String()))
	nextState, next := bootstrap.ApplyRuntimeReady(m.toCoreState(), msg.Org, msg.Account)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogReady() TransitionOutcome {
	m.scope.Debug("datadog ready")
	nextState, next := bootstrap.ApplyDatadogReady(m.toCoreState())
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogNeeded() TransitionOutcome {
	m.scope.Debug("datadog setup needed")
	nextState, next := bootstrap.ApplyDatadogNeeded(m.toCoreState())
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogRegionSelected(msg msgs.DatadogRegionSelected) TransitionOutcome {
	m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
	nextState, next := bootstrap.ApplyDatadogRegionSelected(m.toCoreState(), msg.Site)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogAPIKeyEntered(msg msgs.DatadogAPIKeyEntered) TransitionOutcome {
	m.scope.Debug("datadog api key validated")
	nextState, next := bootstrap.ApplyDatadogAPIKeyEntered(m.toCoreState(), msg.APIKey)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogAccountCreated(msg msgs.DatadogAccountCreated) TransitionOutcome {
	m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
	nextState, next := bootstrap.ApplyDatadogAccountCreated(m.toCoreState(), msg.DatadogAccountID)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleDatadogDiscoveryComplete() TransitionOutcome {
	m.scope.Info("datadog discovery complete")
	nextState, next := bootstrap.ApplyDatadogDiscoveryComplete(m.toCoreState())
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleWorkspaceSelected(msg msgs.WorkspaceSelected) TransitionOutcome {
	m.scope.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
	nextState, next := bootstrap.ApplyWorkspaceSelected(m.toCoreState(), msg.Workspace)
	m.applyCoreState(nextState)
	return advance(next)
}

func (m *Model) handleSyncComplete() TransitionOutcome {
	completion, ok := bootstrap.CompleteOnboarding(m.toCoreState())
	if !ok {
		m.scope.Error("sync completed without required onboarding state",
			slog.Bool("has_user", m.state.user != nil),
			slog.Bool("has_org", m.state.org != nil),
			slog.Bool("has_account", m.state.account != nil),
			slog.Bool("has_workspace", m.state.workspace != nil),
		)
		return noop()
	}

	m.scope.Info("onboarding complete",
		slog.String("org_id", completion.Org.ID.String()),
		slog.String("account_id", completion.Account.ID.String()),
		slog.String("workspace_id", string(completion.Workspace.ID)),
	)
	return advanceWith("", func() tea.Msg {
		return msgs.OnboardingComplete{
			User:      completion.User,
			Org:       completion.Org,
			Account:   completion.Account,
			Workspace: completion.Workspace,
		}
	})
}
