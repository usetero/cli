package bootstrap

import "testing"

func TestRequirementForGate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		gate Gate
		want GateRequirement
	}{
		{gate: GateOrgSelect, want: GateRequirement{}},
		{gate: GateAccountSelect, want: GateRequirement{NeedsOrg: true}},
		{gate: GateRuntimeInit, want: GateRequirement{NeedsOrg: true, NeedsAccount: true}},
		{gate: GateDatadogAPIKey, want: GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDSite: true}},
		{gate: GateDatadogAppKey, want: GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDSite: true, NeedsDDAPIKey: true}},
		{gate: GateDatadogDiscovery, want: GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDAccount: true}},
		{gate: GateSync, want: GateRequirement{NeedsOrg: true, NeedsAccount: true}},
	}

	for _, tc := range cases {
		t.Run(tc.gate.String(), func(t *testing.T) {
			t.Parallel()
			got := RequirementForGate(tc.gate)
			if got != tc.want {
				t.Fatalf("RequirementForGate(%s) = %#v, want %#v", tc.gate, got, tc.want)
			}
		})
	}
}
