package onboarding

import "github.com/usetero/cli/internal/core/bootstrap"

// Gate represents a deterministic onboarding gate/state.
type Gate = bootstrap.Gate

const (
	GatePreflight        Gate = "preflight"
	GateAuthenticate     Gate = bootstrap.GateAuthenticate
	GateRoleSelect       Gate = bootstrap.GateRoleSelect
	GateOrgSelect        Gate = bootstrap.GateOrgSelect
	GateOrgCreate        Gate = bootstrap.GateOrgCreate
	GateAccountSelect    Gate = bootstrap.GateAccountSelect
	GateAccountCreate    Gate = bootstrap.GateAccountCreate
	GateRuntimeInit      Gate = bootstrap.GateRuntimeInit
	GateDatadogCheck     Gate = bootstrap.GateDatadogCheck
	GateDatadogRegion    Gate = bootstrap.GateDatadogRegion
	GateDatadogAPIKey    Gate = bootstrap.GateDatadogAPIKey
	GateDatadogAppKey    Gate = bootstrap.GateDatadogAppKey
	GateDatadogDiscovery Gate = bootstrap.GateDatadogDiscovery
	GateWorkspaceSelect  Gate = bootstrap.GateWorkspaceSelect
	GateSync             Gate = bootstrap.GateSync
)
