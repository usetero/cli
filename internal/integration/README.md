# Integration Test Layout

This directory separates deterministic integration checks from live environment checks.

- `hermetic/` (`//go:build integration`): deterministic, no external services.
- `live/` (`//go:build integration_live`): non-production end-to-end checks against real services.

Use task entrypoints from `Taskfile.yml`:

- `task test:integration`
- `task test:integration:live`
- `task test:tags:compile`
