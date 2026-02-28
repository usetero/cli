package onboarding

import "testing"

func TestRunGateUnsupportedDoesNotSetStep(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	cmd := m.runGate(Gate("unsupported_gate"))
	if cmd != nil {
		t.Fatal("expected nil command for unsupported gate")
	}
	if m.step != nil {
		t.Fatalf("expected step to remain nil, got %T", m.step)
	}
}
