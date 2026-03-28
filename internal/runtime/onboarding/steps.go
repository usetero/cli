package onboarding

// Step is the next onboarding interaction required from the UI.
type Step string

const (
	StepOrganizationSelect Step = "organization.select"
	StepOrganizationCreate Step = "organization.create"
	StepAccountSelect      Step = "account.select"
	StepAccountCreate      Step = "account.create"
	StepWorkspaceSelect    Step = "workspace.select"
	StepDatadogRegion      Step = "datadog.region"
	StepDatadogAPIKey      Step = "datadog.api_key"
	StepDatadogAppKey      Step = "datadog.app_key"
	StepDatadogDiscovery   Step = "datadog.discovery"
	StepPowerSyncReady     Step = "powersync.ready"
	StepDone               Step = "done"
)

func nextStep(state State) Step {
	if state.SelectedOrganization == nil {
		if len(state.Organizations) == 0 {
			return StepOrganizationCreate
		}
		return StepOrganizationSelect
	}
	if state.SelectedAccount == nil {
		if len(state.Accounts) == 0 {
			return StepAccountCreate
		}
		return StepAccountSelect
	}
	if state.SelectedWorkspace == nil {
		return StepWorkspaceSelect
	}
	if state.DatadogAccount == nil {
		if !state.DatadogDraft.Site.Valid() {
			return StepDatadogRegion
		}
		if !state.DatadogDraft.HasAPIKey {
			return StepDatadogAPIKey
		}
		return StepDatadogAppKey
	}
	if state.DatadogStatus == nil || !state.DatadogStatus.ReadyForUse {
		return StepDatadogDiscovery
	}
	if !state.PowerSyncReady {
		return StepPowerSyncReady
	}
	return StepDone
}
