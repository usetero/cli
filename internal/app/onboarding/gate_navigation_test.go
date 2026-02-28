package onboarding

import (
	"testing"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func TestRunGateUnsupportedDoesNotSetStep(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	cmd := m.runGate(Gate("unsupported_gate"), "test")
	if cmd == nil {
		t.Fatal("expected recovery command for unsupported gate")
	}
	if m.step != nil {
		if m.gate != bootstrap.GatePreflight {
			t.Fatalf("gate = %s, want %s", m.gate, bootstrap.GatePreflight)
		}
	} else {
		t.Fatal("expected recovery step to be set")
	}
}

func TestRunGateRecordsGateEntry(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	cmd := m.runGate(bootstrap.GateRoleSelect, "test")
	if cmd == nil {
		t.Fatal("expected command for supported gate")
	}
	if m.gate != bootstrap.GateRoleSelect {
		t.Fatalf("gate = %s, want %s", m.gate, bootstrap.GateRoleSelect)
	}
	if m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be set")
	}
}

func TestCompleteCurrentGateClearsTimingState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	_ = m.runGate(bootstrap.GateRoleSelect, "test")
	if m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be set")
	}

	m.completeCurrentGate("sync_complete")
	if !m.gateEnteredAt.IsZero() {
		t.Fatal("expected gateEnteredAt to be reset")
	}
}
