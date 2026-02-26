# CLI

CLI commands are presentation adapters over API/services.

## Design Rules

1. Keep command handlers thin.
2. Put remote operations behind `internal/api` or service boundaries.
3. Keep formatting separate from transport logic.
4. Return actionable errors; do not leak raw internals by default.

## Recommended Command Shape

1. Parse args/flags.
2. Validate request intent.
3. Build and execute API/service call.
4. Render structured output.
5. Return typed errors for caller and tests.

## Testing

1. Unit test argument parsing and validation.
2. Unit test output formatting contracts.
3. Mock remote dependencies at boundaries.
4. Avoid network in unit tests.

## Anti-Patterns

1. Embedding business policy in command handlers.
2. Mixing parsing, remote calls, and rendering in one function.
3. Assertions against incidental string formatting in unrelated tests.
