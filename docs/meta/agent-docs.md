# Agent Docs Standard

`AGENTS.md` files are operational constraint files for contributors and coding
agents. They are not substitutes for the manual in `docs/`.

Use `AGENTS.md` to prevent predictable mistakes inside a local part of the tree.
Use the manual to explain the system and the recurring doctrine behind it.

## What Belongs In `AGENTS.md`

Put correctness-critical local instructions there, especially when violating the
instruction would produce common or expensive mistakes.

Examples:

- layer boundaries that are easy to violate locally,
- workflow constraints for a sensitive directory,
- rules about tests, code generation, or file ownership,
- small local conventions that protect maintainability.

## What Does Not Belong There

Do not turn `AGENTS.md` into a second architecture manual.

If a concept needs explanation, motivation, and broad mental context, it
belongs in `docs/` and should be referenced from `AGENTS.md` when necessary.
