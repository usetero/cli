# CLI

CLI commands are presentation-only entrypoints that call the control plane.

## Principles

1. Keep command handlers thin.
2. Put remote calls behind `internal/api`.
3. Avoid embedding business logic in command parsing/output formatting.

## Typical Shape

1. Parse flags/args.
2. Build API request.
3. Execute request.
4. Render result.

## Testing

1. Unit-test argument/format behavior.
2. Mock remote clients at boundaries.
3. Prefer deterministic output assertions.

