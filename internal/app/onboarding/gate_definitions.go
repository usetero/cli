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
	newStep func(m *Model) Step
}

var gateDefinitions = defaultGateDefinitions()

func gateDef(newStep func(m *Model) Step, opts ...func(*gateDefinition)) gateDefinition {
	def := gateDefinition{
		newStep: newStep,
	}
	for _, opt := range opts {
		opt(&def)
	}
	return def
}

func (m *Model) definitionForGate(gate Gate) gateDefinition {
	if def, ok := gateDefinitions[gate]; ok {
		return def
	}
	panic("unsupported onboarding gate: " + gate.String())
}

func defaultGateDefinitions() map[Gate]gateDefinition {
	return map[Gate]gateDefinition{
		GatePreflight: gateDef(func(m *Model) Step {
			return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
		}),
		GateAuthenticate: gateDef(func(m *Model) Step {
			return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
		}),
		GateRoleSelect: gateDef(func(m *Model) Step {
			return role.New(m.theme, m.userPrefs, m.scope)
		}),
		GateOrgSelect: gateDef(func(m *Model) Step {
			return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
		}),
		GateOrgCreate: gateDef(func(m *Model) Step {
			return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
		}),
		GateAccountSelect: gateDef(func(m *Model) Step {
			return accounts.NewSelect(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
		}),
		GateAccountCreate: gateDef(func(m *Model) Step {
			return accounts.NewCreate(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
		}),
		GateRuntimeInit: gateDef(func(m *Model) Step {
			return runtimeinit.New(m.theme, *m.state.Org, *m.state.Account, m.scope)
		}),
		GateDatadogCheck: gateDef(func(m *Model) Step {
			return datadog.NewCheck(m.ctx, m.theme, *m.state.Account, m.services, m.scope)
		}),
		GateDatadogRegion: gateDef(func(m *Model) Step {
			return datadog.NewRegion(m.theme, m.scope)
		}),
		GateDatadogAPIKey: gateDef(func(m *Model) Step {
			return datadog.NewAPIKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.services, m.scope)
		}),
		GateDatadogAppKey: gateDef(func(m *Model) Step {
			return datadog.NewAppKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.state.DDAPIKey, m.services, m.scope)
		}),
		GateDatadogDiscovery: gateDef(func(m *Model) Step {
			return datadog.NewDiscovery(m.ctx, m.theme, m.state.DDAccount, m.services, m.scope)
		}),
		GateWorkspaceSelect: gateDef(func(m *Model) Step {
			return workspaces.NewSelect(m.ctx, m.theme, *m.state.Account, m.services, m.orgPrefs, m.scope)
		}),
		GateSync: gateDef(func(m *Model) Step {
			return sync.New(m.theme, m.syncer, m.scope)
		}),
	}
}
