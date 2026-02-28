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
		step := m.newStepForGate(gate)
		if step == nil {
			t.Fatalf("gate %s produced nil step", gate)
		}
	}
}

func TestNewStepForGateUnsupportedPanics(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unsupported gate")
		}
	}()

	_ = m.newStepForGate(Gate("unsupported_gate"))
}
