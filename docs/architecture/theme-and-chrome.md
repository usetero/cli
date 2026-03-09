# Theme And Chrome

This page defines the presentation contract for reusable TUI visuals.

Layout decides where content goes. Theme, `present`, and chrome decide how the
shared UI looks.

## Theme owns semantic tokens

The theme should provide semantic styles, not random per-screen colors.

Examples:

- shell surfaces
- cards
- text roles
- list roles
- input roles
- progress roles

That keeps presentation coherent and makes it possible to improve quality
globally instead of patching screen-specific styling forever.

Theme is also the active background context.

Terminals do not have transparent layers in any useful design sense. If a
surface changes background, that surface should pass a derived theme with the
new background to its children. Children should render against
`theme.Background` and should not need to know whether they are on the page
background or a raised surface.

## Present owns reusable surfaces

`present` is the typed presentation language for reusable content and surfaces.

Examples:

- cards
- error cards
- notices and status blocks
- text roles
- content stacks and rows

If something is purely presentational and should be shared across multiple
screens, it belongs in `present`.

Use typed builders and semantic nodes, not arbitrary pre-rendered strings.
Surfaces should accept structured content so background, padding, and spacing
stay centralized.

Inside `present`, use two levels of content:

- `Node`: generic composition for page content
- `Block`: structured multiline content for surfaces

That split matters. Surfaces should prefer `Block` so they can own width,
spacing, and background fill without inheriting reset bugs from arbitrary ANSI
blobs.

## Chrome owns shell framing

`chrome` is for shell-only presentation with no interaction state.

Examples:

- shell framing
- header/body/footer layout
- body placement rules
- shared brand motifs

If the concern is "where does this content go in the app frame?", it belongs in
`chrome`. If the concern is "what surface/text should this content render
through?", it belongs in `present`.

## Components own interaction

Interactive models belong in `components`, not `chrome`.

Examples:

- text inputs
- selectable lists
- progress widgets that manage state
- help-bar or status widgets if they manage their own behavior

The split is:

- chrome: shell and frame layout
- present: typed content and surface rendering
- components: interactive, Bubble Tea models

Inside `chrome`, keep the responsibilities explicit:

- shell/layout: viewport regions, measurement, body placement
- brand: wordmark and recurring brand motifs

Inside `present`, keep the responsibilities explicit:

- text roles: title, body, muted, error, success
- layout nodes: stack, row, raw embedded component output
- blocks: structured multiline content for surfaces
- surfaces: cards, notices, status blocks, sections, fields

## Screen styling policy

Screens should compose semantic theme tokens, `present` builders, and shared
components.

Screens should not:

- hard-code ad hoc colors,
- invent one-off borders and spacing for standard states,
- duplicate card or footer presentation logic,
- pass arbitrary ANSI blobs into surfaces that should own their own layout.

If multiple screens need the same visual pattern, promote it into a `present`
primitive.

## The help bar and status surfaces

The help bar and status bar are not incidental text.

They are chrome-level user orientation surfaces and should feel intentional:

- stable placement,
- consistent muted/accent treatment,
- predictable spacing,
- shared shell rhythm.

That is a presentation-system concern, not a per-screen concern.

## Quality bar

A good presentation system in this repo should let an engineer improve visual
quality by changing a small number of shared surfaces and semantic tokens.

If a visual improvement requires touching many screens individually, the system
is not abstracted at the right level yet.

## Failure patterns

The presentation system is drifting when:

- screens hard-code colors or borders directly,
- the same visual pattern is implemented differently in multiple places,
- chrome starts owning interaction state,
- chrome starts owning content surfaces instead of frame layout,
- present starts acting like a generic style bag instead of a typed rendering
  language,
- components start duplicating shell or card presentation logic.
- shared styles omit explicit backgrounds and start leaking the terminal's
  default background through nested surfaces.

Those are signals to promote or simplify a shared primitive, not to add more
screen-specific styling.

## Fast code entry points

- [`internal/interfaces/tui/theme`](../../internal/interfaces/tui/theme)
- [`internal/interfaces/tui/chrome`](../../internal/interfaces/tui/chrome)
- [`internal/interfaces/tui/present`](../../internal/interfaces/tui/present)
- [`internal/interfaces/tui/components/helpbar`](../../internal/interfaces/tui/components/helpbar)
