package onboarding

import "testing"

func TestDisplayPolicyForGate(t *testing.T) {
	t.Parallel()

	runtimePolicy := displayPolicyForGate(GateRuntimeInit)
	if !runtimePolicy.hidden {
		t.Fatalf("runtime init gate should be hidden by default")
	}
	if runtimePolicy.status == "" {
		t.Fatalf("runtime init gate should provide default status text")
	}

	rolePolicy := displayPolicyForGate(GateRoleSelect)
	if rolePolicy.hidden {
		t.Fatalf("role select should be visible by default")
	}
}
