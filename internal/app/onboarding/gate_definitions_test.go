package onboarding

import "testing"

func TestNewStepForGateCoverage(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")
	m.state.DDSite = "US1"
	m.state.DDAPIKey = "api-key"
	m.state.DDAccount = "dd-1"

	expected := []Gate{
		GatePreflight,
		GateAuthenticate,
		GateRoleSelect,
		GateOrgSelect,
		GateOrgCreate,
		GateAccountSelect,
		GateAccountCreate,
		GateRuntimeInit,
		GateDatadogCheck,
		GateDatadogRegion,
		GateDatadogAPIKey,
		GateDatadogAppKey,
		GateDatadogDiscovery,
		GateWorkspaceSelect,
		GateSync,
	}

	for _, gate := range expected {
		step, ok := m.newStepForGate(gate)
		if !ok {
			t.Fatalf("gate %s is unexpectedly unsupported", gate)
		}
		if step == nil {
			t.Fatalf("gate %s produced nil step", gate)
		}
	}
}

func TestNewStepForGateUnsupportedReturnsNotOK(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	step, ok := m.newStepForGate(Gate("unsupported_gate"))
	if ok {
		t.Fatal("expected unsupported gate to return ok=false")
	}
	if step != nil {
		t.Fatalf("expected nil step for unsupported gate, got %T", step)
	}
}
