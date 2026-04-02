package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestRewindGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target Gate
		req    GateRequirement
		state  State
		want   Gate
	}{
		{
			name:   "no requirements unchanged",
			target: GateOrgSelect,
			req:    GateRequirement{},
			state:  State{},
			want:   GateOrgSelect,
		},
		{
			name:   "account requirement rewinds to org when org missing",
			target: GateAccountSelect,
			req:    GateRequirement{NeedsOrg: true, NeedsAccount: true},
			state:  State{},
			want:   GateOrgSelect,
		},
		{
			name:   "datadog app key requirement rewinds to api key when missing",
			target: GateDatadogAppKey,
			req:    GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDSite: true, NeedsDDAPIKey: true},
			state:  State{Org: &domain.Organization{ID: "org-1"}, Account: &domain.Account{ID: "acc-1"}, DDSite: "US1"},
			want:   GateDatadogAPIKey,
		},
		{
			name:   "sync requirement stays on sync once org and account are present",
			target: GateSync,
			req:    GateRequirement{NeedsOrg: true, NeedsAccount: true},
			state:  State{Org: &domain.Organization{ID: "org-1"}, Account: &domain.Account{ID: "acc-1"}},
			want:   GateSync,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := RewindGate(tc.target, tc.req, tc.state)
			if got != tc.want {
				t.Fatalf("RewindGate() = %q, want %q", got, tc.want)
			}
		})
	}
}
