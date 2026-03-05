package preflight

import (
	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

func (m *Model) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *Model) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Getting ready", "Preparing your account...")
}
