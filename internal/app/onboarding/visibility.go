package onboarding

// gateDisplayPolicy defines default rendering behavior for each gate.
type gateDisplayPolicy struct {
	hidden bool
	status string
}

func (m *Model) displayPolicyForGate(gate Gate) gateDisplayPolicy {
	return m.definitionForGate(gate).display
}

// VisibilityProvider is an optional step contract for transient step rendering.
type VisibilityProvider interface {
	Hidden() bool
	StatusText() string
}
