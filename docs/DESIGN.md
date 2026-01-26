# Design

Who we're building for, the principles that guide decisions, and the patterns you'll see throughout the CLI.

If you're contributing features or improvements, this is your guide to making decisions that feel consistent with the rest of the product. For how the CLI is structured and built, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Who We're Building For

The CLI serves two personas with different goals and workflows. Understanding both helps you design features that work for the right audience.

### Engineering Leadership

VPs of Engineering, SRE leads, platform team managers. They're responsible for observability budgets, data quality across teams, and organizational efficiency.

They care about org-wide visibility—which services produce quality telemetry, which don't. Cost and waste metrics. Team accountability. Progress toward goals. They need high-level overviews that show patterns across the organization, with the ability to drill into problem areas without getting lost in details.

They don't want to manually chase down every team about their telemetry, be seen as "cost police" cutting corners, or get surprised when quality rules take effect.

### Engineers

The people building and maintaining services. They own the code, the instrumentation, and the day-to-day operations.

They care about their specific services and what telemetry they're producing. They want actionable guidance—not "your logs are expensive" but "here's exactly what to fix." They want to learn what makes observability data good versus wasteful, and they want to meet the quality expectations their leadership set.

They don't want vague complaints about cost, their valuable debugging data dropped without warning, or another dashboard to check.

---

## Design Principles

### Chat First

Chat is the canvas. Everything happens through or from chat. Users start there, navigate deeper into focused views, and always escape back. This isn't a dashboard with a chat widget—it's a conversational application where chat is the primary navigation mechanism.

This means every feature you build should work naturally from chat. If a user can't discover it or reach it through conversation, it's not well integrated. Pages, views, and visualizations all exist as things chat can surface. Slash commands and @ references are shortcuts, not separate modes.

### Progressive Disclosure

Don't overwhelm users with everything at once. Start with high-level insights and let them drill down on request.

The pattern: summary first (the headline), breakdown on request (what's driving this?), examples on demand (show me actual data). Most users just need the summary. Power users can dive deep. Cognitive load matters—show people the signal first, let them choose to see the noise.

This extends to visualizations. Inline in chat, a visualization is compact—enough to understand the shape of the answer. Expanded to full size, it has all the controls and interactivity. Two levels of detail, same data.

### Action-Oriented

Analysis without action is frustrating. Every insight should lead to "what can I do about this?"

When the AI surfaces a problem—waste in a service, a policy violation, a quality issue—the next step should be obvious. Offer options. Block the logs in Datadog for immediate savings, or fix the instrumentation in code for a permanent solution, or just show examples first. Never dead-end with information. The AI has the same action surface as the user—it can do anything the user can do through the same interfaces.

### Role-Adapted

The same CLI adapts to who you are. During onboarding, we ask about role. Leadership users get org-wide views by default. Engineers see their specific services. Both can switch contexts when needed, but the defaults respect their workflow.

Leadership thinks top-down: organization, team, service, problem. Engineers think bottom-up: my service, specific issue, fix it, learn from it. Design features that support these natural directions instead of forcing one pattern on everyone.

### Context Builds

The CLI maintains context as you work. Reference a service and it becomes active context. Pull in a log event, create a visualization—they accumulate. The AI sees all of it. You don't repeat yourself.

This extends across sessions. When you @ a previous session, its context comes with it—a summary and the key entities, not the full transcript. Context compounds. The more you use Tero, the richer the connections between entities become.

---

## Interaction Patterns

### First Run

The first impression matters. Onboarding should be quick, personalized, and lead to value fast.

The flow: collect email, authenticate, ask about role, set up the organization and integrations. Then drop the user into chat with enough context to be immediately useful. Don't make them configure everything before showing them anything. People stay engaged when they see value.

### Typical Sessions

Different personas follow different patterns.

Leadership starts broad: how's the org doing? Which services need attention? What's causing the problem? Who owns it? Is it getting better? They navigate from organization down to specifics.

Engineers start narrow: how are my services doing? What specific patterns are wasteful? Show me examples. Let me fix it. Why is this considered waste? They navigate from their service outward to understanding.

Both patterns work naturally in chat. The AI adapts its responses to the user's role and the direction of their exploration.

### Navigation

Chat is home. Slash commands jump to specific pages (`/services`, `/policies`). @ references bring entities into context. Expanding a visualization opens a full-size modal. Escape goes back. You can navigate as deep as you want and always return to where you were.

The key constraint: navigating deeper should never feel like you've lost your chat. Modals and overlays preserve the sense of place. Hard page transitions create mental friction—the user wonders "where am I? how do I get back?" We avoid that.

### Taking Action

Actions come in three flavors with increasing permanence.

Block in the vendor tool—quick win, immediate savings, reversible. Fix in code—permanent solution, requires review and merge. Automate—for high-confidence patterns only, always reviewable after the fact.

Never surprise users. Make it clear what will happen, get confirmation for destructive actions, allow review and rollback.

---

These principles apply across all CLI interfaces—the TUI, the MCP server, and future traditional commands. The presentation changes, but the design philosophy stays the same.
