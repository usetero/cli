package onboarding

import "testing"

func TestRewindGateFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target Gate
		state  onboardingState
		want   Gate
	}{
		{name: "no requirements gate unchanged", target: GateOrgSelect, state: onboardingState{}, want: GateOrgSelect},
		{name: "account gate rewinds to org", target: GateAccountSelect, state: onboardingState{}, want: GateOrgSelect},
		{name: "datadog check rewinds to account", target: GateDatadogCheck, state: onboardingState{org: ptrOrg("org-1")}, want: GateAccountSelect},
		{name: "datadog api rewinds to region when site missing", target: GateDatadogAPIKey, state: onboardingState{org: ptrOrg("org-1"), account: ptrAccount("acc-1")}, want: GateDatadogRegion},
		{name: "datadog app rewinds to api key when api key missing", target: GateDatadogAppKey, state: onboardingState{org: ptrOrg("org-1"), account: ptrAccount("acc-1"), ddSite: "US1"}, want: GateDatadogAPIKey},
		{name: "discovery rewinds to datadog check without dd account", target: GateDatadogDiscovery, state: onboardingState{org: ptrOrg("org-1"), account: ptrAccount("acc-1")}, want: GateDatadogCheck},
		{name: "sync rewinds to workspace", target: GateSync, state: onboardingState{org: ptrOrg("org-1"), account: ptrAccount("acc-1")}, want: GateWorkspaceSelect},
		{name: "sync stays when requirements met", target: GateSync, state: onboardingState{org: ptrOrg("org-1"), account: ptrAccount("acc-1"), workspace: ptrWorkspace("ws-1")}, want: GateSync},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			m.state = tc.state
			got := m.rewindGateFor(tc.target)
			if got != tc.want {
				t.Fatalf("rewindGateFor(%s) = %s, want %s", tc.target, got, tc.want)
			}
		})
	}
}
