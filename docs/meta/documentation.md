# Documentation Standard

These docs should help an engineer build the right mental model of the system
and then make correct changes quickly.

Good documentation in this repository should:

1. explain why a part of the system exists,
2. clarify what owns what,
3. encode the stable rules that prevent drift,
4. point a reader toward the right code when needed.

## Writing Guidance

Prefer durable explanations over inventories of current implementation detail.

Start with context, then explain shape and ownership, then anchor to a few key
code paths where helpful. Avoid writing pages that only list files, commands, or
headings with no model behind them.

## Maintenance Guidance

Treat docs as part of the implementation contract:

1. update docs when the underlying model changes,
2. remove stale material aggressively,
3. keep one canonical page per concept,
4. rewrite cleanly when the mental model changes instead of patching around
   drift.
