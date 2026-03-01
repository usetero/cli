package preflight

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

func (m *Model) Hidden() bool { return false }

func (m *Model) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Getting ready", Details: "Preparing your account..."}
}
