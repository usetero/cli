# Tero CLI

The terminal interface to the [Tero](https://usetero.com) control plane. Connect a
Datadog account and explore the issues, waste, and posture Tero finds in your
observability data — interactively in your terminal, or as JSON for scripting.

*Built by the creators of [Vector.dev](https://vector.dev).*

Tero analyzes what your logs mean semantically — patterns, quality, and value —
and surfaces the waste that doesn't help you during incidents (typically 40%+ of
log volume). The CLI is read-only: it connects to your existing observability
platform over an API and shows you what Tero found. It deploys no agents,
collectors, or pipelines.

---

## Install

```bash
# Quick install (macOS and Linux)
curl -sSfL https://sh.usetero.com | sh

# Homebrew (macOS and Linux)
brew install usetero/tap/tero

# Scoop (Windows)
scoop bucket add tero https://github.com/usetero/scoop-bucket
scoop install tero

# Docker
docker pull usetero/tero
```

Verify the install:

```bash
tero --version
```

---

The rest of this README follows the [Diátaxis](https://diataxis.fr) structure:

- **[Getting started](#getting-started)** — a guided first run (start here).
- **[How-to guides](#how-to-guides)** — recipes for specific tasks.
- **[Reference](#reference)** — commands, flags, keys, and configuration.
- **[Concepts](#concepts)** — how Tero works and why.

---

## Getting started

A first run takes about five minutes. You'll authenticate, connect a Datadog
account, and land in the issue explorer.

### 1. Launch the app

```bash
tero
```

`tero` with no arguments opens the interactive terminal UI and walks you through
onboarding.

### 2. Authenticate

On first run you'll be prompted to log in (or create an account). This opens a
browser-based device login. Once you're authenticated, the CLI remembers you —
you won't be asked again until you log out.

### 3. Connect a Datadog account

If the selected account has no Datadog connection yet, onboarding walks you
through it:

1. **Pick your Datadog region** — US1, US3, US5, EU1, AP1, or US1-FED.
2. **Enter your Datadog API key** — *Datadog → Organization Settings → API Keys.*
3. **Enter your Datadog Application key** — *Datadog → Organization Settings →
   Application Keys.*

Tero validates the keys, registers the account, and begins analyzing your data.
Access is **read-only**: Tero only reads telemetry metadata.

### 4. Explore your issues

After onboarding you land in the **issue explorer** — a read-only list of the
active issues Tero found, highest priority first. Use:

- `↑` / `↓` (or `k` / `j`) to move through issues
- `r` to refresh
- `ctrl+d` to open the status drawer (Issues, Checks, Services, Log events, Edge)
- `/` to open the command palette (refresh, switch org/account, theme, quit)
- `ctrl+c` to quit

### 5. Or just ask the CLI directly

You don't need the UI to see your account. The same data is available as plain
commands:

```bash
tero status     # account health, services, log events, cost, open issues
tero issues     # active issues
tero services   # enabled services
tero checks     # product checks and posture
```

That's it. Run `tero` anytime to explore, or use the commands above in scripts.

---

## How-to guides

### Authenticate, or check who you are

```bash
tero auth login      # browser device login
tero auth status     # show the current user and org
tero auth logout     # clear stored credentials
```

### Connect (or reconnect) a Datadog account

Connecting Datadog happens interactively inside `tero`. It runs automatically
during onboarding when the active account has no Datadog connection. To connect
a *different* account, run `tero`, press `/`, and choose **Switch Account** —
onboarding re-runs for the account you select.

> There is no headless `tero datadog connect` command yet; connection is
> interactive only.

### See your account status at a glance

```bash
tero status
```

```
Account Status
Health         OK
Ready for use  true
Services       3 active / 17 total
Log events     197 (197 analyzed)
Volume         4.2k/hr
Cost           $74/yr
Open issues    1 (0 high, 1 medium, 0 low)
```

### List your active issues

```bash
tero issues
```

```
PRIORITY  ID     SERVICE     TITLE
medium    ISS-4  accounting  Order receipt logs include full customer shipping addresses
```

### Inspect services, checks, and edge instances

```bash
tero services   # enabled services with health, volume, and cost
tero checks     # product checks with open findings, active issues, and cost
tero edge       # registered edge instances
```

### Get machine-readable output for scripting

Every surface command supports `-o json` (default is `table`):

```bash
tero issues -o json
tero status -o json | jq '.open_issues'
tero services -o json | jq -r '.[] | select(.health != "OK") | .name'
```

```json
[
  {
    "id": "019eaa9e-3242-7ba8-92b1-1b034a4b532d",
    "display_id": "ISS-4",
    "priority": "medium",
    "service": "accounting",
    "title": "Order receipt logs include full customer shipping addresses"
  }
]
```

### Switch organization or account

Inside `tero`, press `/` to open the command palette and choose **Switch
Organization** or **Switch Account**. To switch org from a script:

```bash
tero auth switch <organization>
```

### Start over

```bash
tero reset       # clear stored preferences and authentication for this environment
```

---

## Reference

### Commands

| Command | Description |
|---------|-------------|
| `tero` | Launch the interactive UI (onboarding → issue explorer). |
| `tero status` | Account health, service/event counts, cost, and open-issue summary. |
| `tero issues` | Active issues (priority, ID, service, title). |
| `tero checks` | Product checks with findings, active issues, and cost. |
| `tero services` | Enabled services with health, volume, and cost. |
| `tero edge` | Edge instances registered for the account. |
| `tero auth login` | Authenticate via browser device login. |
| `tero auth status` | Show the current user and organization. |
| `tero auth logout` | Clear stored credentials. |
| `tero auth switch [org]` | Switch the active organization. |
| `tero auth token` | Print the current access token. |
| `tero reset` | Clear preferences and authentication for the current environment. |

### Global flags

| Flag | Description |
|------|-------------|
| `-o, --output <table\|json>` | Output format for surface commands. Default `table`. |
| `--endpoint <url>` | Override the control-plane endpoint. |
| `-d, --debug` | Enable debug logging. |
| `-v, --version` | Print the CLI version. |
| `-h, --help` | Help for any command. |

### Interactive UI keys

| Key | Action |
|-----|--------|
| `↑` / `↓` (or `k` / `j`) | Navigate issues / drawer rows. |
| `r` | Refresh the issue list. |
| `ctrl+d` | Toggle the status drawer (Issues, Checks, Services, Log events, Edge). |
| `tab` | Next tab in the drawer. |
| `esc` | Close the drawer. |
| `/` | Open the command palette. |
| `ctrl+c` | Quit. |

### Environment variables

| Variable | Description |
|----------|-------------|
| `TERO_ENV` | Environment to target: `prd` (default), `dev`, or `local`. |
| `TERO_API_ENDPOINT` | Override the control-plane GraphQL endpoint. |
| `TERO_DEBUG` | Set to `1` or `true` to enable debug logging. |

### Files

Credentials and preferences live under `~/.tero/environments/<env>/`. Logs are
written to `~/.tero/environments/<env>/tero.log`. Run `tero internal inspect
paths` to print the resolved locations.

---

## Concepts

### How it works

The CLI is a thin presentation layer over the Tero **control plane**. It holds no
local database and runs no sync engine: every command reads (or writes) directly
to the control plane over GraphQL. Intelligence — the semantic analysis of your
logs, the issues, the cost modeling — lives server-side. The CLI's job is to
authenticate you, connect your data source, and show you the results.

### What Tero finds

Tero builds a semantic catalog of your log events — what each pattern *means*,
how much it costs, and whether it helps during an incident. From that it surfaces
**issues** (things worth your attention, like leaking PII or high-cost noise) and
runs **checks** across cost and compliance domains. The CLI lets you browse all
of this per account.

### Safety

- **Read-only by default.** Tero reads telemetry metadata to build its catalog.
  It does not store your raw log content.
- **No infrastructure.** No agents, collectors, or pipeline configs — just a
  read-only API connection to your existing platform.
- **Opt-in everything.** Connecting Datadog requires keys you provide; nothing is
  configured without your action.

### What this isn't

- **Not a pipeline.** Tero doesn't route, sample, or transform data in flight. It
  analyzes what you already have and helps you improve it at the source.
- **Not a cost tool.** Reduced spend is a side effect of better data, not the
  goal. Tero explains *why* a pattern is or isn't valuable.

### Supported sources

Datadog today. CloudWatch, Splunk, and others are on the roadmap.

---

## Resources

- **[Documentation](https://tero.com/docs)** — full platform docs and guides
- **[GitHub Issues](https://github.com/usetero/cli/issues)** — bug reports and feature requests
- **[Contact us](https://tero.com/contact)** — questions or feedback
- **[Contributing](CONTRIBUTING.md)** — developer documentation for working on the CLI

## About

Tero is from the creators of [Vector.dev](https://vector.dev) (acquired by
Datadog). We've spent a decade inside enterprise observability systems and seen
this problem from every angle — as engineers, founders, and inside major vendors.

We built Tero because observability data quality is broken and nobody's fixing
it: not the vendors (they profit from waste), not the pipelines (they can't
understand semantic meaning), and not the cost tools (they show you bills, not
solutions). Tero understands what your data means, identifies what's wrong, and
helps you fix it at the source.

---

**Copyright © 2025 Tero Edge, Inc.**
