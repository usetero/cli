# Contributing to Tero CLI

## CRITICAL RULES FOR AI ASSISTANTS

### DESTRUCTIVE COMMANDS REQUIRE USER APPROVAL

**NEVER execute these without asking the user first:**
- `rm -rf` - Recursive directory deletion
- `rm` with multiple files or wildcards
- `mv` that overwrites existing files
- `sed -i` - In-place file editing
- ANY command that deletes or modifies multiple files

**SAFE OPERATIONS (no approval needed):**
- `mcp__acp__Read` - Read files
- `mcp__acp__Write` - Write individual files (with safety checks)
- `mcp__acp__Edit` - Edit individual files (exact string replacement)
- `Bash` for read-only operations (ls, find, grep)

**WORKFLOW FOR DESTRUCTIVE OPERATIONS:**
1. Describe what you want to do and why
2. Show the exact command you would run
3. ASK: "Should I run this command?"
4. Wait for explicit approval
5. ONLY then execute if approved

**Example:**
"I need to remove the old onboarding orchestrator files. I would run:
```
rm internal/tui/pages/onboarding/onboarding.go internal/tui/pages/onboarding/messages.go
```
Should I run this command?"

**This is non-negotiable. Violating this rule wastes hours of work.**

---

## Introduction

Tero is a control plane for observability. It sits on top of your existing tools (Datadog, Splunk, CloudWatch, etc.), understands what your telemetry means semantically, and helps you improve it—identify waste, take action, reduce cost.

The CLI is how users interact with Tero. It's a presentation layer that makes the control plane's intelligence accessible through multiple interfaces—an interactive TUI for conversational exploration, an MCP server for coding agents, and eventually traditional commands for scripting and automation.

This guide helps you get set up and productive working on the CLI.

---

## Getting Started

### Prerequisites

The project uses [Hermit](https://cashapp.github.io/hermit/) for toolchain dependencies. When you enter the directory, Hermit automatically activates the right versions of Go, Node, and other tools.

If you don't have Hermit installed, ensure you have:
- Go 1.21+
- Node.js (for GraphQL schema fetching)

### First Build

```bash
git clone https://github.com/usetero/cli
cd cli
task build
```

This builds the `tero` binary to `bin/tero`.

### Run It

```bash
task run
```

This runs the CLI directly without building a binary first—useful for fast iteration during development.

---

## Development Workflow

### The Commands You'll Use

**`task build`** - Build the CLI binary

**`task run`** - Run without building (fast iteration)

**`task test`** - Run tests

**`task do`** - Format, lint, and test (run this before committing)

**`task client:generate`** - Regenerate GraphQL client from control plane schema

**`task dev`** - Generate client + build (full dev cycle)

### Typical Workflows

**Working on a feature:**
```bash
# Make your changes
task run      # Test them
task test     # Make sure tests pass
task do       # Format, lint, test before committing
```

**When the control plane GraphQL schema changes:**
```bash
# 1. Make sure control plane is running at localhost:8081
# 2. Regenerate the GraphQL client
task client:generate

# 3. Build and test
task dev
```

**Before submitting a PR:**
```bash
task do  # This formats, lints, and tests
```

---

## Understanding the Codebase

### Project Structure

```
cmd/tero/          # Entry point, Cobra commands
internal/
  tui/             # All TUI code
    page/          # Pages (onboarding, chat)
    components/    # Reusable components
  client/          # Generated GraphQL client (never edit directly)
  auth/            # WorkOS authentication
  mcp/             # MCP server implementation
  config/          # CLI configuration
docs/              # Architecture and design documentation
```

### Where to Learn

The code is the source of truth for implementation patterns. The architecture docs explain how everything fits together and why we make the decisions we do.

**Read these to understand the architecture:**

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - How the CLI is architected, how it communicates with the control plane, key architectural principles

- **[docs/DESIGN.md](docs/DESIGN.md)** - Design philosophy, UX principles, who we're building for and why

- **[docs/TUI.md](docs/TUI.md)** - TUI-specific architecture, component patterns, layout management

These docs build your mental model. The code shows you the patterns in practice. Look at existing pages, components, and features to see how things are done.

---

## Contributing

### Code Quality

Run `task do` before committing. This formats your code, runs the linter, and runs tests. If `task do` passes, you're good.

### Follow Existing Patterns

The codebase has established patterns for components, pages, layout, state management, and more. Look at how existing code works and follow those patterns. The architecture docs explain the principles behind these patterns.

### Key Principles

These are architectural fundamentals—read the architecture docs for details:

- **CLI is presentation only** - Never implement intelligence in the CLI. That belongs in the control plane.
- **Control plane is source of truth** - The CLI caches for performance but the control plane owns all data.
- **Components render, pages decide** - Clear separation of concerns.
- **GraphQL for all communication** - Generated client, never edited manually.

### Pull Requests

- Keep PRs focused on one thing
- Make sure `task do` passes
- Include tests for new functionality
- Reference the architecture docs if you're establishing new patterns

---

## Getting Help

- **Issues:** https://github.com/usetero/cli/issues
- **Questions:** https://github.com/usetero/cli/discussions
- **Architecture questions:** Read the docs in `docs/` first, then ask in discussions

---

Thanks for contributing to Tero!
