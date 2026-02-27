package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

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

// setStep sets the current step and initializes it.
func (m *Model) setStep(step Step) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

func (m *Model) goToGate(gate Gate) tea.Cmd {
	m.gate = gate
	m.scope.Debug("onboarding gate transition", slog.String("gate", gate.String()))
	return m.setStep(m.stepForGate(gate))
}

func (m *Model) stepForGate(gate Gate) Step {
	if rewind := rewindGateFor(gate, m.state); rewind != gate {
		m.scope.Warn("rewinding onboarding gate due to unmet requirements", "requested_gate", gate.String(), "rewind_gate", rewind.String())
		gate = rewind
	}

	switch gate {
	case GatePreflight:
		return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
	case GateAuthenticate:
		return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
	case GateRoleSelect:
		return role.New(m.theme, m.userPrefs, m.scope)
	case GateOrgSelect:
		return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
	case GateOrgCreate:
		return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
	case GateAccountSelect:
		return accounts.NewSelect(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
	case GateAccountCreate:
		return accounts.NewCreate(m.ctx, m.theme, *m.state.org, m.services, m.orgPrefs, m.scope)
	case GateRuntimeInit:
		return runtimeinit.New(m.theme, *m.state.org, *m.state.account, m.scope)
	case GateDatadogCheck:
		return datadog.NewCheck(m.ctx, m.theme, *m.state.account, m.services, m.scope)
	case GateDatadogRegion:
		return datadog.NewRegion(m.theme, m.scope)
	case GateDatadogAPIKey:
		return datadog.NewAPIKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.services, m.scope)
	case GateDatadogAppKey:
		return datadog.NewAppKey(m.ctx, m.theme, *m.state.account, m.state.ddSite, m.state.ddAPIKey, m.services, m.scope)
	case GateDatadogDiscovery:
		return datadog.NewDiscovery(m.ctx, m.theme, m.state.ddAccount, m.services, m.scope)
	case GateWorkspaceSelect:
		return workspaces.NewSelect(m.ctx, m.theme, *m.state.account, m.services, m.orgPrefs, m.scope)
	case GateSync:
		return sync.New(m.theme, m.syncer, m.scope)
	default:
		panic("unsupported onboarding gate: " + gate.String())
	}
}
