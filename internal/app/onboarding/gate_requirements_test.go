package onboarding

import (
	"testing"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func TestRewindGateFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target Gate
		state  bootstrap.State
		want   Gate
	}{
		{name: "no requirements gate unchanged", target: bootstrap.GateOrgSelect, state: bootstrap.State{}, want: bootstrap.GateOrgSelect},
		{name: "account gate rewinds to org", target: bootstrap.GateAccountSelect, state: bootstrap.State{}, want: bootstrap.GateOrgSelect},
		{name: "datadog check rewinds to account", target: bootstrap.GateDatadogCheck, state: bootstrap.State{Org: ptrOrg("org-1")}, want: bootstrap.GateAccountSelect},
		{name: "datadog api rewinds to region when site missing", target: bootstrap.GateDatadogAPIKey, state: bootstrap.State{Org: ptrOrg("org-1"), Account: ptrAccount("acc-1")}, want: bootstrap.GateDatadogRegion},
		{name: "datadog app rewinds to api key when api key missing", target: bootstrap.GateDatadogAppKey, state: bootstrap.State{Org: ptrOrg("org-1"), Account: ptrAccount("acc-1"), DDSite: "US1"}, want: bootstrap.GateDatadogAPIKey},
		{name: "discovery rewinds to datadog check without dd account", target: bootstrap.GateDatadogDiscovery, state: bootstrap.State{Org: ptrOrg("org-1"), Account: ptrAccount("acc-1")}, want: bootstrap.GateDatadogCheck},
		{name: "sync rewinds to workspace", target: bootstrap.GateSync, state: bootstrap.State{Org: ptrOrg("org-1"), Account: ptrAccount("acc-1")}, want: bootstrap.GateWorkspaceSelect},
		{name: "sync stays when requirements met", target: bootstrap.GateSync, state: bootstrap.State{Org: ptrOrg("org-1"), Account: ptrAccount("acc-1"), Workspace: ptrWorkspace("ws-1")}, want: bootstrap.GateSync},
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

func TestNewStepForGateValidatesRequiredState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")
	_, err := m.newStepForGate(bootstrap.GateDatadogAPIKey)
	if err == nil {
		t.Fatal("expected missing datadog site error")
	}
}
