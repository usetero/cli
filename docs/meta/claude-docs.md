# CLAUDE.md Files

How to write effective AI agent instructions for this codebase.

## The Core Constraint

Every line in a `CLAUDE.md` is injected into the prompt for every interaction in that directory. This is both its power and its cost. Content is guaranteed to be read — unlike `docs/` files, which require the agent to choose to open them. But every line consumes context window, diluting attention on the actual task.

This shapes everything: what goes in, what stays out, how long it gets.

## Two Systems

**`docs/` — for humans.** Mental models, decisions, deep dives. Engineers read these to build understanding.

**`CLAUDE.md` — for AI agents.** Directives, workflows, mistake prevention. Auto-loaded into the prompt. The agent follows these whether it wants to or not.

Same information, different extraction. Don't collapse them into one system.

## What Belongs in CLAUDE.md

**"If the agent skips this, will it produce wrong code?"**

**Yes → CLAUDE.md.** Auto-loaded, can't be skipped.

**No → `docs/`.** Reference from `CLAUDE.md` if useful, but correctness can't depend on it.

Rules produce correct code. Context produces informed code. `CLAUDE.md` is for rules.

## What Does NOT Belong in CLAUDE.md

The agent can read the code. Don't restate what's visible in the source — it wastes context window and goes stale.

**Don't document:**
- Entity descriptions (the agent can read the schema)
- Method signatures and constructor names (the agent can read the code)
- Internal implementation details like constants, thresholds, retry counts (the agent can read the code)
- How a specific algorithm works (the agent can read the code)

**Do document:**
- Which file to follow as the canonical example (when multiple exist)
- Multi-step workflows the agent can't infer (run X, then Y, then Z)
- Architectural rules that are decisions, not visible in code ("display methods go on the domain type, not in extract layers")
- What NOT to do — mistakes that look correct but break things
- Consistency rules — "always do X this way" when the codebase has multiple valid-looking approaches

The signal: if the agent would write correct code after reading the source files, the rule doesn't need to exist. CLAUDE.md fills the gaps between what the code shows and what the agent needs to know.

## Root vs Domain

The root `CLAUDE.md` carries codebase-wide patterns: style, architecture, conventions, testing approach, commit format. It's always loaded.

Domain `CLAUDE.md` files carry only what's unique to that domain:
- Domain-specific workflows (how to add a new policy category, how to write a query)
- Domain-specific gotchas (background color discipline, sqlc naming)
- Deviations from root patterns (if any)

**If a rule applies to more than one domain, it belongs in root.** Domain docs should never repeat root patterns — the agent sees both. A domain doc that restates testing conventions or commit style is wasting context window.

**Not every domain needs a CLAUDE.md.** If a domain follows root conventions, has no special workflows, and no known gotchas, it doesn't need one. Don't create empty docs to fill a template.

## Self-Contained for Correctness

A `CLAUDE.md` must never depend on the agent reading another file to produce correct output. References to `docs/` are for depth, never for correctness.

```markdown
# Good: rules inline, reference for depth
## Background Colors

Every lipgloss style MUST include `.Background(theme.Bg)`. To change a
child's background, pass `theme.WithBg(newColor)` — never remove backgrounds.

For the full layout system, see `docs/TEA.md`.
```

```markdown
# Bad: correctness behind a gate
## Background Colors

Read `docs/TEA.md` before writing any UI code.
```

The first version works even if the agent never opens `TEA.md`. The second fails silently.

## Template

```markdown
# {Domain}

One line: what this directory owns.

## How It Works

The structural overview an agent needs to understand how the pieces
connect. Not implementation details — the pipeline, the stages, the
data flow, the organization.

Include this when: the domain has meaningful complexity spread across
multiple files/subdirectories where reading any single file doesn't
reveal the overall system.

Skip this when: the domain is a single-concern package where reading
the main file gives you the full picture.

Keep it structural — stages, flow, organization. Not prose or
philosophy. An agent should read this and know where to look for
what, and how things connect.

## Patterns

Rules the agent would get wrong by reading the code alone.

For each rule: "Would the agent follow a different, wrong pattern
if this didn't exist?" If the code makes it obvious, skip it.
If root CLAUDE.md already covers it, don't repeat it.

Common entries:
- Canonical file to follow (when multiple valid-looking examples exist)
- Domain-specific deviations from root patterns
- "Always X, never Y" rules where Y looks reasonable

## Workflows

Sequences the agent can't reconstruct from the code.

For each workflow: would the agent discover these steps by reading
the source? If yes, skip it.

Common entries:
- Multi-file change sequences (edit A, then regenerate B, then run C)
- External tool dependencies (specific task commands)
- Order-sensitive operations

## Common Mistakes

Mistakes that have actually happened or would look correct to the agent.

For each mistake: "An agent reading this code would reasonably do X,
but X is wrong because Y."

Format: **What goes wrong** — what to do instead.
```

Every section is optional. If no section has content, the domain doesn't need a CLAUDE.md.

## Writing Style

Directives, not prose. Write for a capable but literal colleague who follows exactly what you say and skips what you don't.

**Direct.** "Run `task do` after changes." Not "You may want to consider running..."

**Specific.** "Add the analysis type to `policy_analysis_waste.go`." Not "Add the type to the appropriate file."

**Concrete.** Actual commands, actual file paths, actual patterns. Generalities get misinterpreted.

**Brief.** If a rule takes a paragraph, the explanation belongs in `docs/` and a one-line summary belongs here.

## Length

Each `CLAUDE.md` covers its directory. Don't repeat rules from the root — it's always loaded. Don't cover sibling directories — they have their own.

**Aim for under 150 lines.** Past that, the context window cost compounds and agent attention degrades. Complex domains can go longer if every line earns its place.

If a `CLAUDE.md` grows past budget, ask what can move to `docs/` without sacrificing correctness.

## Evolution

Most rules originate the same way: an AI agent makes a mistake, a human fixes it, it happens again. The fix belongs in a `CLAUDE.md` the moment it recurs.

When adding a rule, ask: is this a root pattern or a domain gotcha? Root pattern → update root. Domain gotcha → update the domain doc.

Update when:
- An agent makes the same mistake twice
- A workflow changes
- A new pattern is established

Don't update when:
- The code is self-explanatory
- The rule is already in the root `CLAUDE.md`
- It's a one-off edge case

## The Test

Could an AI agent read this `CLAUDE.md` and produce correct code on its first attempt for common tasks in this directory?

If it would still make obvious mistakes, the doc is missing a rule. If it's so long the agent loses focus, the doc needs trimming. Both fail the same way: wrong code.
