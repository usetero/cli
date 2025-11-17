# Patterns & Rules

Non-obvious rules that prevent mistakes.

## Configuration

**All CLI config goes in `CLIConfig` (`internal/config/cli.go`). Never read `os.Getenv()` elsewhere.**

Why: One place to find all config. No surprise env vars scattered throughout the code.

## Services & Interfaces

**Services define domain logic. Interfaces stay generic.**

Example: `auth.Service` owns authentication (tokens, users). `SecureStorage` is just `Get/Set/Delete`. This lets you swap WorkOS without changing auth logic.

Why: Easy to test, easy to swap implementations.
