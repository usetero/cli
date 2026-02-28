package bootstrap

// EventFromMessage adapts onboarding message contracts to canonical events.
func EventFromMessage(msg Message) (Event, bool) {
	switch msg := msg.(type) {
	case PreflightResolved:
		return Event{
			Kind:      EventPreflightResolved,
			Preflight: msg.State,
		}, true
	case Authenticated:
		return Event{Kind: EventAuthenticated, User: msg.User}, true
	case RoleSelected:
		return Event{Kind: EventRoleSelected, Role: msg.Role}, true
	case OrgSelected:
		return Event{Kind: EventOrgSelected, Org: msg.Org}, true
	case NoOrgs:
		return Event{Kind: EventNoOrgs}, true
	case OrgCreated:
		return Event{Kind: EventOrgCreated, Org: msg.Org}, true
	case AccountSelected:
		return Event{Kind: EventAccountSelected, Org: msg.Org, Account: msg.Account}, true
	case NoAccounts:
		return Event{Kind: EventNoAccounts, Org: msg.Org}, true
	case AccountCreated:
		return Event{Kind: EventAccountCreated, Org: msg.Org, Account: msg.Account}, true
	case RuntimeReady:
		return Event{Kind: EventRuntimeReady, Org: msg.Org, Account: msg.Account}, true
	case DatadogReady:
		return Event{Kind: EventDatadogReady}, true
	case DatadogNeeded:
		return Event{Kind: EventDatadogNeeded}, true
	case DatadogRegionSelected:
		return Event{Kind: EventDatadogRegionSelected, Site: msg.Site}, true
	case DatadogAPIKeyEntered:
		return Event{Kind: EventDatadogAPIKeyEntered, APIKey: msg.APIKey}, true
	case DatadogAccountCreated:
		return Event{Kind: EventDatadogAccountCreated, DatadogAccountID: msg.DatadogAccountID}, true
	case DatadogDiscoveryComplete:
		return Event{Kind: EventDatadogDiscoveryDone}, true
	case WorkspaceSelected:
		return Event{Kind: EventWorkspaceSelected, Workspace: msg.Workspace}, true
	case SyncComplete:
		return Event{Kind: EventSyncComplete}, true
	default:
		return Event{}, false
	}
}
