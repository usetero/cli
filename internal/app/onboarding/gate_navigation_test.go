package onboarding

import "testing"

func TestRunGateUnsupportedDoesNotSetStep(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	cmd := m.runGate(Gate("unsupported_gate"), "test")
	if cmd != nil {
		t.Fatal("expected nil command for unsupported gate")
	}
	if m.step != nil {
		t.Fatalf("expected step to remain nil, got %T", m.step)
	}
}

func TestRunGateRecordsGateEntry(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	cmd := m.runGate(GateRoleSelect, "test")
	if cmd == nil {
		t.Fatal("expected command for supported gate")
	}
	if m.gate != GateRoleSelect {
		t.Fatalf("gate = %s, want %s", m.gate, GateRoleSelect)
	}
	if m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be set")
	}
}

func TestCompleteCurrentGateClearsTimingState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	_ = m.runGate(GateRoleSelect, "test")
	if m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be set")
	}

	m.completeCurrentGate("sync_complete")
	if !m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be reset")
	}
}
