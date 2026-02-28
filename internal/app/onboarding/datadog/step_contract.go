package datadog

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *RegionModel) Hidden() bool { return false }

func (m *RegionModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Datadog setup", Details: "Choose your Datadog site."}
}

func (m *APIKeyModel) Hidden() bool { return false }

func (m *APIKeyModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog API key."}
}

func (m *AppKeyModel) Hidden() bool { return false }

func (m *AppKeyModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog application key."}
}

func (m *DiscoveryModel) Hidden() bool { return false }

func (m *DiscoveryModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Datadog setup", Details: "Discovering Datadog services..."}
}
