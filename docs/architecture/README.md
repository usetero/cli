# Architecture

These docs exist to help you choose the right boundary before you change code.
Most regressions in this repo do not come from bad syntax; they come from
correct local edits made in the wrong layer.

Each page in this section should help you answer four questions quickly:

- why does this part of the system exist,
- how does it behave when things go wrong,
- what must not be broken,
- where do I confirm that in code.

Read this section when you are:

- moving responsibilities between packages,
- changing cross-cutting runtime behavior,
- deciding where new logic should live,
- trying to understand how the rebuilt app is organized.

## Suggested order

Start with:

1. [system-overview.md](system-overview.md)
2. [data-flow.md](data-flow.md)
3. [runtime-architecture.md](runtime-architecture.md)

Then, if you are working in the TUI:

4. [ui-runtime.md](ui-runtime.md)
5. [ui-messages.md](ui-messages.md)
6. [ui-layout.md](ui-layout.md)
7. [theme-and-chrome.md](theme-and-chrome.md)

That sequence mirrors the system itself:

- ownership first,
- truth and flow second,
- runtime/lifecycle third,
- UI structure and presentation last.

## How to use architecture docs with code

Use these pages to decide ownership first. Then open code under:

- [`cmd`](../../cmd)
- [`internal/domains`](../../internal/domains)
- [`internal/infrastructure`](../../internal/infrastructure)
- [`internal/runtime`](../../internal/runtime)
- [`internal/interfaces`](../../internal/interfaces)

The goal is not to memorize diagrams. The goal is to make design decisions that
stay consistent as the repository grows.
