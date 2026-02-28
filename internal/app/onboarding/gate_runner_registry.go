package onboarding

import (
	"github.com/usetero/cli/internal/app/onboarding/accounts"
	"github.com/usetero/cli/internal/app/onboarding/auth"
	"github.com/usetero/cli/internal/app/onboarding/datadog"
	"github.com/usetero/cli/internal/app/onboarding/organizations"
	"github.com/usetero/cli/internal/app/onboarding/preflight"
	"github.com/usetero/cli/internal/app/onboarding/role"
	"github.com/usetero/cli/internal/app/onboarding/runtimeinit"
	"github.com/usetero/cli/internal/app/onboarding/sync"
	"github.com/usetero/cli/internal/app/onboarding/workspaces"
)

type gateDefinition struct {
	runner      GateRunner
	requirement gateRequirement
	display     gateDisplayPolicy
}

func (m *Model) definitionForGate(gate Gate) gateDefinition {
	if def, ok := m.definitions[gate]; ok {
		return def
	}
	panic("unsupported onboarding gate: " + gate.String())
}

func defaultGateDefinitions() map[Gate]gateDefinition {
	return map[Gate]gateDefinition{
		GatePreflight: {
			runner: gateRunnerFunc{
				gate: GatePreflight,
				new: func(m *Model) Step {
					return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
				},
			},
		},
		GateAuthenticate: {
			runner: gateRunnerFunc{
				gate: GateAuthenticate,
				new: func(m *Model) Step {
					return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
				},
			},
		},
		GateRoleSelect: {
			runner: gateRunnerFunc{
				gate: GateRoleSelect,
				new: func(m *Model) Step {
					return role.New(m.theme, m.userPrefs, m.scope)
				},
			},
		},
		GateOrgSelect: {
			runner: gateRunnerFunc{
				gate: GateOrgSelect,
				new: func(m *Model) Step {
					return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
				},
			},
		},
		GateOrgCreate: {
			runner: gateRunnerFunc{
				gate: GateOrgCreate,
				new: func(m *Model) Step {
					return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
				},
			},
		},
		GateAccountSelect: {
			runner: gateRunnerFunc{
				gate: GateAccountSelect,
				new: func(m *Model) Step {
					return accounts.NewSelect(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true},
		},
		GateAccountCreate: {
			runner: gateRunnerFunc{
				gate: GateAccountCreate,
				new: func(m *Model) Step {
					return accounts.NewCreate(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true},
		},
		GateRuntimeInit: {
			runner: gateRunnerFunc{
				gate: GateRuntimeInit,
				new: func(m *Model) Step {
					return runtimeinit.New(m.theme, *m.state.org, *m.state.account, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true},
			display:     gateDisplayPolicy{hidden: true, status: "Initializing account runtime..."},
		},
		GateDatadogCheck: {
			runner: gateRunnerFunc{
				gate: GateDatadogCheck,
				new: func(m *Model) Step {
					return datadog.NewCheck(m.ctx, m.theme, *m.state.account, m.services, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true},
		},
		GateDatadogRegion: {
			runner: gateRunnerFunc{
				gate: GateDatadogRegion,
				new: func(m *Model) Step {
					return datadog.NewRegion(m.theme, m.scope)
				},
			},
		},
		GateDatadogAPIKey: {
			runner: gateRunnerFunc{
				gate: GateDatadogAPIKey,
				new: func(m *Model) Step {
					return datadog.NewAPIKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.services, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true, needsDDSite: true},
		},
		GateDatadogAppKey: {
			runner: gateRunnerFunc{
				gate: GateDatadogAppKey,
				new: func(m *Model) Step {
					return datadog.NewAppKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.state.ddAPIKey, m.services, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true, needsDDSite: true, needsDDAPIKey: true},
		},
		GateDatadogDiscovery: {
			runner: gateRunnerFunc{
				gate: GateDatadogDiscovery,
				new: func(m *Model) Step {
					return datadog.NewDiscovery(m.ctx, m.theme, m.state.ddAccount, m.services, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true, needsDDAccount: true},
		},
		GateWorkspaceSelect: {
			runner: gateRunnerFunc{
				gate: GateWorkspaceSelect,
				new: func(m *Model) Step {
					return workspaces.NewSelect(m.ctx, m.theme, *m.state.account, m.services, m.orgPrefs, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true},
		},
		GateSync: {
			runner: gateRunnerFunc{
				gate: GateSync,
				new: func(m *Model) Step {
					return sync.New(m.theme, m.syncer, m.scope)
				},
			},
			requirement: gateRequirement{needsOrg: true, needsAccount: true, needsWorkspace: true},
		},
	}
}
