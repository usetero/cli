# Logging

Logs tell the story of what the code did. When an engineer reads the logs, they should understand what happened without scratching their head or drowning in noise.

## The Story

Every file has a story to tell. An auth page: "User provided email, it was saved, auth completed." An organization selection step: "Started loading organizations, got 3 results, user selected one." A Datadog key validation: "User entered key, validating it, validation succeeded."

Think of logs as a narrative. The key events are chapter markers—what happened and when. The supporting details are footnotes—useful context when you need to understand how or why. The routine operations don't appear at all—they'd just be noise.

When you run `tero` with good logging, you see:

```
INFO  email saved email=ben@company.com
INFO  auth completed email=ben@company.com
INFO  page transition from=*auth.Model to=*onboarding.Page
INFO  role selected role=engineer
INFO  loading organizations
INFO  organizations loaded count=1
INFO  organization selected id=org_123 name=Acme
INFO  account created id=acc_456 name=Production
INFO  checking datadog account accountID=acc_456
INFO  datadog account found site=us1
INFO  page transition from=*steps.CheckDatadogAccountStep to=*chat.model
```

You can follow the journey. Key events are visible. The noise is gone. That's what we're aiming for.

## Levels Tell Different Parts of the Story

**Info** is the main narrative. Key events, state changes, user actions. This is what you see by default. An engineer should be able to follow the entire user journey from Info logs alone.

**Debug** is supporting detail. Variable values, branches taken, why we made certain decisions. Only visible when explicitly enabled, so it can be verbose where it helps understand flow. Use it for things like "Auto-selected organization (only one available)" or "Loaded preferences from cache."

**Warn** is when something unexpected happened but we recovered. Degraded experience, fallback behavior, things that worked but not how we expected. "Config file unreadable, using defaults."

**Error** is when something actually failed. User-visible problems, data we couldn't save, operations that didn't work. Always include enough context to debug—what failed and why.

## Writing Good Logs

**Follow Go conventions.** Lowercase messages, no ending punctuation. This matches Go's error message style and keeps logs consistent:

```go
logger.Info("organization selected", "id", orgID, "name", name)
logger.Error("failed to save email", "error", err)
```

Not: "Organization selected." or "Failed to save email!" or "Saved the email successfully."

**Tell what happened, not what will happen.** Past tense for completed actions: "window resized", "user authenticated", "organization created." When you do need to log before an operation (because it might take time), use present continuous: "loading organizations", "validating API key."

**Keep style consistent.** Subject first, concise, no fluff. "page transition" not "successfully transitioning to the next page." "email saved" not "we have saved the user's email address."

**Include relevant context.** Just enough to understand what happened:

```go
logger.Info("organization selected", "id", orgID, "name", name)
logger.Error("API call failed", "endpoint", "/graphql", "error", err)
```

Not too much (timestamps, function names, stack traces for non-errors). Not too little (just "organization selected" with no ID). The Goldilocks amount.

**Use explicit attribute types.** Use `slog.String()`, `slog.Int()`, etc. when possible to avoid mismatched key-value pairs:

```go
logger.Info("organizations loaded", slog.Int("count", len(orgs)))
logger.Debug("using cached preferences", slog.String("orgID", id))
```

## Where to Add Logs

**Pages** log lifecycle and state changes. When created, when completed, when state changes (email saved, preferences updated). Not rendering (happens every frame) or size updates (parent already logs dimensions).

**Steps** log completion, remote operations, and errors. "Role selected", "Loading organizations", "Organizations loaded", "Organization created", "Failed to create account." Auto-selection logic goes at Debug level since it's supporting detail.

**Components** mostly stay quiet. Simple rendering components (logo, header, footer) need no logs—they're pure presentation. Only log if a component does complex work like remote calls (remotelist) or has interesting state machines.

**Orchestration** already handled. TUI logs page transitions and window resizing. Flow logs step initialization. You generally don't need to add more here.

## What Not to Log

**Don't log rendering.** View() runs every frame—logging it creates useless noise.

**Don't log routine operations.** SetSize(), Update() for simple key presses, methods that just return values. These happen constantly and tell no story.

**Don't log what the parent already logged.** If the TUI logs "Page transition from=auth to=onboarding", the onboarding page doesn't need to log "Onboarding starting." Redundant.

**Don't log function entry/exit.** "Entering selectOrg" and "Exiting selectOrg" add no value. Just log the meaningful event: "Organization selected."

## Auditing a File

When reviewing a file for logging quality:

1. **Read the entire file.** Understand what it does, what its responsibilities are.

2. **Identify the story.** What are the 3-5 key moments? For an auth page: created, email saved, completed. For a list step: started loading, finished loading, item selected. For a creation step: started creating, created successfully (or failed).

3. **Check existing logs.** Are they at the right level? Do they tell the story? Are they consistent in style?

4. **Remove noise.** Function entry/exit, rendering logs, redundant logs that parents already emit.

5. **Add missing chapters.** Key events that should be visible but aren't logged.

6. **Make it consistent.** Past tense, concise, proper context. Match the style across the codebase.

After the audit, running the application should produce a clean narrative that any engineer can follow without prior knowledge of the code.
