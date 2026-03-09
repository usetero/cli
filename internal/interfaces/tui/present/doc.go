// Package present provides a typed presentation language for TUI content.
//
// The package defines semantic content nodes and renders them against a theme
// context. It owns content and surface rendering, while chrome owns shell/frame
// layout.
//
// The package has two main contracts:
//   - Node: generic presentation composition for page content.
//   - Block: structured multiline content that can be rendered safely inside
//     surfaces such as cards and notices.
//
// Surface constructors should prefer Block so they can own line splitting,
// width normalization, spacing, and background fill without depending on
// arbitrary pre-rendered ANSI content.
package present
