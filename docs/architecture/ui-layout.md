# UI Layout

This page defines the layout contract for the rebuilt TUI.

The goal is to make placement predictable and reusable, not to fix layout
screen by screen forever.

## The shell model

The root shell owns regions.

Conceptually, the app has three regions:

- a top-aligned header region,
- a flexible middle body region,
- a bottom-docked footer/help region.

The root and chrome own those regions. Screens do not.

## Why this contract exists

The UI should feel terminal-native and consistent across flows. That only
happens if shell structure is owned at the top and screen content is placed
within a stable layout system.

Without that contract, every screen starts inventing spacing, max width, and
alignment locally, and the app quality regresses one fix at a time.

## The body-frame model

Once header and footer are measured, chrome owns the body rectangle.

Chrome should place rendered content in that rectangle according to layout
hints, not according to ad hoc per-screen padding.

That means the parent owns:

- body size,
- header/footer docking,
- content placement rules,
- shared shell spacing.

The child owns:

- its own content tree through `present`,
- optional layout hints,
- no shell-aware padding hacks.

## Layout hints

The useful layout vocabulary for this app is small:

- `WidthMode`: `Fill` or `Intrinsic`
- `HeightMode`: `Fill` or `Intrinsic`
- `VerticalAlign`: `Top`, `Bottom`, or `Center`
- `MaxWidth`: optional width cap for readable content blocks

Those hints scale much better than "every screen sets random padding until it
looks okay."

## Default expectations

These defaults should guide the app:

- header/status bar is top-aligned
- help bar is bottom-docked
- body is flexible
- onboarding body content is usually bottom-aligned and intrinsic-height
- status and empty states often want centered or bottom-aligned intrinsic
  content
- long-running surfaces such as chat often want fill-height behavior

The important point is that these are shell or page-frame decisions, not leaf
widget decisions.

## Fixed versus intrinsic sizing

Do not think of screens only as "fixed height" or "variable height."

The better split is:

- fixed shell regions are measured by the parent
- the child may fill the remaining region or render at intrinsic size
- chrome places the child inside the available body region

That gives us a system that scales as more surfaces are added.

## Width responsibility

Children should not own terminal-wide max width policy.

Chrome or page-frame helpers should provide readable constraints such as
`MaxWidth` for cards, forms, and status views. That keeps width policy
consistent across screens.

## Failure patterns

When layout ownership drifts, the symptoms are predictable:

- content gets stranded at the top of a tall terminal,
- footer/help surfaces float or collapse unpredictably,
- screens add one-off padding hacks to look correct,
- cards and forms use inconsistent widths and spacing.

If that happens, fix chrome or page-frame policy first. Do not patch several
screens independently and call the problem solved.

## Error and emphasis surfaces

Errors, alerts, and other emphasized content should use shared `present`
surfaces. They should not each invent their own border, padding, and width
rules.

This is how the UI gets visually cohesive without every screen becoming a
custom art project.

## Where this maps to code

Today the key layout code lives in:

- [`internal/interfaces/tui/root`](../../internal/interfaces/tui/root)
- [`internal/interfaces/tui/chrome/layout.go`](../../internal/interfaces/tui/chrome/layout.go)
- reusable surfaces under [`internal/interfaces/tui/present`](../../internal/interfaces/tui/present)
- screen models under [`internal/interfaces/tui/screens`](../../internal/interfaces/tui/screens)

When layout feels wrong, fix the shell contract or shared chrome first. Do not
start by patching individual screens unless the problem is truly local.
