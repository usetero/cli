package auth

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

func (m *AuthenticateModel) Hidden() bool { return false }

func (m *AuthenticateModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Sign in", Details: "Sign in to continue."}
}
