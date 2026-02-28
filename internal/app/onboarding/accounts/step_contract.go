package accounts

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *SelectModel) Hidden() bool { return false }

func (m *SelectModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Select account", Details: "Select an account to continue."}
}

func (m *CreateModel) Hidden() bool { return false }

func (m *CreateModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Create account", Details: "Create an account to continue."}
}
