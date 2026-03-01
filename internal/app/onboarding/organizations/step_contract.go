package organizations

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

func (m *SelectModel) Hidden() bool { return false }

func (m *SelectModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Select organization", Details: "Select an organization to continue."}
}

func (m *CreateModel) Hidden() bool { return false }

func (m *CreateModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Create organization", Details: "Create an organization to continue."}
}
