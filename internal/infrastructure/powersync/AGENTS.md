# PowerSync Infrastructure

This package owns the concrete PowerSync runtime integration in the CLI:

- loading the SQLite extension,
- applying the generated PowerSync client schema artifact,
- running syncer/uploader control flows.

It does not own canonical client-schema authorship.

## Ownership

- The canonical PowerSync client schema inputs live in the sibling
  control-plane repo under `internal/infra/powersync/`.
- The CLI consumes generated artifacts from
  `internal/infrastructure/sqlite/`.
- The extension runtime should apply
  `internal/infrastructure/sqlite/powersync_schema.json`, not a locally
  invented schema artifact.

## Boundaries

- Do not add local authored indexes here.
- Do not reintroduce live-service schema generation as the primary path.
- Do not put app-specific SQLite views/functions in this package.

Those belong either:
- in control-plane PowerSync schema authorship, if they are part of the synced
  client schema contract, or
- in CLI SQLite infrastructure, if they are local Tero query capabilities.
