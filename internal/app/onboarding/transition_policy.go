package onboarding

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

// transitionOutcomeFor is the single transition policy entrypoint.
// It maps onboarding messages to the next gate and state mutations.
func (m *Model) transitionOutcomeFor(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.PreflightResolved:
		return m.handlePreflightResolved(msg), true
	case msgs.Authenticated:
		return m.handleAuthenticated(msg), true
	case msgs.RoleSelected:
		return m.handleRoleSelected(msg), true
	case msgs.OrgSelected:
		return m.handleOrgSelected(msg), true
	case msgs.NoOrgs:
		return m.handleNoOrgs(), true
	case msgs.OrgCreated:
		return m.handleOrgCreated(msg), true
	case msgs.AccountSelected:
		return m.handleAccountSelected(msg), true
	case msgs.NoAccounts:
		return m.handleNoAccounts(msg), true
	case msgs.AccountCreated:
		return m.handleAccountCreated(msg), true
	case msgs.RuntimeReady:
		return m.handleRuntimeReady(msg), true
	case msgs.DatadogReady:
		return m.handleDatadogReady(), true
	case msgs.DatadogNeeded:
		return m.handleDatadogNeeded(), true
	case msgs.DatadogRegionSelected:
		return m.handleDatadogRegionSelected(msg), true
	case msgs.DatadogAPIKeyEntered:
		return m.handleDatadogAPIKeyEntered(msg), true
	case msgs.DatadogAccountCreated:
		return m.handleDatadogAccountCreated(msg), true
	case msgs.DatadogDiscoveryComplete:
		return m.handleDatadogDiscoveryComplete(), true
	case msgs.WorkspaceSelected:
		return m.handleWorkspaceSelected(msg), true
	case msgs.SyncComplete:
		return m.handleSyncComplete(), true
	default:
		return noop(), false
	}
}
