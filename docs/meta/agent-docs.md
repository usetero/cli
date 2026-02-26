# Agent Instruction Files

How to design high-signal instruction files for coding agents in this repository.

Canonical source of truth:

1. `AGENTS.md`

Generated sibling files:

1. `CLAUDE.md` (for tools that expect this filename)
2. Cursor rule files in `.cursor/rules/*.mdc` (if configured)

## Purpose

Instruction files are operational constraints loaded automatically by agent runtimes. Every line has cost. Every line should reduce failure risk.

Use these files to prevent wrong code, not to explain the codebase in prose.

## Two Documentation Systems

1. `docs/` is for humans:
   architecture, rationale, deep dives, background.
2. `AGENTS.md` is for agents:
   directives, workflows, constraints, known traps.

If correctness depends on a fact, it must be in `AGENTS.md`, not only in `docs/`.

## Inclusion Rule

Use this test for every line:

1. If an agent skips this line, can it still produce correct code?
2. If no: keep it in `AGENTS.md`.
3. If yes: move it to `docs/`.

## What Belongs in AGENTS.md

1. Decision constraints not obvious from local code.
2. Workflow sequences the agent cannot infer safely.
3. Canonical patterns when multiple plausible patterns exist.
4. "Always do X, never Y" rules where Y looks valid but is wrong here.
5. Recurring failure preventers from real mistakes.

## What Does Not Belong

1. API signatures/structs/constants visible in source.
2. Broad architecture prose better suited for human docs.
3. Historical narratives and one-off incidents.
4. Repetition of root guidance in domain files.

## Root vs Domain Files

1. Root `AGENTS.md`: global repository rules.
2. Domain `AGENTS.md`: local differences, workflows, traps.
3. If a rule applies broadly, move it to root.
4. Not every directory needs an `AGENTS.md`.

## Writing Style

1. Directive, not essay.
2. Concrete paths and commands.
3. Short and unambiguous.
4. Self-contained for correctness.

## Recommended Structure

Use only sections that add value:

1. Scope
2. Critical Rules
3. Workflows
4. Common Failure Modes

## Length Guidance

Keep files compact and high-signal.

Target:

1. Most domain files under ~120 lines.
2. Root file can be longer only for truly global constraints.

## Generation and Consistency

Edit `AGENTS.md` directly, then run your generation workflow (copy/symlink) to refresh `CLAUDE.md` and other derived artifacts.

If this repository adds a dedicated task, document it here.

## Quality Bar

An instruction file is good if it is:

1. Correct
2. Minimal
3. Actionable
4. Specific to its scope
5. Proven useful by reduced agent mistakes
