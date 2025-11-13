# Design

## Introduction

The Tero CLI is the presentation layer for Tero. While the control plane does the hard work of analyzing your observability data, discovering waste, and understanding quality—the CLI's job is to make that intelligence accessible, understandable, and actionable.

This document explains who we're building for, the principles that guide our design decisions, and the patterns you'll see throughout the codebase. If you're contributing features or improvements, this is your guide to making decisions that feel consistent with the rest of the CLI.

---

## Who We're Building For

The CLI serves two distinct personas with different goals and workflows. Understanding both helps you design features that work for the right audience.

### Engineering Leadership

These are VPs of Engineering, SRE leads, and platform team managers. They're responsible for observability budgets, data quality across teams, and organizational efficiency.

**What they care about:**
- Org-wide visibility: which services are producing quality telemetry, which aren't
- Cost and waste metrics: where money is being spent, where it's being wasted
- Team accountability: which teams need to improve, how to help them
- Progress toward goals: waste reduction targets, cost objectives, quality SLOs

**What they need from the CLI:**
- High-level overview that shows patterns across the organization
- Ability to drill into problem areas without getting lost in details
- Tools to communicate with teams about improvements (without micromanaging)
- Proof that observability spend is justified and improving

**What they don't want:**
- To manually chase down every team about their telemetry
- To be seen as "cost police" cutting corners at engineering's expense
- Surprises when quality rules get enforced

### Engineers

These are the people building and maintaining services. They own the code, the instrumentation, and the day-to-day operations.

**What they care about:**
- Their specific services: understanding what telemetry they're producing
- Actionable guidance: not "your logs are expensive" but "here's exactly what to fix"
- Learning: what makes observability data good vs wasteful
- Meeting expectations: hitting the quality goals their leadership set

**What they need from the CLI:**
- Service-specific insights focused on what they own
- Concrete actions they can take to improve quality
- Examples and explanations that teach observability best practices
- Control over changes (no surprises, no automatic drops without permission)

**What they don't want:**
- Vague complaints about cost or quality
- Their valuable debugging data dropped without warning
- Another dashboard to check or tool to learn

---

## Design Principles

These principles guide every interaction, every feature, and every UX decision in the CLI.

### Conversational Over Transactional

Traditional CLIs work in transactions: you run a command with flags, you get output, done. Tero is conversational—you have an ongoing dialogue that builds context and understanding.

**Not this:**
```
$ tero service list --filter waste --sort cost --threshold 1000
$ tero service show checkout-api --metrics waste --breakdown events
$ tero logs block --service checkout-api --event debug_log
```

**Instead this:**
```
> show me services with high waste

checkout-api: 45% waste ($15K/month)
payment-api: 32% waste ($8K/month)
...

> tell me about checkout-api

checkout-api has 3 major waste patterns:
• debug_log: 2M/hr, $8K/month
• health_check: 1M/hr, $3K/month
...

> block the debug logs

Done! Created Datadog exclusion rule.
Savings: $8K/month
```

**Why:** People think in questions, not command-line flags. Natural language lets us understand intent and maintain context across multiple interactions. The CLI remembers what you're talking about—you don't have to repeat yourself.

### Progressive Disclosure

Don't overwhelm users with everything at once. Start with high-level insights and let them drill down into details on request.

