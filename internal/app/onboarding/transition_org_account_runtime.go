package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

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
