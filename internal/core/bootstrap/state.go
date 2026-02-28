package bootstrap

import "github.com/usetero/cli/internal/domain"

// State is the core bootstrap state needed for phase transitions.
type State struct {
	Org     *domain.Organization
	Account *domain.Account
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
