package app

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

func TestOnboardingGateRecoverySurfacesToastAndReturnsToPreflight(t *testing.T) {
	t.Parallel()

	m := newViewTestModel(t)
	m.onboarding.SetTestingGateBuildHook(func(gate bootstrap.Gate) error {
		if gate == bootstrap.GateRoleSelect {
			return errors.New("forced gate build failure")
		}
		return nil
	})

	_, cmd := m.Update(bootstrap.PreflightResolved{
		State: bootstrap.PreflightState{
			Outcome:      bootstrap.PreflightOutcomeResolved,
			HasValidAuth: true,
			Role:         "",
		},
	})

	if m.onboarding.TestingCurrentGate() != bootstrap.GatePreflight {
		t.Fatalf("gate = %s, want %s", m.onboarding.TestingCurrentGate(), bootstrap.GatePreflight)
	}

	msgs := collectCmdMsgs(cmd)
	var toast appevents.Error
	var found bool
	for _, msg := range msgs {
		if e, ok := msg.(appevents.Error); ok {
			toast = e
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected onboarding recovery to emit error toast message")
	}
	if !strings.Contains(toast.Message, "Onboarding state changed. Rechecking setup.") {
		t.Fatalf("unexpected toast message: %q", toast.Message)
	}

	_, _ = m.Update(toast)
	if !strings.Contains(m.View().Content, "Onboarding state changed. Rechecking setup.") {
		t.Fatalf("expected toast to be rendered in app view, got: %q", m.View().Content)
	}
}

func collectCmdMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	if msg == nil {
		return nil
	}

	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, collectCmdMsgs(c)...)
		}
		return out
	}

	return []tea.Msg{msg}
}
