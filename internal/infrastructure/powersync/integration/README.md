# PowerSync Integration Test Charter

This suite protects behavior across package boundaries in the PowerSync stack.

## Invariants

- Pipeline: queued local mutations are uploaded, acknowledged by checkpoint, and removed from queue.
- Lifecycle: sync and upload loops start/stop/restart safely without deadlock.
- Recovery: transient failures emit stalled signals and later recover while preserving forward progress.
- Replay robustness: real captured stream lines are processed without entering fatal syncer state.

## Scope

- Uses real extension-backed SQLite (`powersync_control`, `ps_*` tables).
- Uses deterministic in-memory fakes for HTTP/token/handlers only.
- Focuses on user-visible runtime semantics, not line coverage.

## File layout

- `pipeline_test.go`: happy-path end-to-end flow.
- `lifecycle_test.go`: start/stop/restart and cancellation safety.
- `recovery_test.go`: stalled -> recovered behavior and eventual drain.
- `replay_test.go`: replay captured NDJSON streams for robustness.
- `testkit_test.go`: hermetic helpers/fakes shared by this suite.
