package auth

import (
	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

func (m *AuthenticateModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *AuthenticateModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Sign in", "Sign in to continue.")
}
