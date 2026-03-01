# Toasts

Toasts are the user-facing status channel in the TUI.
They are not a substitute for logs; they are a product-facing feedback
mechanism for actions the user should notice.

## When a toast is the right tool

Use a toast when the user needs immediate feedback about an operation outcome:

- something failed and they need to know,
- something succeeded and confirmation matters,
- a warning needs acknowledgment.

If the information is only for engineers or does not affect user decision
making, log it instead.
Toasts should answer "what happened and what should I do next?" quickly.

## Message types and where they live

Toast message contracts are defined in `internal/app/events`.
The main types are `Error`, `Warning`, `Success`, and `Info`.

These are emitted as commands and handled by the toast model in the root app
message flow.

## Sticky behavior guidance

Sticky toasts are for blocking or high-friction failures where silent dismissal
would hide an unresolved state.

Non-sticky toasts are for transient confirmations and informational updates.

If you are unsure, default to non-sticky first and escalate only when user
action is truly blocked.
Overusing sticky toasts makes the app feel noisy and can hide the genuinely
critical interruptions.
