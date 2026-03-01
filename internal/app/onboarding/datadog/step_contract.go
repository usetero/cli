package datadog

import onbstatus "github.com/usetero/cli/internal/app/onboarding/status"

func (m *RegionModel) Hidden() bool { return false }

func (m *RegionModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Datadog setup", Details: "Choose your Datadog site."}
}

func (m *APIKeyModel) Hidden() bool { return false }

func (m *APIKeyModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog API key."}
}

func (m *AppKeyModel) Hidden() bool { return false }

func (m *AppKeyModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog application key."}
}

func (m *DiscoveryModel) Hidden() bool { return false }

func (m *DiscoveryModel) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{Title: "Datadog setup", Details: "Discovering Datadog services..."}
}
