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

func gateDef(gate Gate, newStep func(m *Model) Step, opts ...func(*gateDefinition)) gateDefinition {
	def := gateDefinition{
		runner: gateRunnerFunc{
			gate: gate,
			new:  newStep,
		},
	}
	for _, opt := range opts {
		opt(&def)
	}
	return def
}

func withRequirement(req gateRequirement) func(*gateDefinition) {
	return func(def *gateDefinition) {
		def.requirement = req
	}
}

func withDisplay(display gateDisplayPolicy) func(*gateDefinition) {
	return func(def *gateDefinition) {
		def.display = display
	}
}

func (m *Model) definitionForGate(gate Gate) gateDefinition {
	if def, ok := m.definitions[gate]; ok {
		return def
	}
	panic("unsupported onboarding gate: " + gate.String())
}

func defaultGateDefinitions() map[Gate]gateDefinition {
	return map[Gate]gateDefinition{
		GatePreflight: gateDef(GatePreflight, func(m *Model) Step {
			return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
		}),
		GateAuthenticate: gateDef(GateAuthenticate, func(m *Model) Step {
			return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
		}),
		GateRoleSelect: gateDef(GateRoleSelect, func(m *Model) Step {
			return role.New(m.theme, m.userPrefs, m.scope)
		}),
		GateOrgSelect: gateDef(GateOrgSelect, func(m *Model) Step {
			return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
		}),
		GateOrgCreate: gateDef(GateOrgCreate, func(m *Model) Step {
			return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
		}),
		GateAccountSelect: gateDef(GateAccountSelect, func(m *Model) Step {
			return accounts.NewSelect(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true})),
		GateAccountCreate: gateDef(GateAccountCreate, func(m *Model) Step {
			return accounts.NewCreate(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true})),
		GateRuntimeInit: gateDef(GateRuntimeInit, func(m *Model) Step {
			return runtimeinit.New(m.theme, *m.state.org, *m.state.account, m.scope)
		},
			withRequirement(gateRequirement{needsOrg: true, needsAccount: true}),
			withDisplay(gateDisplayPolicy{hidden: true, status: "Initializing account runtime..."}),
		),
		GateDatadogCheck: gateDef(GateDatadogCheck, func(m *Model) Step {
			return datadog.NewCheck(m.ctx, m.theme, *m.state.account, m.services, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true})),
		GateDatadogRegion: gateDef(GateDatadogRegion, func(m *Model) Step {
			return datadog.NewRegion(m.theme, m.scope)
		}),
		GateDatadogAPIKey: gateDef(GateDatadogAPIKey, func(m *Model) Step {
			return datadog.NewAPIKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.services, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true, needsDDSite: true})),
		GateDatadogAppKey: gateDef(GateDatadogAppKey, func(m *Model) Step {
			return datadog.NewAppKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.state.ddAPIKey, m.services, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true, needsDDSite: true, needsDDAPIKey: true})),
		GateDatadogDiscovery: gateDef(GateDatadogDiscovery, func(m *Model) Step {
			return datadog.NewDiscovery(m.ctx, m.theme, m.state.ddAccount, m.services, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true, needsDDAccount: true})),
		GateWorkspaceSelect: gateDef(GateWorkspaceSelect, func(m *Model) Step {
			return workspaces.NewSelect(m.ctx, m.theme, *m.state.account, m.services, m.orgPrefs, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true})),
		GateSync: gateDef(GateSync, func(m *Model) Step {
			return sync.New(m.theme, m.syncer, m.scope)
		}, withRequirement(gateRequirement{needsOrg: true, needsAccount: true, needsWorkspace: true})),
	}
}
