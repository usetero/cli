package onboarding

// gateDisplayPolicy defines default rendering behavior for each gate.
type gateDisplayPolicy struct {
	hidden bool
	status string
}

func (m *Model) displayPolicyForGate(gate Gate) gateDisplayPolicy {
	if def, ok := m.definitions[gate]; ok {
		return def.display
	}
	return gateDisplayPolicy{hidden: false}
}

// VisibilityProvider is an optional step contract for transient step rendering.
type VisibilityProvider interface {
	Hidden() bool
	StatusText() string
}
