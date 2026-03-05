package stepkit

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

// AlwaysVisible is the default Hidden() implementation for non-transient steps.
func AlwaysVisible() bool { return false }

// StaticStatus returns a fixed onboarding step status payload.
func StaticStatus(title, details string) onbstatus.StepStatus {
	return onbstatus.StepStatus{
		Title:   title,
		Details: details,
	}
}
