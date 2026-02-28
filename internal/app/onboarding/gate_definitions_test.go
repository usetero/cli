package onboarding

import (
	"testing"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func TestNewStepForGateCoverage(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")
	m.state.DDSite = "US1"
	m.state.DDAPIKey = "api-key"
	m.state.DDAccount = "dd-1"
	m.state.Workspace = ptrWorkspace("ws-1")

	expected := []Gate{
		bootstrap.GatePreflight,
		bootstrap.GateAuthenticate,
		bootstrap.GateRoleSelect,
		bootstrap.GateOrgSelect,
		bootstrap.GateOrgCreate,
		bootstrap.GateAccountSelect,
		bootstrap.GateAccountCreate,
		bootstrap.GateRuntimeInit,
		bootstrap.GateDatadogCheck,
		bootstrap.GateDatadogRegion,
		bootstrap.GateDatadogAPIKey,
		bootstrap.GateDatadogAppKey,
		bootstrap.GateDatadogDiscovery,
		bootstrap.GateWorkspaceSelect,
		bootstrap.GateSync,
	}

	for _, gate := range expected {
		step, err := m.newStepForGate(gate)
		if err != nil {
			t.Fatalf("gate %s is unexpectedly unsupported: %v", gate, err)
		}
		if step == nil {
			t.Fatalf("gate %s produced nil step", gate)
		}
	}
}

func TestNewStepForGateUnsupportedReturnsError(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	step, err := m.newStepForGate(Gate("unsupported_gate"))
	if err == nil {
		t.Fatal("expected unsupported gate to return error")
	}
	if step != nil {
		t.Fatalf("expected nil step for unsupported gate, got %T", step)
	}
}
