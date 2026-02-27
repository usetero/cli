package onboarding

// Gate represents a deterministic onboarding gate/state.
type Gate string

const (
	GatePreflight        Gate = "preflight"
	GateAuthenticate     Gate = "authenticate"
	GateRoleSelect       Gate = "role_select"
	GateOrgSelect        Gate = "org_select"
	GateOrgCreate        Gate = "org_create"
	GateAccountSelect    Gate = "account_select"
	GateAccountCreate    Gate = "account_create"
	GateDatadogCheck     Gate = "datadog_check"
	GateDatadogRegion    Gate = "datadog_region"
	GateDatadogAPIKey    Gate = "datadog_api_key"
	GateDatadogAppKey    Gate = "datadog_app_key"
	GateDatadogDiscovery Gate = "datadog_discovery"
	GateWorkspaceSelect  Gate = "workspace_select"
	GateSync             Gate = "sync"
)

func (g Gate) String() string { return string(g) }
