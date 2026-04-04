# SQLite Infrastructure

This package owns the local SQLite runtime kernel and the generated client-side
schema artifacts consumed by the CLI.

## Ownership

- `schema.sql`, `jsonb_schema.json`, and `powersync_schema.json` are generated.
- Their canonical source lives in the sibling control-plane repo under
  `internal/infra/powersync/`.
- This package copies those generated artifacts during `task generate:schema`
  and exposes the PowerSync schema artifact to the extension runtime.
- For cross-repo worktree changes, set `TERO_CONTROL_PLANE_ROOT` to the desired
  control-plane checkout before running `go generate` or `task generate:schema`.

## Boundaries

- Do not hand-author local schema or index truth here.
- Do not add Tero-specific analytical SQL primitives to the PowerSync schema
  artifact. Those belong as local SQLite capabilities layered on top.
- Do not treat deployed PowerSync state as the authoring source. The control
  plane repo is canonical; deployed drift should be checked separately.

## Common Mistakes

- Editing `schema.sql`, `jsonb_schema.json`, or `powersync_schema.json`
  directly.
- Reintroducing CLI-local index ownership.
- Putting app-specific views/functions in the PowerSync extension package
  instead of the SQLite infrastructure layer.
