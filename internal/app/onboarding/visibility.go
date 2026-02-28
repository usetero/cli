package onboarding

// VisibilityProvider is an optional step contract for transient step rendering.
type VisibilityProvider interface {
	Hidden() bool
	StatusText() string
}
