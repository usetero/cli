package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

func (m *Model) handleTransition(msg tea.Msg) tea.Cmd {
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

		nextGate := GateDatadogCheck
		if msg.State.Outcome == msgs.PreflightOutcomeFailed || !msg.State.HasValidAuth {
			nextGate = GateAuthenticate
		} else if msg.State.Role != msgs.RolePlatform && msg.State.Role != msgs.RoleEngineer {
			nextGate = GateRoleSelect
		} else if msg.State.Org == nil {
			nextGate = GateOrgSelect
		} else if msg.State.Account == nil {
			nextGate = GateAccountSelect
		}

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
		return m.goToGate(nextGate)

	case msgs.Authenticated:
		m.scope.Info("user authenticated", "user_id", msg.User.ID)
		m.state.user = &msg.User
		return m.goToGate(GateRoleSelect)

	case msgs.RoleSelected:
		m.scope.Info("role selected", slog.String("role", msg.Role))
		return m.goToGate(GateOrgSelect)

	case msgs.OrgSelected:
		m.scope.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
		m.state.org = &msg.Org
		m.services = m.services.WithAccountID("")
		return m.goToGate(GateAccountSelect)

	case msgs.NoOrgs:
		m.scope.Debug("no organizations found")
		return m.goToGate(GateOrgCreate)

	case msgs.OrgCreated:
		m.scope.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
		m.state.org = &msg.Org
		m.services = m.services.WithAccountID("")
		return m.goToGate(GateAccountSelect)

	case msgs.AccountSelected:
		m.scope.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
		m.state.org = &msg.Org
		m.state.account = &msg.Account
		m.services = m.services.WithAccountID(msg.Account.ID)
		return m.goToGate(GateDatadogCheck)

	case msgs.NoAccounts:
		m.scope.Debug("no accounts found")
		m.state.org = &msg.Org
		return m.goToGate(GateAccountCreate)

	case msgs.AccountCreated:
		m.scope.Info("account created", slog.String("account_id", msg.Account.ID.String()))
		m.state.org = &msg.Org
		m.state.account = &msg.Account
		m.services = m.services.WithAccountID(msg.Account.ID)
		return m.goToGate(GateDatadogCheck)

	case msgs.DatadogReady:
		m.scope.Debug("datadog ready")
		return m.goToGate(GateWorkspaceSelect)

	case msgs.DatadogNeeded:
		m.scope.Debug("datadog setup needed")
		return m.goToGate(GateDatadogRegion)

	case msgs.DatadogRegionSelected:
		m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
		m.state.ddSite = msg.Site
		return m.goToGate(GateDatadogAPIKey)

	case msgs.DatadogAPIKeyEntered:
		m.scope.Debug("datadog api key validated")
		m.state.ddAPIKey = msg.APIKey
		return m.goToGate(GateDatadogAppKey)

	case msgs.DatadogAccountCreated:
		m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
		m.state.ddAccount = msg.DatadogAccountID
		return m.goToGate(GateDatadogDiscovery)

	case msgs.DatadogDiscoveryComplete:
		m.scope.Info("datadog discovery complete")
		return m.goToGate(GateWorkspaceSelect)

	case msgs.WorkspaceSelected:
		m.scope.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
		m.state.workspace = &msg.Workspace
		return m.goToGate(GateSync)

	case msgs.SyncComplete:
		m.scope.Info("onboarding complete",
			slog.String("org_id", m.state.org.ID.String()),
			slog.String("account_id", m.state.account.ID.String()),
			slog.String("workspace_id", string(m.state.workspace.ID)),
		)
		return func() tea.Msg {
			return msgs.OnboardingComplete{
				User:      m.state.user,
				Org:       *m.state.org,
				Account:   *m.state.account,
				Workspace: *m.state.workspace,
			}
		}
	}

	return nil
}
