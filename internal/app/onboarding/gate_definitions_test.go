package onboarding

import "testing"

func TestDefaultGateDefinitionsCoverageAndContracts(t *testing.T) {
	t.Parallel()

	defs := defaultGateDefinitions()
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

	if len(defs) != len(expected) {
		t.Fatalf("definition count = %d, want %d", len(defs), len(expected))
	}

	for _, gate := range expected {
		def, ok := defs[gate]
		if !ok {
			t.Fatalf("missing definition for gate %s", gate)
		}
		if def.newStep == nil {
			t.Fatalf("gate %s has nil step factory", gate)
		}
	}
}

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
