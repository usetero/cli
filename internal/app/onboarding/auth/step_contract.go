package auth

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *AuthenticateModel) Hidden() bool { return false }

func (m *AuthenticateModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Sign in", Details: "Authenticate to continue."}
}
