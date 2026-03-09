// Package chrome provides shared, stateless TUI shell primitives.
//
// Chrome owns frame layout and brand rendering, not content surfaces or
// interaction state.
//
// Internal organization:
//   - layout.go: shell measurement and body placement
//   - brand.go: shared brand rendering
package chrome
