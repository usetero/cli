package onboarding

import "time"

func (m *Model) enterGate(gate Gate, trigger string) {
	now := time.Now()
	if trigger == "" {
		trigger = "unspecified"
	}

	if m.gate != "" && !m.gateEnteredAt.IsZero() {
		m.scope.Info("onboarding gate exit",
			"gate", m.gate.String(),
			"next_gate", gate.String(),
			"trigger", trigger,
			"duration_ms", now.Sub(m.gateEnteredAt).Milliseconds(),
		)
	}

	m.gate = gate
	m.gateEnteredAt = now
	m.scope.Info("onboarding gate enter",
		"gate", gate.String(),
		"trigger", trigger,
	)
}

func (m *Model) completeCurrentGate(trigger string) {
	if m.gate == "" || m.gateEnteredAt.IsZero() {
		return
	}
	if trigger == "" {
		trigger = "unspecified"
	}

	m.scope.Info("onboarding gate exit",
		"gate", m.gate.String(),
		"next_gate", "complete",
		"trigger", trigger,
		"duration_ms", time.Since(m.gateEnteredAt).Milliseconds(),
	)
	m.gateEnteredAt = time.Time{}
}
