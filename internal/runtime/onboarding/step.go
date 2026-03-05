package onboarding

// Step is the next onboarding interaction required from the UI.
type Step string

const (
	StepRoleSelect         Step = "role.select"
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
