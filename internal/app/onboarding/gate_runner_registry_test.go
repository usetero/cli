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
		if def.runner == nil {
			t.Fatalf("gate %s has nil runner", gate)
		}
		if def.runner.Gate() != gate {
			t.Fatalf("gate %s runner mismatch: got %s", gate, def.runner.Gate())
		}
	}
}

func TestDefaultGateDefinitionsPolicies(t *testing.T) {
	t.Parallel()

	defs := defaultGateDefinitions()

	if !defs[GateRuntimeInit].display.hidden {
		t.Fatalf("runtime init should be hidden by default")
	}
	if defs[GateRuntimeInit].display.status == "" {
		t.Fatalf("runtime init hidden policy should include status text")
	}

	if !defs[GateAccountSelect].requirement.NeedsOrg {
		t.Fatalf("account select should require org")
	}
	if !defs[GateDatadogAPIKey].requirement.NeedsDDSite {
		t.Fatalf("datadog api key should require selected site")
	}
	if !defs[GateSync].requirement.NeedsWorkspace {
		t.Fatalf("sync should require workspace")
	}
}
