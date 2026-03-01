package accounts

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

func (m *SelectModel) Hidden() bool { return false }

func (m *SelectModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Select account", Details: "Select an account to continue."}
}

func (m *CreateModel) Hidden() bool { return false }

func (m *CreateModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Create account", Details: "Create an account to continue."}
}
