# Tero CLI

Improve your observability data quality from the terminal.

*Built by the creators of [Vector.dev](https://vector.dev).*

## What is this?

Tero helps you find and fix waste in your observability data.

Connect your Datadog account (read-only) and Tero will:
- Understand what your logs mean semantically - patterns, quality, value
- Identify waste - typically 40%+ of volume that doesn't help during incidents
- Help you remove it with informed actions - only what won't hurt you

**If someone sent you here:** Your team lead or SRE found waste in one of your services. The CLI will show you exactly what patterns are wasteful and why, then help you fix them. Takes 10 minutes.

**If you're evaluating Tero:** This is how you interact with the platform. Install it, connect your Datadog account, see what we find. Takes 5 minutes.

## Quick Start

**Install:**

```bash
brew install tero
```

**Run:**

```bash
tero
```

On first run, `tero` will:
1. Ask you to authenticate (or create an account)
2. Walk you through connecting your Datadog account (read-only API key)
3. Analyze your data and show you what it found

After that, just run `tero` anytime to explore waste, check status, or take action.

## What does it do?

`tero` is an interactive chat interface. Ask questions, get answers about your observability data.

The CLI doesn't just identify waste—it teaches you what makes observability data valuable. Each recommendation explains why something is or isn't useful during incidents, helping your team get better at instrumentation over time.

**Common workflows:**

```
"How much waste do I have?"
→ Shows total waste across your account, broken down by service

"What's wrong with checkout-api?"
→ Shows specific waste patterns in that service
→ Explains what each pattern is and why it's waste
→ Lets you take action

"Show me the database_connection_debug logs"
→ Displays actual log samples
→ Explains the pattern and cost impact

"Block those logs"
→ Creates exclusion rule in Datadog
→ Confirms savings
```

**Example session:**

```
$ tero

Welcome back, Ben.

I analyzed your Datadog account. Found $89K/year in waste across 12 services.

Services with the most waste:
  1. checkout-api: $50K/year
  2. user-auth: $24K/year
  3. payment-processor: $15K/year

Want to start with checkout-api?

> yes

Found 24 waste patterns in checkout-api. Here are the top 3:

┌────────────────────────────┬──────────┬──────────────┐
│ Event                      │ Volume   │ Cost Impact  │
├────────────────────────────┼──────────┼──────────────┤
│ middleware_execution_trace │ 2M/hr    │ $26K/year    │
│ database_query_start       │ 800K/hr  │ $11K/year    │
│ cache_operation_success    │ 1M/hr    │ $13K/year    │
└────────────────────────────┴──────────┴──────────────┘

> what is middleware_execution_trace?

This log tracks execution through the middleware chain. Template:

  2024-01-15 10:23:45.123 INFO middleware executing name=cors path=/api/checkout request_id=req_x7k2m
  2024-01-15 10:23:45.125 INFO middleware executing name=auth path=/api/checkout request_id=req_x7k2m
  2024-01-15 10:23:45.127 INFO middleware executing name=rate_limit path=/api/checkout request_id=req_x7k2m
  2024-01-15 10:23:45.129 INFO middleware executing name=request_logger path=/api/checkout request_id=req_x7k2m

Logs 6-8 times per request as it moves through the middleware chain. At 250K requests/hour, that's 2M logs/hour.

Likely added during development to debug middleware ordering. When middleware fails, you get actual error logs. Success logs don't help during incidents.

Safe to drop.

> block it

Created exclusion rule in Datadog. This will save $26K/year immediately.

This log is emitted from src/middleware/logger.ts - want me to help you remove it from the code so it doesn't come back?

> yes

[Scanning your local repository for the logging statement...]
Found in src/middleware/logger.ts:45

[Opens your editor with the change ready to review]

Done. Blocked in Datadog and removed from code. Want to see the next waste pattern?
```

## Safety

**Read-only access:** Tero only reads data from Datadog. We never write or modify anything without your explicit confirmation.

**No pipeline required:** We're not a data pipeline or routing tool. Tero connects via API to your existing observability platforms - no new infrastructure to deploy or manage.

**No infrastructure changes:** No agents to install. No collector configs to update. No deployment required. Just a read-only API connection.

**Opt-in actions:** When you choose to block waste, we configure your existing tools (Datadog exclusion rules, code changes, etc.). Everything is reversible.

## What This Isn't

**Not a cost-cutting tool.** Tero helps you improve observability quality. Reduced costs are a side effect of better data.

**Not a pipeline.** We don't route, sample, or transform your data in flight. We analyze what you have and help you improve it at the source.

**Not automatic.** We never drop data without your explicit approval. You're in control of every action.

## Common Questions

**What Datadog permissions does Tero need?**

Read-only access to start. Tero will request specific write permissions (like creating exclusion rules) only when you choose to take action. See [setup guide](https://tero.com/docs/setup) for details.

**What data does Tero collect?**

We analyze metadata about your telemetry (log event names, volumes, services, costs) to build our semantic catalog. We don't store your actual log content. See [Privacy Policy](https://tero.com/privacy).

**Does this work with other observability tools?**

Datadog only right now. CloudWatch, Splunk, and others coming soon.

**More questions?**

See our [full documentation](https://tero.com/docs) or [contact us](https://tero.com/contact).

## Resources

- **[Documentation](https://tero.com/docs)** - Full platform docs and guides
- **[GitHub Issues](https://github.com/usetero/cli/issues)** - Bug reports and feature requests
- **[Contact Us](https://tero.com/contact)** - Questions or feedback
- **[Contributing](CONTRIBUTING.md)** - Developer documentation for working on the CLI

## About

Tero is from the creators of [Vector.dev](https://vector.dev) (acquired by Datadog). We've spent a decade inside enterprise observability systems and seen this problem from every angle - as engineers, founders, and inside major vendors.

We built Tero because observability data quality is broken and nobody's fixing it. Not the vendors (they profit from waste), not the pipelines (they can't understand semantic meaning), and not the cost tools (they show you bills, not solutions).

Tero is different. We understand what your data means, identify what's wrong, and help you fix it - the right way.

---

**Copyright © 2025 Tero, Inc.**
