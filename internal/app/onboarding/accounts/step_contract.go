package accounts

import (
	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

func (m *SelectModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *SelectModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Select account", "Select an account to continue.")
}

func (m *CreateModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *CreateModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Create account", "Create an account to continue.")
}
