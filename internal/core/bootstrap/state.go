package bootstrap

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// State is the core bootstrap state needed for phase transitions.
type State struct {
	User      *auth.User
	Org       *domain.Organization
	Account   *domain.Account
	Workspace *domain.Workspace
	DDSite    domain.DatadogSite
	DDAPIKey  string
	DDAccount domain.DatadogAccountID
}

// PreflightResolved captures the resolved bootstrap context.
type PreflightResolved struct {
	Outcome      PreflightOutcome
	HasValidAuth bool
	Role         string
	Org          *domain.Organization
	Account      *domain.Account
}

// ApplyPreflight applies preflight resolution to state and returns next gate.
func ApplyPreflight(state State, resolved PreflightResolved) (State, Gate) {
	next := DecideNextGate(PreflightInput{
		Outcome:      resolved.Outcome,
		HasValidAuth: resolved.HasValidAuth,
		Role:         resolved.Role,
		HasOrg:       resolved.Org != nil,
		HasAccount:   resolved.Account != nil,
	})

	if resolved.Org != nil {
		state.Org = resolved.Org
	}
	if resolved.Account != nil {
		state.Account = resolved.Account
	}

	return state, next
}

func ApplyAuthenticated(state State, user auth.User) (State, Gate) {
	state.User = &user
	return state, GateRoleSelect
}

func ApplyRoleSelected(state State, role string) (State, Gate) {
	_ = role
	return state, GateOrgSelect
}

func ApplyOrgSelected(state State, org domain.Organization) (State, Gate) {
	state.Org = &org
	return state, GateAccountSelect
}

func ApplyNoOrgs(state State) (State, Gate) {
	return state, GateOrgCreate
}

func ApplyOrgCreated(state State, org domain.Organization) (State, Gate) {
	state.Org = &org
	return state, GateAccountSelect
}

func ApplyAccountSelected(state State, org domain.Organization, account domain.Account) (State, Gate) {
	state.Org = &org
	state.Account = &account
	return state, GateRuntimeInit
}

func ApplyNoAccounts(state State, org domain.Organization) (State, Gate) {
	state.Org = &org
	return state, GateAccountCreate
}

func ApplyAccountCreated(state State, org domain.Organization, account domain.Account) (State, Gate) {
	state.Org = &org
	state.Account = &account
	return state, GateRuntimeInit
}

func ApplyRuntimeReady(state State, org domain.Organization, account domain.Account) (State, Gate) {
	state.Org = &org
	state.Account = &account
	return state, GateDatadogCheck
}

func ApplyDatadogReady(state State) (State, Gate) {
	return state, GateWorkspaceSelect
}

func ApplyDatadogNeeded(state State) (State, Gate) {
	return state, GateDatadogRegion
}

func ApplyDatadogRegionSelected(state State, site domain.DatadogSite) (State, Gate) {
	state.DDSite = site
	return state, GateDatadogAPIKey
}

func ApplyDatadogAPIKeyEntered(state State, apiKey string) (State, Gate) {
	state.DDAPIKey = apiKey
	return state, GateDatadogAppKey
}

func ApplyDatadogAccountCreated(state State, datadogAccountID domain.DatadogAccountID) (State, Gate) {
	state.DDAccount = datadogAccountID
	return state, GateDatadogDiscovery
}

func ApplyDatadogDiscoveryComplete(state State) (State, Gate) {
	return state, GateWorkspaceSelect
}

func ApplyWorkspaceSelected(state State, workspace domain.Workspace) (State, Gate) {
	state.Workspace = &workspace
	return state, GateSync
}
