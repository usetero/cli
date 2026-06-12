package bootstrap

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// EventKind identifies a deterministic bootstrap transition input.
type EventKind string

const (
	EventPreflightResolved     EventKind = "preflight_resolved"
	EventAuthenticated         EventKind = "authenticated"
	EventRoleSelected          EventKind = "role_selected"
	EventOrgSelected           EventKind = "org_selected"
	EventNoOrgs                EventKind = "no_orgs"
	EventOrgCreated            EventKind = "org_created"
	EventAccountSelected       EventKind = "account_selected"
	EventNoAccounts            EventKind = "no_accounts"
	EventAccountCreated        EventKind = "account_created"
	EventRuntimeReady          EventKind = "runtime_ready"
	EventDatadogReady          EventKind = "datadog_ready"
	EventDatadogNeeded         EventKind = "datadog_needed"
	EventDatadogRegionSelected EventKind = "datadog_region_selected"
	EventDatadogAPIKeyEntered  EventKind = "datadog_apikey_entered"
	EventDatadogAccountCreated EventKind = "datadog_account_created"
	EventDatadogDiscoveryDone  EventKind = "datadog_discovery_done"
)

// Event is the canonical transition input consumed by the bootstrap engine.
type Event struct {
	Kind             EventKind
	Preflight        PreflightState
	User             auth.User
	Role             string
	Org              domain.Organization
	Account          domain.Account
	Site             domain.DatadogSite
	APIKey           string
	DatadogAccountID domain.DatadogAccountID
}

// TransitionKind is the deterministic output shape for ApplyEvent.
type TransitionKind string

const (
	TransitionNoop     TransitionKind = "noop"
	TransitionAdvance  TransitionKind = "advance"
	TransitionComplete TransitionKind = "complete"
)

// Transition is the output of applying one event to bootstrap state.
type Transition struct {
	Kind       TransitionKind
	State      State
	Next       Gate
	Completion Completion
}

// ApplyEvent applies one canonical bootstrap event.
func ApplyEvent(state State, event Event) Transition {
	switch event.Kind {
	case EventPreflightResolved:
		nextState, next := ApplyPreflight(state, event.Preflight)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventAuthenticated:
		nextState, next := ApplyAuthenticated(state, event.User)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventRoleSelected:
		nextState, next := ApplyRoleSelected(state, event.Role)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventOrgSelected:
		nextState, next := ApplyOrgSelected(state, event.Org)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventNoOrgs:
		nextState, next := ApplyNoOrgs(state)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventOrgCreated:
		nextState, next := ApplyOrgCreated(state, event.Org)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventAccountSelected:
		nextState, next := ApplyAccountSelected(state, event.Org, event.Account)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventNoAccounts:
		nextState, next := ApplyNoAccounts(state, event.Org)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventAccountCreated:
		nextState, next := ApplyAccountCreated(state, event.Org, event.Account)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventRuntimeReady:
		nextState, next := ApplyRuntimeReady(state, event.Org, event.Account)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventDatadogReady:
		// Datadog already configured: onboarding is complete (no workspace step).
		return completeOrNoop(state)
	case EventDatadogNeeded:
		nextState, next := ApplyDatadogNeeded(state)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventDatadogRegionSelected:
		nextState, next := ApplyDatadogRegionSelected(state, event.Site)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventDatadogAPIKeyEntered:
		nextState, next := ApplyDatadogAPIKeyEntered(state, event.APIKey)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventDatadogAccountCreated:
		nextState, next := ApplyDatadogAccountCreated(state, event.DatadogAccountID)
		return Transition{Kind: TransitionAdvance, State: nextState, Next: next}
	case EventDatadogDiscoveryDone:
		// Datadog discovery finished: onboarding is complete. The account is the
		// working context; there is no workspace step.
		return completeOrNoop(state)
	default:
		return Transition{Kind: TransitionNoop, State: state}
	}
}

// completeOrNoop completes onboarding when the required state is present,
// otherwise no-ops.
func completeOrNoop(state State) Transition {
	completion, ok := CompleteOnboarding(state)
	if !ok {
		return Transition{Kind: TransitionNoop, State: state}
	}
	return Transition{Kind: TransitionComplete, State: state, Completion: completion}
}
