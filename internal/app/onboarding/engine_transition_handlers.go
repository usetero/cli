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

func (m *Model) handleOrgAccountRuntimeTransition(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.OrgSelected:
		m.scope.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
		m.state.org = &msg.Org
		m.services = m.services.WithAccountID("")
		return advance(GateAccountSelect), true

	case msgs.NoOrgs:
		m.scope.Debug("no organizations found")
		return advance(GateOrgCreate), true

	case msgs.OrgCreated:
		m.scope.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
		m.state.org = &msg.Org
		m.services = m.services.WithAccountID("")
		return advance(GateAccountSelect), true

	case msgs.AccountSelected:
		m.scope.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
		m.state.org = &msg.Org
		m.state.account = &msg.Account
		return advance(GateRuntimeInit), true

	case msgs.NoAccounts:
		m.scope.Debug("no accounts found")
		m.state.org = &msg.Org
		return advance(GateAccountCreate), true

	case msgs.AccountCreated:
		m.scope.Info("account created", slog.String("account_id", msg.Account.ID.String()))
		m.state.org = &msg.Org
		m.state.account = &msg.Account
		return advance(GateRuntimeInit), true

	case msgs.RuntimeReady:
		m.scope.Info("runtime initialized", slog.String("account_id", msg.Account.ID.String()))
		m.state.org = &msg.Org
		m.state.account = &msg.Account
		m.services = m.services.WithAccountID(msg.Account.ID)
		return advance(GateDatadogCheck), true
	default:
		return noop(), false
	}
}

func (m *Model) handleDatadogTransition(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.DatadogReady:
		m.scope.Debug("datadog ready")
		return advance(GateWorkspaceSelect), true
	case msgs.DatadogNeeded:
		m.scope.Debug("datadog setup needed")
		return advance(GateDatadogRegion), true
	case msgs.DatadogRegionSelected:
		m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
		m.state.ddSite = msg.Site
		return advance(GateDatadogAPIKey), true
	case msgs.DatadogAPIKeyEntered:
		m.scope.Debug("datadog api key validated")
		m.state.ddAPIKey = msg.APIKey
		return advance(GateDatadogAppKey), true
	case msgs.DatadogAccountCreated:
		m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
		m.state.ddAccount = msg.DatadogAccountID
		return advance(GateDatadogDiscovery), true
	case msgs.DatadogDiscoveryComplete:
		m.scope.Info("datadog discovery complete")
		return advance(GateWorkspaceSelect), true
	default:
		return noop(), false
	}
}

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
