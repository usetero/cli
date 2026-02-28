package organizations

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *SelectModel) Hidden() bool { return false }

func (m *SelectModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Select organization", Details: "Select an organization to continue."}
}

func (m *CreateModel) Hidden() bool { return false }

func (m *CreateModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Create organization", Details: "Create an organization to continue."}
}
