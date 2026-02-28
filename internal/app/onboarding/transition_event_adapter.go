package onboarding

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
)

func bootstrapEventFor(msg tea.Msg) (bootstrap.Event, bool) {
	switch msg := msg.(type) {
	case msgs.PreflightResolved:
		return bootstrap.Event{
			Kind: bootstrap.EventPreflightResolved,
			Preflight: bootstrap.PreflightResolved{
				Outcome:      bootstrap.PreflightOutcome(msg.State.Outcome),
				HasValidAuth: msg.State.HasValidAuth,
				Role:         msg.State.Role,
				Org:          msg.State.Org,
				Account:      msg.State.Account,
			},
		}, true
	case msgs.Authenticated:
		return bootstrap.Event{Kind: bootstrap.EventAuthenticated, User: msg.User}, true
	case msgs.RoleSelected:
		return bootstrap.Event{Kind: bootstrap.EventRoleSelected, Role: msg.Role}, true
	case msgs.OrgSelected:
		return bootstrap.Event{Kind: bootstrap.EventOrgSelected, Org: msg.Org}, true
	case msgs.NoOrgs:
		return bootstrap.Event{Kind: bootstrap.EventNoOrgs}, true
	case msgs.OrgCreated:
		return bootstrap.Event{Kind: bootstrap.EventOrgCreated, Org: msg.Org}, true
	case msgs.AccountSelected:
		return bootstrap.Event{Kind: bootstrap.EventAccountSelected, Org: msg.Org, Account: msg.Account}, true
	case msgs.NoAccounts:
		return bootstrap.Event{Kind: bootstrap.EventNoAccounts, Org: msg.Org}, true
	case msgs.AccountCreated:
		return bootstrap.Event{Kind: bootstrap.EventAccountCreated, Org: msg.Org, Account: msg.Account}, true
	case msgs.RuntimeReady:
		return bootstrap.Event{Kind: bootstrap.EventRuntimeReady, Org: msg.Org, Account: msg.Account}, true
	case msgs.DatadogReady:
		return bootstrap.Event{Kind: bootstrap.EventDatadogReady}, true
	case msgs.DatadogNeeded:
		return bootstrap.Event{Kind: bootstrap.EventDatadogNeeded}, true
	case msgs.DatadogRegionSelected:
		return bootstrap.Event{Kind: bootstrap.EventDatadogRegionSelected, Site: msg.Site}, true
	case msgs.DatadogAPIKeyEntered:
		return bootstrap.Event{Kind: bootstrap.EventDatadogAPIKeyEntered, APIKey: msg.APIKey}, true
	case msgs.DatadogAccountCreated:
		return bootstrap.Event{Kind: bootstrap.EventDatadogAccountCreated, DatadogAccountID: msg.DatadogAccountID}, true
	case msgs.DatadogDiscoveryComplete:
		return bootstrap.Event{Kind: bootstrap.EventDatadogDiscoveryDone}, true
	case msgs.WorkspaceSelected:
		return bootstrap.Event{Kind: bootstrap.EventWorkspaceSelected, Workspace: msg.Workspace}, true
	case msgs.SyncComplete:
		return bootstrap.Event{Kind: bootstrap.EventSyncComplete}, true
	default:
		return bootstrap.Event{}, false
	}
}