**The pattern:**
- **Level 1:** Summary (the headline)
- **Level 2:** Breakdown (what's driving this?)
- **Level 3:** Examples (show me actual data)

**Example flow:**
```
> how's checkout-api doing?

checkout-api: 45% waste ($15K/month)

> what's causing the waste?

Top 3 patterns:
• debug_log: 2M/hr, $8K/month
• health_check: 1M/hr, $3K/month
• stack_trace: 500K/hr, $4K/month

> show me examples of debug_log

[Syntax-highlighted log samples with context]
```

**Why:** Cognitive load matters. Show people the signal first, let them choose to see the noise. Most users just need the summary—power users can dive deep.

### Action-Oriented

Analysis without action is frustrating. Every insight should lead to "What can I do about this?"

**Not enough:**
```
checkout-api has 2M debug logs per hour costing $8K/month.
```

**Better:**
```
checkout-api has 2M debug logs per hour costing $8K/month.

What would you like to do?
[a] Block these logs in Datadog (immediate savings)
[b] Help me remove from code (permanent fix)
[c] Show me examples first
```

**Why:** People use this CLI to improve their observability, not just understand it. Make the path from insight to action as short as possible. Always offer options—never dead-end with information.

### Role-Based Experience

The same CLI adapts to who you are. Leadership and engineers use the same tool but get experiences tailored to their needs.

**Leadership sees:**
- Org-wide metrics and trends
- Team/service breakdowns
- Progress toward organizational goals
- Tools for communicating with teams

**Engineers see:**
- Their specific services and ownership
- Detailed technical recommendations
- Hands-on actions (fix code, configure tools)
- Learning resources (why is this waste?)

**How we adapt:**
- During onboarding, we ask about role and responsibilities
- Leadership users automatically get org-wide views
- Engineer users see only their services by default
- Both can switch contexts when needed

**Why:** Different roles have different jobs. Don't make engineers wade through org charts, don't make leaders debug individual log lines. Respect their time and focus.

### Contextual and Continuous

The CLI maintains context as you work. Whether you're in an interactive conversation or making tool calls through the MCP server, the system understands what you're doing and builds on it.

**Not this:** Every interaction starts from scratch, no memory
```
$ tero service show checkout-api
$ tero service show checkout-api --metrics waste
$ tero logs block --service checkout-api --event debug_log
```

**Instead this:** Context flows naturally
```
> show me checkout-api

[Details about checkout-api]

> what's the waste breakdown?

[Already knows we're talking about checkout-api]

> block the debug logs

[Still in checkout-api context, knows which service]
```

**Why:** Complex problems need exploration, not one-shot queries. Maintaining context makes the experience natural and efficient. The control plane manages conversation history—the CLI just presents the current state beautifully.

---

## Interaction Patterns

These are the common flows and patterns you'll see throughout the CLI. Understanding them helps you design features that feel consistent.

### First Run (Onboarding)

The first impression matters. Onboarding should be quick, personalized, and immediately useful.

**Principles for onboarding:**
- **Low friction:** Don't ask 20 questions before showing value
- **Personalized:** Adapt to role (leadership vs engineer)
- **Immediate value:** Show insights, not tutorials or empty states

**The flow:**
1. Collect email (used to find or create organization)
2. Authenticate (WorkOS handles SSO if configured)
3. Ask about role (leadership or engineer)
4. If engineer: which services do you work on?
5. Show initial insights based on their role

**Why start with insights:** People stay engaged when they see value. Don't make them configure everything before showing them anything useful.

### Typical Sessions

Different personas follow different patterns. Design features that support their natural workflows.

**Leadership flow:**
```
1. Overview → How's the org doing?
2. Identify problems → Which services/teams need attention?
3. Drill into specifics → What exactly is wrong?
4. Understand ownership → Which team owns this?
5. Track progress → Is it getting better?
```

Leadership users think top-down: organization → team → service → problem.

**Engineer flow:**
```
1. Service status → How are my services doing?
2. Identify issues → What specific patterns are wasteful?
3. Understand why → Show me examples, explain the problem
4. Take action → Block logs, fix code, configure rules
5. Learn → Why is this considered waste?
```

Engineer users think bottom-up: my service → specific problem → fix it → learn from it.

### Taking Action

Actions come in three flavors with increasing permanence. Always offer options and let users choose.

**1. Block in vendor tool** (Quick win)
- Creates exclusion rule in Datadog/Splunk/etc.
- Immediate savings, code unchanged
- Reversible if needed

**2. Fix in code** (Permanent solution)
- Scans codebase, finds log statements
- Generates PR to remove/improve instrumentation
- Requires review and merge

**3. Automate** (High-confidence only)
- User enables automation for certain patterns
- CLI only acts on 100% certainty (debug logs in production, health check spam)
- Always reviewable after the fact

**Key principle:** Never surprise users. Make it clear what will happen, get confirmation for destructive actions, allow review and rollback.

---

These principles apply across all CLI interfaces—whether you're using the interactive TUI, integrating through MCP, or running traditional commands. The presentation changes, but the design philosophy stays the same: make Tero's intelligence accessible, understandable, and actionable.
