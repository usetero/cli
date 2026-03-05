package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// setStep sets the current step and initializes it.
func (m *Model) setStep(step Step) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

func (m *Model) goToGate(gate Gate, trigger string) tea.Cmd {
	requested := gate
	if rewind := m.rewindGateFor(gate); rewind != gate {
		m.scope.Warn("rewinding onboarding gate due to unmet requirements", "requested_gate", requested.String(), "rewind_gate", rewind.String())
		gate = rewind
	}
	m.scope.Debug("onboarding gate transition", slog.String("gate", gate.String()))
	return m.runGate(gate, trigger)
}

func (m *Model) runGate(gate Gate, trigger string) tea.Cmd {
	step, err := m.newStepForGate(gate)
	if err != nil {
		m.scope.Error("failed to build onboarding gate",
			slog.String("gate", gate.String()),
			slog.String("trigger", trigger),
			slog.String("error", err.Error()),
			slog.Bool("recovery", gate != bootstrap.GatePreflight),
		)
		if gate == bootstrap.GatePreflight {
			return appevents.PublishErrorToastCmd("Onboarding setup failed. Please restart.", err, true)
		}
		m.scope.Warn("recovering onboarding to preflight",
			slog.String("failed_gate", gate.String()),
			slog.String("trigger", trigger),
			slog.Bool("recovery", true),
		)
		recovery := m.runGate(bootstrap.GatePreflight, "gate_recovery")
		return tea.Batch(
			appevents.PublishErrorToastCmd("Onboarding state changed. Rechecking setup.", err, false),
			recovery,
		)
	}
	m.enterGate(gate, trigger)
	return m.setStep(step)
}
