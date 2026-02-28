package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

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
	return advance(nextGate)
}

func (m *Model) handleAuthenticated(msg msgs.Authenticated) TransitionOutcome {
	m.scope.Info("user authenticated", "user_id", msg.User.ID)
	m.state.user = &msg.User
	return advance(GateRoleSelect)
}

func (m *Model) handleRoleSelected(msg msgs.RoleSelected) TransitionOutcome {
	m.scope.Info("role selected", slog.String("role", msg.Role))
	return advance(GateOrgSelect)
}

func (m *Model) handleOrgSelected(msg msgs.OrgSelected) TransitionOutcome {
	m.scope.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
	m.state.org = &msg.Org
	m.services = m.services.WithAccountID("")
	return advance(GateAccountSelect)
}

func (m *Model) handleNoOrgs() TransitionOutcome {
	m.scope.Debug("no organizations found")
	return advance(GateOrgCreate)
}

func (m *Model) handleOrgCreated(msg msgs.OrgCreated) TransitionOutcome {
	m.scope.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
	m.state.org = &msg.Org
	m.services = m.services.WithAccountID("")
	return advance(GateAccountSelect)
}

func (m *Model) handleAccountSelected(msg msgs.AccountSelected) TransitionOutcome {
	m.scope.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
	m.state.org = &msg.Org
	m.state.account = &msg.Account
	return advance(GateRuntimeInit)
}

func (m *Model) handleNoAccounts(msg msgs.NoAccounts) TransitionOutcome {
	m.scope.Debug("no accounts found")
	m.state.org = &msg.Org
	return advance(GateAccountCreate)
}

func (m *Model) handleAccountCreated(msg msgs.AccountCreated) TransitionOutcome {
	m.scope.Info("account created", slog.String("account_id", msg.Account.ID.String()))
	m.state.org = &msg.Org
	m.state.account = &msg.Account
	return advance(GateRuntimeInit)
}

func (m *Model) handleRuntimeReady(msg msgs.RuntimeReady) TransitionOutcome {
	m.scope.Info("runtime initialized", slog.String("account_id", msg.Account.ID.String()))
	m.state.org = &msg.Org
	m.state.account = &msg.Account
	m.services = m.services.WithAccountID(msg.Account.ID)
	return advance(GateDatadogCheck)
}

func (m *Model) handleDatadogReady() TransitionOutcome {
	m.scope.Debug("datadog ready")
	return advance(GateWorkspaceSelect)
}

func (m *Model) handleDatadogNeeded() TransitionOutcome {
	m.scope.Debug("datadog setup needed")
	return advance(GateDatadogRegion)
}

func (m *Model) handleDatadogRegionSelected(msg msgs.DatadogRegionSelected) TransitionOutcome {
	m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
	m.state.ddSite = msg.Site
	return advance(GateDatadogAPIKey)
}

func (m *Model) handleDatadogAPIKeyEntered(msg msgs.DatadogAPIKeyEntered) TransitionOutcome {
	m.scope.Debug("datadog api key validated")
	m.state.ddAPIKey = msg.APIKey
	return advance(GateDatadogAppKey)
}

func (m *Model) handleDatadogAccountCreated(msg msgs.DatadogAccountCreated) TransitionOutcome {
	m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
	m.state.ddAccount = msg.DatadogAccountID
	return advance(GateDatadogDiscovery)
}

func (m *Model) handleDatadogDiscoveryComplete() TransitionOutcome {
	m.scope.Info("datadog discovery complete")
	return advance(GateWorkspaceSelect)
}

func (m *Model) handleWorkspaceSelected(msg msgs.WorkspaceSelected) TransitionOutcome {
	m.scope.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
	m.state.workspace = &msg.Workspace
	return advance(GateSync)
}

func (m *Model) handleSyncComplete() TransitionOutcome {
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
	})
}
