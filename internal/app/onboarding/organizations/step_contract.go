package organizations

import (
	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

func (m *SelectModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *SelectModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Select organization", "Select an organization to continue.")
}

func (m *CreateModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *CreateModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Create organization", "Create an organization to continue.")
}
