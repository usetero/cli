# Patterns & Rules

Rules to keep the codebase consistent. When in doubt, look here first.

## Configuration

**Rule: All CLI config goes in `CLIConfig` (`internal/config/cli.go`). Never read `os.Getenv()` anywhere else.**

- Production defaults hardcoded in `getDefault*()` functions
- Environment variables override defaults (read once in `LoadCLIConfig()`)
- Pass config explicitly down the stack—no globals
- CLI flags only for user-facing options (like `--endpoint` for self-hosted)
- Env vars only for internal/dev config (like `TERO_WORKOS_CLIENT_ID` for staging)

**Why:** One place to find all config. No surprise env vars scattered in code.

## Environment Files

**Rule: Use `.envrc` (checked in) + `.envrc.local` (gitignored) for configuration.**

- `.envrc` has development defaults (local control plane, staging auth)
- `.envrc.local` for personal overrides (not tracked)
- No `.env` files

**Why:** direnv pattern. Simple. Consistent.

## Secrets

**Rule: OAuth tokens go in OS keychain (`internal/keyring`). Nothing else.**

- Access tokens and refresh tokens only
- Use `task reset:auth` to clear

**Why:** These tokens grant full API access. Keychain encrypts them properly.

## Services & Interfaces

**Rule: Services define domain logic. Interfaces stay generic.**

Example: `auth.Service` owns authentication concepts (tokens, users). `SecureStorage` interface is just `Get/Set/Delete`. This way you can swap WorkOS for something else without changing auth logic.

**Why:** Easy to test, easy to swap implementations, clear boundaries.

## Commits

**Rule: Conventional commits. One logical change per commit.**

Format: `type: description`
- `feat:` - New feature
- `fix:` - Bug fix  
- `chore:` - Maintenance
- `docs:` - Documentation

**Why:** Automated releases, clear history, easy reverts.

## Testing

**Rule: `task signoff` must pass before pushing.**

Runs: format, lint (Go + shell), tests, build, install script test.

**Why:** Catches issues before CI. Fast feedback loop.
