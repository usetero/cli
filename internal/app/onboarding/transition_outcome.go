package onboarding

import tea "charm.land/bubbletea/v2"

// TransitionKind describes the result of handling a step message.
type TransitionKind string

const (
	TransitionNoop    TransitionKind = "noop"
	TransitionAdvance TransitionKind = "advance"
)

// TransitionOutcome is the typed contract between transition handlers and the engine.
type TransitionOutcome struct {
	Kind TransitionKind
	Next Gate
	Cmd  tea.Cmd
}

func advance(next Gate) TransitionOutcome {
	return TransitionOutcome{Kind: TransitionAdvance, Next: next}
}

func advanceWith(next Gate, cmd tea.Cmd) TransitionOutcome {
	return TransitionOutcome{Kind: TransitionAdvance, Next: next, Cmd: cmd}
}

func noop() TransitionOutcome {
	return TransitionOutcome{Kind: TransitionNoop}
}

func (m *Model) applyTransitionOutcome(out TransitionOutcome) tea.Cmd {
	switch out.Kind {
	case TransitionAdvance:
		if out.Next != "" {
			nextCmd := m.goToGate(out.Next)
			if out.Cmd != nil {
				return tea.Batch(nextCmd, out.Cmd)
			}
			return nextCmd
		}
		return out.Cmd
	case TransitionNoop:
		return nil
	default:
		return out.Cmd
	}
}
