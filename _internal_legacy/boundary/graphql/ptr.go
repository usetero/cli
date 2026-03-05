package graphql

// ptr returns a pointer to the given value.
// Useful for setting optional fields in generated GraphQL inputs.
func ptr[T any](v T) *T {
	return &v
}

// deref safely dereferences a pointer, returning the zero value if nil.
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
