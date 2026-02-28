package role

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *Model) Hidden() bool { return false }

func (m *Model) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Select role", Details: "Select your role to continue."}
}
