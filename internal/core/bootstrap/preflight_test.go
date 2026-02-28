package bootstrap

import "testing"

func TestDecideNextGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   PreflightInput
		want Gate
	}{
		{
			name: "failed outcome routes to authenticate",
			in: PreflightInput{
				Outcome:      PreflightOutcomeFailed,
				HasValidAuth: true,
				Role:         RolePlatform,
				HasOrg:       true,
				HasAccount:   true,
			},
			want: GateAuthenticate,
		},
		{
			name: "invalid auth routes to authenticate",
			in: PreflightInput{
				Outcome:      PreflightOutcomeResolved,
				HasValidAuth: false,
			},
			want: GateAuthenticate,
		},
		{
			name: "missing role routes to role select",
			in: PreflightInput{
				Outcome:      PreflightOutcomeResolved,
				HasValidAuth: true,
			},
			want: GateRoleSelect,
		},
		{
			name: "missing org routes to org select",
			in: PreflightInput{
				Outcome:      PreflightOutcomeResolved,
				HasValidAuth: true,
				Role:         RolePlatform,
			},
			want: GateOrgSelect,
		},
		{
			name: "missing account routes to account select",
			in: PreflightInput{
				Outcome:      PreflightOutcomeResolved,
				HasValidAuth: true,
				Role:         RoleEngineer,
				HasOrg:       true,
			},
			want: GateAccountSelect,
		},
		{
			name: "all resolved routes to runtime init",
			in: PreflightInput{
				Outcome:      PreflightOutcomeResolved,
				HasValidAuth: true,
				Role:         RoleEngineer,
				HasOrg:       true,
				HasAccount:   true,
			},
			want: GateRuntimeInit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DecideNextGate(tc.in)
			if got != tc.want {
				t.Fatalf("DecideNextGate() = %q, want %q", got, tc.want)
			}
		})
	}
}
