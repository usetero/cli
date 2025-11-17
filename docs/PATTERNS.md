# Code Patterns

This document captures architectural patterns and decisions in the Tero CLI. It explains the "why" behind our code organization and provides guidance for maintaining consistency as the codebase evolves.

## Configuration

### CLI Configuration Pattern

**Pattern:** Single source of truth in `CLIConfig` with environment variable overrides.

**Location:** `internal/config/cli.go`

**How it works:**

1. **Define configuration in `CLIConfig` struct** - All CLI-level configuration lives here
2. **Set production defaults in code** - `getDefault*()` functions return production values
3. **Allow environment variable overrides** - `LoadCLIConfig()` reads from env vars
4. **Pass config explicitly down the stack** - No global state, clear data flow

**Example:**

```go
type CLIConfig struct {
    APIEndpoint    string  // Tero control plane endpoint
    WorkOSClientID string  // WorkOS OAuth client ID
    Debug          bool    // Debug logging
}

func LoadCLIConfig(version string) *CLIConfig {
    cfg := &CLIConfig{
        APIEndpoint:    getDefaultAPIEndpoint(version),
        WorkOSClientID: getDefaultWorkOSClientID(),
        Debug:          false,
    }
    
    // Override from environment variables
    if endpoint := os.Getenv("TERO_API_ENDPOINT"); endpoint != "" {
        cfg.APIEndpoint = endpoint
    }
    
    return cfg
}
```

**When to expose as CLI flag:**

- User-facing configuration (e.g., `--endpoint` for self-hosted deployments)
- Frequently changed options (e.g., `--debug`)

**When to keep as env var only:**

- Internal/developer configuration (e.g., `TERO_WORKOS_CLIENT_ID` for staging)
- Infrastructure configuration that users shouldn't normally change

**Development vs Production:**

- **Development defaults** (`.envrc`): Local control plane, staging auth
- **Production defaults** (hardcoded): Production control plane, production auth
- Developers override via `.envrc.local` for custom setups

**Why this pattern:**

- Explicit and traceable - Easy to see where values come from
- No magic - No hidden global state or complex binding frameworks
- Testable - Configuration is just a struct
- Consistent - Same pattern for all config values

## Authentication

### Service Layer Pattern

**Pattern:** Domain services coordinate between generic interfaces.

**Location:** `internal/auth/`

**How it works:**

Services like `auth.Service` define domain concepts (access tokens, refresh tokens, user authentication) and translate them to/from generic storage and provider interfaces.

```go
type Service struct {
    provider OAuthProvider   // Generic OAuth interface
    storage  SecureStorage   // Generic key-value storage
    logger   log.Logger
}
```

**Why this pattern:**

- Services own domain logic, interfaces stay generic
- Easy to swap implementations (WorkOS → another OAuth provider)
- Clear separation between "what" (auth flow) and "how" (WorkOS API calls)

### Secure Storage

**Pattern:** OS-native secure storage via `keyring` package.

**Location:** `internal/keyring/`

**What gets stored:**

- OAuth access tokens
- OAuth refresh tokens

**Why keychain:**

These tokens grant full API access and can be long-lived. macOS Keychain (and equivalent on other platforms) provides proper encryption and protection.

**Clearing tokens:**

```bash
# Clear everything
task reset

# Just clear auth tokens
task reset:auth
```

## TUI Architecture

### Mode-Based Architecture

**Pattern:** Top-level `TUI` model routes between modes (onboarding, app).

**Location:** `internal/tui/`

**How it works:**

The TUI has a `currentMode` that handles all updates and rendering. Modes are self-contained and can transition to other modes.

**Why this pattern:**

- Clear separation between different user experiences
- Each mode can have its own state and lifecycle
- Easy to add new modes without affecting existing ones

## Testing

### Install Script Testing

**Pattern:** Automated testing via `task signoff` before every push.

**Location:** `scripts/install.sh`, `Taskfile.yml`

**How it works:**

1. `task lint:scripts` - Shellcheck validates shell script quality
2. `task install:test` - Downloads and installs from GitHub releases
3. Part of `task signoff` - Runs on every PR via `bot-signoff.yaml`

**Why this pattern:**

- Install script is critical user experience
- Real-world test (actual download from GitHub)
- Catches breaking changes before users see them

## Release Process

### Conventional Commits

**Pattern:** All commits follow conventional commit format.

**Format:** `type(scope): description`

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `chore:` - Maintenance, no user-facing changes
- `docs:` - Documentation changes
- `refactor:` - Code restructuring
- `test:` - Test additions or changes
- `ci:` - CI/CD changes

**Why this pattern:**

- Automated changelog generation via release-please
- Clear commit history
- Semantic versioning based on commit types

### Atomic Commits

**Pattern:** Each commit represents one logical unit of work.

**Examples:**

- ✅ Good: `feat: add install script for single-line installation`
- ✅ Good: `fix: remove log output from get_latest_version function`
- ❌ Bad: `feat: add install script, fix bugs, update docs`

**Why this pattern:**

- Easy to review
- Easy to revert if needed
- Clear history of what changed and why

## Task Management

### Taskfile Organization

**Pattern:** Tasks organized by category with consistent naming.

**Location:** `Taskfile.yml`

**Categories:**

- Build tasks: `build`, `build:version`
- Development: `run`, `dev`, `do`, `signoff`
- Testing: `test`, `lint`, `lint:scripts`
- Release: `release:check`, `release:snapshot`
- Cleanup: `clean`, `reset`, `reset:auth`, `reset:config`
- Installation: `install`, `install:test`

**Naming convention:**

- `category:action` for related tasks (e.g., `reset:auth`, `reset:config`)
- Parent task delegates to sub-tasks (e.g., `reset` calls both `reset:auth` and `reset:config`)

**Why this pattern:**

- Discoverable via `task --list`
- Consistent interface for developers
- Composable - tasks can call other tasks

## Adding New Patterns

As the codebase evolves, add new sections to this document following this structure:

1. **Pattern name and intent** - What problem does it solve?
2. **Location** - Where in the codebase?
3. **How it works** - Brief explanation with code example
4. **Why this pattern** - Rationale and benefits
5. **When to use** - Guidelines for application

Keep examples concrete and reference actual code in the repository.
