package onboarding

// gateDisplayPolicy defines default rendering behavior for each gate.
type gateDisplayPolicy struct {
	hidden bool
	status string
}

var gateDisplayPolicies = map[Gate]gateDisplayPolicy{
	GateRuntimeInit: {hidden: true, status: "Initializing account runtime..."},
}

func displayPolicyForGate(gate Gate) gateDisplayPolicy {
	if p, ok := gateDisplayPolicies[gate]; ok {
		return p
	}
	return gateDisplayPolicy{hidden: false}
}

// VisibilityProvider is an optional step contract for transient step rendering.
type VisibilityProvider interface {
	Hidden() bool
	StatusText() string
}
