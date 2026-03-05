package datadog

import (
	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/app/onboarding/stepkit"
)

func (m *RegionModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *RegionModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Datadog setup", "Choose your Datadog site.")
}

func (m *APIKeyModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *APIKeyModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Datadog setup", "Enter your Datadog API key.")
}

func (m *AppKeyModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *AppKeyModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Datadog setup", "Enter your Datadog application key.")
}

func (m *DiscoveryModel) Hidden() bool { return stepkit.AlwaysVisible() }

func (m *DiscoveryModel) Status() onbstatus.StepStatus {
	return stepkit.StaticStatus("Datadog setup", "Discovering Datadog services...")
}
